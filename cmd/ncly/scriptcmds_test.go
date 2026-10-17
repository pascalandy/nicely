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

	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/platform"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
// code, message, and hint, and may add next, a non-empty array of non-empty
// words.
func problems(raw json.RawMessage) ([]map[string]string, error) {
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("empty or null")
	}
	var decoded []map[string]string
	for _, entry := range list {
		p := map[string]string{}
		for _, key := range []string{"code", "message", "hint"} {
			var value string
			if json.Unmarshal(entry[key], &value) != nil || value == "" {
				return nil, fmt.Errorf("want non-empty code, message, and hint, got %s", raw)
			}
			p[key] = value
		}
		size := 3
		if next, present := entry["next"]; present {
			var words []string
			if json.Unmarshal(next, &words) != nil || len(words) == 0 || slices.Contains(words, "") {
				return nil, fmt.Errorf("want next as a non-empty array of words, got %s", next)
			}
			size++
		}
		if len(entry) != size {
			return nil, fmt.Errorf("want only code, message, hint, and next, got %s", raw)
		}
		decoded = append(decoded, p)
	}
	return decoded, nil
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
	// marked matches an innermost marked text, since a translated text can
	// hold another, such as a greeting inside a sentence.
	marked  = regexp.MustCompile(`⟦[^⟦⟧]*⟧`)
	example = regexp.MustCompile(`(?m)^\s*\$ .*$`)
	letter  = regexp.MustCompile(`\p{L}`)
	version = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)*$`)
)

// cmdPseudo fails on any word of a stream that the pseudo-locale did not
// mark: pseudo <stdout|stderr> [word...]. Example lines, flags, ncly, codes
// such as USAGE_INVALID, placeholders such as <tag>, version numbers, and
// the listed words, such as command names, stay unmarked because they are
// never translated. Escape sequences, such as the styles and the cursor moves
// that pty shows, are no words.
func cmdPseudo(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 1 || (args[0] != "stdout" && args[0] != "stderr") {
		ts.Fatalf("usage: pseudo <stdout|stderr> [word...]")
	}
	text := ansi.Strip(ts.ReadFile(args[0]))
	if !marked.MatchString(text) {
		ts.Fatalf("pseudo: no marked text on %s", args[0])
	}
	text = example.ReplaceAllString(text, "")
	for marked.MatchString(text) {
		text = marked.ReplaceAllString(text, " ")
	}
	var stray []string
	for _, word := range strings.Fields(text) {
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

// codeForm is the form of an extension's code, such as HELLO_GRUMPY.
var codeForm = regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)+$`)

// isCode reports whether a word is a code, the only uppercase words that
// stay untranslated: one of core's, or one of the form of an extension's,
// whose set is open.
func isCode(word string) bool {
	_, isError := contract.Code(word).Exit()
	return isError || contract.Code(word).IsWarning() || codeForm.MatchString(word)
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

// pgroup is the pgroup program of the scenarios: it runs a command in a
// process group of its own, as job control starts a background job, without
// the terminal that dash 0.5.12 needs for set -m: pgroup <command>
// [args...]. A shell without job control starts a background command with
// Ctrl-C ignored, and Go keeps that, so pgroup handles Ctrl-C, which exec
// then resets to its default action.
func pgroup() {
	path, err := exec.LookPath(os.Args[1])
	if err == nil {
		signal.Notify(make(chan os.Signal, 1), os.Interrupt)
		err = syscall.Setpgid(0, 0)
	}
	if err == nil {
		err = syscall.Exec(path, os.Args[1:], os.Environ())
	}
	fmt.Fprintln(os.Stderr, "pgroup:", err)
	os.Exit(127)
}

// cmdAgree checks that a help page agrees with the answer of ncly describe
// for the same command, in the same language: agree <help-file> <describe-file>.
// The page starts with the summary, then the line of effects when the command
// has one, naming --dry-run when it supports it. Its usage line shows each
// argument as <name> when required and [name] otherwise, each required flag,
// and each flag group in parentheses when it needs a flag and in brackets
// otherwise. Each flag has a line with its summary, and each example appears.
func cmdAgree(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 2 {
		ts.Fatalf("usage: agree <help-file> <describe-file>")
	}
	if err := agree(ts.ReadFile(args[0]), ts.ReadFile(args[1])); err != nil {
		ts.Fatalf("agree: %v", err)
	}
}

type described struct {
	Path    string `json:"path"`
	Summary string `json:"summary"`
	Args    []struct {
		Name     string `json:"name"`
		Required bool   `json:"required"`
		Variadic bool   `json:"variadic"`
	} `json:"args"`
	Flags []struct {
		Name      string `json:"name"`
		Shorthand string `json:"shorthand"`
		Value     string `json:"value"`
		Required  bool   `json:"required"`
		Summary   string `json:"summary"`
	} `json:"flags"`
	FlagGroups []struct {
		Flags []string `json:"flags"`
		Min   int      `json:"min"`
	} `json:"flag_groups"`
	Effects  map[string]bool `json:"effects"`
	Modes    map[string]bool `json:"modes"`
	Examples []string        `json:"examples"`
}

func agree(page, answer string) error {
	var a struct{ Commands []described }
	if err := json.Unmarshal([]byte(answer), &a); err != nil || len(a.Commands) != 1 || a.Commands[0].Args == nil {
		return fmt.Errorf("want the detail of one command, got %s", answer)
	}
	d := a.Commands[0]
	lines := strings.Split(page, "\n")
	if lines[0] != d.Summary {
		return fmt.Errorf("the page starts with %q, the summary is %q", lines[0], d.Summary)
	}
	heading := 2
	if d.Effects["user_writes"] || d.Effects["network"] || d.Effects["paid"] {
		heading = 4
		if lines[2] == "" || strings.Contains(lines[2], "--dry-run") != d.Modes["dry_run"] {
			return fmt.Errorf("the line of effects %q disagrees with dry_run %t", lines[2], d.Modes["dry_run"])
		}
	}
	if len(lines) <= heading+1 || lines[heading-1] != "" || !strings.HasPrefix(lines[heading+1], "  ncly "+d.Path) {
		return fmt.Errorf("want the usage line of %s after line %d, got %q", d.Path, heading, lines)
	}
	usage := strings.Fields(strings.TrimPrefix(lines[heading+1], "  ncly "+d.Path))
	values := map[string]string{}
	for _, f := range d.Flags {
		values[f.Name] = f.Value
	}
	flag := func(name string) string {
		if values[name] != "" {
			return "--" + name + " <" + values[name] + ">"
		}
		return "--" + name
	}
	var want []string
	for _, arg := range d.Args {
		name := arg.Name
		if arg.Variadic {
			name += "..."
		}
		if arg.Required {
			want = append(want, "<"+name+">")
		} else {
			want = append(want, "["+name+"]")
		}
	}
	grouped := map[string]bool{}
	for _, g := range d.FlagGroups {
		for _, name := range g.Flags {
			grouped[name] = true
		}
	}
	for _, f := range d.Flags {
		if f.Required && !grouped[f.Name] {
			want = append(want, flag(f.Name))
		}
	}
	for _, g := range d.FlagGroups {
		var choices []string
		for _, name := range g.Flags {
			choices = append(choices, flag(name))
		}
		group := strings.Join(choices, " | ")
		if g.Min > 0 {
			want = append(want, "("+group+")")
		} else {
			want = append(want, "["+group+"]")
		}
	}
	want = append(want, "[flags]")
	if got, wanted := strings.Join(usage, " "), strings.Join(want, " "); got != wanted {
		return fmt.Errorf("usage %q, want %q", got, wanted)
	}
	for _, f := range d.Flags {
		name := "    --" + f.Name
		if f.Shorthand != "" {
			name = "-" + f.Shorthand + ", --" + f.Name
		}
		if f.Value != "" {
			name += " <" + f.Value + ">"
		}
		line := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(name) + ` +` + regexp.QuoteMeta(f.Summary) + `$`)
		if !line.MatchString(page) {
			return fmt.Errorf("no line for the flag %s with its summary %q", name, f.Summary)
		}
	}
	for _, e := range d.Examples {
		if !strings.Contains(page, "\n  $ "+e+"\n") {
			return fmt.Errorf("no example %q", e)
		}
	}
	return nil
}

// cmdSchema validates an answer against the schema of the describe answer
// for its command, with a JSON Schema validator of draft 2020-12 that never
// fetches a remote schema: schema <answer-file> <describe-file> <schema|dry_run_schema>.
// Negated, it asserts that the answer breaks that schema.
func cmdSchema(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 3 || (args[2] != "schema" && args[2] != "dry_run_schema") {
		ts.Fatalf("usage: schema <answer-file> <describe-file> <schema|dry_run_schema>")
	}
	var described struct {
		Commands []struct {
			Result map[string]json.RawMessage `json:"result"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(ts.ReadFile(args[1])), &described); err != nil || len(described.Commands) != 1 {
		ts.Fatalf("schema: want the detail of one command in %s", args[1])
	}
	raw, found := described.Commands[0].Result[args[2]]
	if !found {
		ts.Fatalf("schema: the result has no %s", args[2])
	}
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(raw)))
	ts.Check(err)
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(noRemote{})
	ts.Check(compiler.AddResource("describe.json", doc))
	schema, err := compiler.Compile("describe.json")
	ts.Check(err)
	answer, err := jsonschema.UnmarshalJSON(strings.NewReader(ts.ReadFile(args[0])))
	ts.Check(err)
	switch err := schema.Validate(answer); {
	case neg && err == nil:
		ts.Fatalf("schema: %s fits its %s", args[0], args[2])
	case !neg && err != nil:
		ts.Fatalf("schema: %s breaks its %s: %v", args[0], args[2], err)
	}
}

// noRemote refuses to load any schema, so a reference that is not resolved
// within the descriptor fails instead of reaching the network.
type noRemote struct{}

func (noRemote) Load(url string) (any, error) { return nil, fmt.Errorf("no remote schema: %s", url) }

// printCodes prints every code of core's registry, one per line.
func printCodes() {
	for _, code := range contract.Core.Codes() {
		fmt.Println(code)
	}
	os.Exit(0)
}

// mentionedCommand finds a code span of a skill that starts with ncly and a
// lowercase word, such as `ncly skill view nicely`, and captures its words.
// Prose such as "every ncly call" is no command.
var mentionedCommand = regexp.MustCompile("`ncly((?: [a-z][a-z0-9-]*)+)[^`]*`")

// mentionedCode finds a word of the form of a code. A word before "=" or
// one that starts with NCLY_ is an environment variable instead.
var mentionedCode = regexp.MustCompile(`\b[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+\b=?`)

// cmdMentions fails when a skill names an ncly command in a code span that
// describe does not give, or a code that the registry lacks, or holds more
// than 60 lines: mentions <skill-file> <describe-file> <codes-file>. After
// a command path, a word is a command only when that path groups others.
func cmdMentions(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 3 {
		ts.Fatalf("usage: mentions <skill-file> <describe-file> <codes-file>")
	}
	skill := ts.ReadFile(args[0])
	var listing struct {
		Commands []struct {
			Path string `json:"path"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(ts.ReadFile(args[1])), &listing); err != nil {
		ts.Fatalf("mentions: %s: %v", args[1], err)
	}
	paths := map[string]bool{}
	groups := map[string]bool{}
	for _, c := range listing.Commands {
		paths[c.Path] = true
		if i := strings.LastIndex(c.Path, " "); i > 0 {
			groups[c.Path[:i]] = true
		}
	}
	codes := strings.Fields(ts.ReadFile(args[2]))
	var problems []string
	if lines := strings.Count(skill, "\n"); lines > 60 {
		problems = append(problems, fmt.Sprintf("%d lines, more than 60", lines))
	}
	for _, m := range mentionedCommand.FindAllStringSubmatch(skill, -1) {
		words := strings.Fields(m[1])
		path := words[0]
		for _, word := range words[1:] {
			if !groups[path] {
				break
			}
			path += " " + word
		}
		if !paths[path] {
			problems = append(problems, fmt.Sprintf("the command %q, which ncly lacks", "ncly "+path))
		}
	}
	for _, word := range mentionedCode.FindAllString(skill, -1) {
		if strings.HasSuffix(word, "=") || strings.HasPrefix(word, "NCLY_") {
			continue
		}
		if !slices.Contains(codes, word) {
			problems = append(problems, fmt.Sprintf("the code %s, which ncly lacks", word))
		}
	}
	switch {
	case neg && len(problems) == 0:
		ts.Fatalf("mentions: %s names only what ncly has", args[0])
	case !neg && len(problems) > 0:
		ts.Fatalf("mentions: %s names %s", args[0], strings.Join(problems, "; "))
	}
}
