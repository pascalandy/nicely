package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/platform"
	"github.com/rogpeppe/go-internal/testscript"
)

// cmdExits runs a command and asserts its exact exit code, which the built-in
// "! exec" cannot do: exits <code> <command> [args...]
func cmdExits(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("exits takes no negation; give the expected code instead")
	}
	if len(args) < 2 {
		ts.Fatalf("usage: exits <code> <command> [args...]")
	}
	want, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("exits: invalid code %q", args[0])
	}
	got := 0
	var exitErr *exec.ExitError
	switch err := ts.Exec(args[1], args[2:]...); {
	case err == nil:
	case errors.As(err, &exitErr):
		got = exitErr.ExitCode()
	default:
		ts.Fatalf("exits: %v", err)
	}
	if got != want {
		ts.Fatalf("%s exited with %d, want %d", args[1], got, want)
	}
}

// cmdAnswer decodes a whole stream as one answer, the way an agent reads it,
// so a matching substring cannot hide extra text, a wrong type, or a second
// object: answer <stdout|stderr> <ok|CODE>, where CODE leads the errors.
func cmdAnswer(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 2 || (args[0] != "stdout" && args[0] != "stderr") {
		ts.Fatalf("usage: answer <stdout|stderr> <ok|CODE>")
	}
	if err := checkAnswer(ts.ReadFile(args[0]), args[1]); err != nil {
		ts.Fatalf("answer on %s: %v", args[0], err)
	}
}

func checkAnswer(text, want string) error {
	line, ok := strings.CutSuffix(text, "\n")
	if !ok || strings.Contains(line, "\n") {
		return fmt.Errorf("want exactly one line, got %q", text)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &keys); err != nil {
		return fmt.Errorf("not one JSON object: %w", err)
	}
	if v := string(keys["contract_version"]); v != "1" {
		return fmt.Errorf("contract_version is %q, want the integer 1", v)
	}
	wantOK := want == "ok"
	if v := string(keys["ok"]); v != strconv.FormatBool(wantOK) {
		return fmt.Errorf("ok is %q, want %t", v, wantOK)
	}
	if raw, present := keys["warnings"]; present {
		if _, err := problems(raw); err != nil {
			return fmt.Errorf("warnings: %w", err)
		}
	}
	raw, present := keys["errors"]
	if wantOK {
		if present {
			return errors.New("a success holds errors")
		}
		return nil
	}
	errs, err := problems(raw)
	if err != nil {
		return fmt.Errorf("errors: %w", err)
	}
	if lead := errs[0]["code"]; lead != want {
		return fmt.Errorf("leading code is %s, want %s", lead, want)
	}
	return nil
}

// problems decodes a non-empty array whose objects hold exactly a non-empty
// code, message, and hint.
func problems(raw json.RawMessage) ([]map[string]string, error) {
	var list []map[string]string
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("empty or null")
	}
	for _, p := range list {
		if len(p) != 3 || p["code"] == "" || p["message"] == "" || p["hint"] == "" {
			return nil, fmt.Errorf("want non-empty code, message, and hint, got %v", p)
		}
	}
	return list, nil
}

var snapshots sync.Map

// cmdSnapshot records every file and link under $HOME, except Nicely's cache
// and the temporary folder: snapshot <name>. A dry run may fill only the cache.
func cmdSnapshot(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 {
		ts.Fatalf("usage: snapshot <name>")
	}
	snapshots.Store(ts.Getenv("WORK")+"\x00"+args[0], fingerprint(ts))
}

// cmdUnchanged compares $HOME with a snapshot: unchanged <name>. With "!",
// it asserts that something outside Nicely's cache changed.
func cmdUnchanged(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: unchanged <name>")
	}
	before, found := snapshots.Load(ts.Getenv("WORK") + "\x00" + args[0])
	if !found {
		ts.Fatalf("no snapshot %q", args[0])
	}
	diff := compare(before.(map[string]string), fingerprint(ts))
	switch {
	case neg && len(diff) == 0:
		ts.Fatalf("nothing changed outside the cache")
	case !neg && len(diff) > 0:
		ts.Fatalf("changed outside the cache: %s", strings.Join(diff, ", "))
	}
}

func fingerprint(ts *testscript.TestScript) map[string]string {
	home := ts.Getenv("HOME")
	skip := []string{
		filepath.Join(ts.Getenv("XDG_CACHE_HOME"), "nicely"),
		ts.Getenv("TMPDIR"),
	}
	files := map[string]string{}
	err := filepath.WalkDir(home, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if slices.Contains(skip, path) {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(home, path)
		switch {
		case d.IsDir():
			return nil
		case d.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			files[rel] = "link to " + target
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	})
	ts.Check(err)
	return files
}

func compare(before, after map[string]string) []string {
	var diff []string
	for path, sum := range after {
		if before[path] != sum {
			diff = append(diff, path)
		}
	}
	for path := range before {
		if _, kept := after[path]; !kept {
			diff = append(diff, path+" (removed)")
		}
	}
	slices.Sort(diff)
	return diff
}

// closedPipe is the closedpipe program of the scenarios: it runs its
// arguments as a command whose stdout is a pipe with no reader left, and
// exits with that command's code: closedpipe ncly --version. A shell
// redirection cannot close the reader before the first write. As a program,
// it runs with the scenario's environment, and its stderr is the command's.
func closedPipe() {
	reader, writer, err := os.Pipe()
	if err == nil {
		err = reader.Close()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "closedpipe:", err)
		os.Exit(127)
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout, cmd.Stderr = writer, os.Stderr
	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil && !errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, "closedpipe:", err)
		os.Exit(127)
	}
	os.Exit(shellStatus(cmd.ProcessState))
}

// shellStatus is the exit code of a finished command as a shell reports
// it: 128 plus the number of the signal that ended it, if any.
func shellStatus(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}

var (
	marked  = regexp.MustCompile(`(?s)⟦.*?⟧`)
	example = regexp.MustCompile(`(?m)^\s*\$ .*$`)
	letter  = regexp.MustCompile(`\p{L}`)
	version = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)*$`)
)

// cmdPseudo fails on any word of a stream that the pseudo-locale did not
// mark: pseudo <stdout|stderr> [word...]. Example lines, flags, ncly, codes
// such as USAGE_INVALID, placeholders such as <tag>, version numbers, and
// the listed words, such as command names, stay unmarked because they are
// never translated.
func cmdPseudo(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 1 || (args[0] != "stdout" && args[0] != "stderr") {
		ts.Fatalf("usage: pseudo <stdout|stderr> [word...]")
	}
	text := ts.ReadFile(args[0])
	if !marked.MatchString(text) {
		ts.Fatalf("pseudo: no marked text on %s", args[0])
	}
	var stray []string
	for _, word := range strings.Fields(marked.ReplaceAllString(example.ReplaceAllString(text, ""), " ")) {
		word = strings.Trim(word, `.,:;!?()"'`)
		switch {
		case !letter.MatchString(word),
			strings.HasPrefix(word, "-"),
			word == "ncly",
			isCode(word),
			strings.HasPrefix(word, "<") || strings.HasPrefix(word, "["),
			version.MatchString(word),
			slices.Contains(args[1:], word):
			continue
		}
		stray = append(stray, word)
	}
	if len(stray) > 0 {
		ts.Fatalf("pseudo: text outside the catalog on %s: %s", args[0], strings.Join(stray, " "))
	}
}

// isCode reports whether a word is a registered error or warning code, the
// only uppercase words that stay untranslated.
func isCode(word string) bool {
	_, isError := contract.Code(word).Exit()
	return isError || contract.Code(word).IsWarning()
}

// cmdWaitfile waits until a file exists, such as the process ID that a
// fixture writes once it is ready for a signal: waitfile <path>
func cmdWaitfile(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 {
		ts.Fatalf("usage: waitfile <path>")
	}
	if !poll(func() bool { _, err := os.Stat(ts.MkAbs(args[0])); return err == nil }) {
		ts.Fatalf("waitfile: %s never appeared", args[0])
	}
}

// cmdGone checks that the process whose ID a file holds no longer runs,
// waiting a moment for a signal to take effect: gone <path>
func cmdGone(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 {
		ts.Fatalf("usage: gone <path>")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(ts.ReadFile(args[0])))
	if err != nil {
		ts.Fatalf("gone: %s holds no process ID", args[0])
	}
	if !poll(func() bool { return !platform.Running(pid) }) {
		ts.Fatalf("gone: process %d still runs", pid)
	}
}

// cmdTerm sends SIGTERM to the background commands, which the built-in
// kill cannot send: term
func cmdTerm(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 0 {
		ts.Fatalf("usage: term")
	}
	for _, cmd := range ts.BackgroundCmds() {
		ts.Check(cmd.Process.Signal(syscall.SIGTERM))
	}
}

// poll reports whether done holds within 10 seconds.
func poll(done func() bool) bool {
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if done() {
			return true
		}
	}
	return done()
}

// exitCode is the exitcode program of the scenarios: it runs a command,
// passes on the Ctrl-C and SIGTERM that it receives, and writes the
// command's exit code to a file: exitcode [-group] <file> <command>
// [args...]. It gives the exact exit code of a background command, which
// the built-in wait only tells as a failure. With -group, the command runs
// in a process group of its own, which gets each signal as the foreground
// group of a terminal gets Ctrl-C: the command and every child that stays
// in its group, at once.
func exitCode() {
	args := os.Args[1:]
	group := args[0] == "-group"
	if group {
		args = args[1:]
	}
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: group}
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "exitcode:", err)
		os.Exit(127)
	}
	go func() {
		for sig := range signals {
			if group {
				_ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			} else {
				_ = cmd.Process.Signal(sig)
			}
		}
	}()
	_ = cmd.Wait()
	if err := os.WriteFile(args[0], []byte(strconv.Itoa(shellStatus(cmd.ProcessState))+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "exitcode:", err)
		os.Exit(127)
	}
}
