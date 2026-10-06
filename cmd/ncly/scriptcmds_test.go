package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

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

// cmdClosedPipe runs a command whose stdout is a pipe with no reader left,
// asserts its exit code, and saves its stderr to a file:
// closedpipe <code> <stderr-file> <command> [args...]. A shell redirection
// cannot close the reader before the first write, so the test does it.
func cmdClosedPipe(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 3 {
		ts.Fatalf("usage: closedpipe <code> <stderr-file> <command> [args...]")
	}
	want, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("closedpipe: invalid code %q", args[0])
	}
	reader, writer, err := os.Pipe()
	ts.Check(err)
	ts.Check(reader.Close())
	defer func() { _ = writer.Close() }()
	path, err := lookPath(ts, args[2])
	ts.Check(err)
	var stderr strings.Builder
	cmd := exec.Command(path, args[3:]...)
	cmd.Dir = ts.Getenv("WORK")
	cmd.Env = append(os.Environ(), "HOME="+ts.Getenv("HOME"), "PATH="+ts.Getenv("PATH"),
		"XDG_CONFIG_HOME="+ts.Getenv("XDG_CONFIG_HOME"), "XDG_CACHE_HOME="+ts.Getenv("XDG_CACHE_HOME"))
	cmd.Stdout, cmd.Stderr = writer, &stderr
	got := 0
	var exitErr *exec.ExitError
	switch err := cmd.Run(); {
	case err == nil:
	case errors.As(err, &exitErr):
		got = exitErr.ExitCode()
	default:
		ts.Fatalf("closedpipe: %v", err)
	}
	ts.Check(os.WriteFile(ts.MkAbs(args[1]), []byte(stderr.String()), 0o600))
	if got != want {
		ts.Fatalf("%s exited with %d, want %d; stderr: %q", args[2], got, want, stderr.String())
	}
}

// lookPath finds a command in the PATH of the script, which holds ncly.
func lookPath(ts *testscript.TestScript, name string) (string, error) {
	for _, dir := range filepath.SplitList(ts.Getenv("PATH")) {
		if path := filepath.Join(dir, name); fileExists(path) {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s not found in PATH", name)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
