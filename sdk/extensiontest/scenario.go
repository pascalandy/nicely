package extensiontest

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
)

// commands returns the scenario commands that the kit adds, which the
// package comment lists.
func (k *kit) commands() map[string]func(*testscript.TestScript, bool, []string) {
	direct := "ncly-" + k.manifest.Slug()
	expect := func(program string) func(*testscript.TestScript, bool, []string) {
		return func(ts *testscript.TestScript, neg bool, args []string) {
			switch exit := k.call(ts, program, args); {
			case neg && exit == 0:
				ts.Fatalf("%s: unexpected success", program)
			case !neg && exit != 0:
				ts.Fatalf("%s: exit %d", program, exit)
			}
		}
	}
	return map[string]func(*testscript.TestScript, bool, []string){
		"ncly":  expect("ncly"),
		direct:  expect(direct),
		"exits": k.exits,
		"grant": k.grant,
		"fake":  k.fakeChild,
	}
}

// call runs ncly, or ncly-<slug> directly, checks the answer, and
// returns the exit code.
func (k *kit) call(ts *testscript.TestScript, program string, args []string) int {
	path, words := filepath.Join(k.bin, "ncly"), args
	if program != "ncly" {
		path, words = k.installed(ts.Getenv), append(k.manifest.Root(), args...)
	}
	exit, err := exitCode(ts.Exec(path, args...))
	if err != nil {
		ts.Fatalf("%s: %v", program, err)
	}
	if problem := k.checkAnswer(words, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"), ts.Getenv("NCLY_JSON")); problem != "" {
		ts.Fatalf("`%s %s`: %s", program, strings.Join(args, " "), problem)
	}
	return exit
}

// exits runs a command and checks its exact exit code, which ! cannot do:
// exits <code> <command> [args...]. ncly and ncly-<slug> run as the
// commands of the same name do.
func (k *kit) exits(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 2 {
		ts.Fatalf("usage: exits <code> <command> [args...]")
	}
	want, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("exits: invalid code %q", args[0])
	}
	var got int
	if args[1] == "ncly" || args[1] == "ncly-"+k.manifest.Slug() {
		got = k.call(ts, args[1], args[2:])
	} else if got, err = exitCode(ts.Exec(args[1], args[2:]...)); err != nil {
		ts.Fatalf("exits: %v", err)
	}
	if got != want {
		ts.Fatalf("%s exited with %d, want %d", args[1], got, want)
	}
}

// exitCode returns the exit code of a command that ts.Exec ran. A signal
// that ended it gives 128 plus its number, as a shell reports it. A command
// that never ran gives its error.
func exitCode(err error) (int, error) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return 0, err
	}
	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal()), nil
	}
	return exitErr.ExitCode(), nil
}

// grant grants the key of a service that a command declares to the
// extension at its scratch path: grant <service>. It writes what ncly auth
// grant writes for a human. The key stays the sentinel of the service.
func (k *kit) grant(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 || !slices.Contains(k.services(), args[0]) {
		ts.Fatalf("usage: grant <service>, where a command of the manifest declares the service")
	}
	type grant struct {
		Service   string `json:"service"`
		Domain    string `json:"domain"`
		Path      string `json:"path"`
		GrantedAt string `json:"granted_at"`
	}
	var file struct {
		FormatVersion int     `json:"format_version"`
		Grants        []grant `json:"grants"`
	}
	path := filepath.Join(ts.Getenv("XDG_STATE_HOME"), "nicely", "grants.json")
	if data, err := os.ReadFile(path); err == nil {
		ts.Check(json.Unmarshal(data, &file))
	}
	file.FormatVersion = 1
	file.Grants = slices.DeleteFunc(file.Grants, func(g grant) bool { return g.Service == args[0] && g.Domain == k.manifest.Slug() })
	file.Grants = append(file.Grants, grant{args[0], k.manifest.Slug(), k.installed(ts.Getenv), time.Now().UTC().Format(time.RFC3339)})
	data, err := json.Marshal(file)
	ts.Check(err)
	ts.Check(os.MkdirAll(filepath.Dir(path), 0o755))
	ts.Check(os.WriteFile(path, append(data, '\n'), 0o644))
}

// fakeChild puts the fake child in $WORK/bin as name, in place of any file
// of that name. It runs the actions of $WORK/bin/<name>.fake: fake <name>
func (k *kit) fakeChild(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 || strings.ContainsRune(args[0], filepath.Separator) {
		ts.Fatalf("usage: fake <name>")
	}
	path := filepath.Join(ts.Getenv("WORK"), "bin", args[0])
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		ts.Fatalf("fake: %v", err)
	}
	ts.Check(link(k.fake, path))
}

// checkAnswer checks the answer of a call of the extension against the
// schema of its command, as ncly describe gives it. words holds the words
// after ncly. It returns the problem, or "" for a call that prints no
// answer, such as a help page, human text, or a call outside the root.
func (k *kit) checkAnswer(words []string, exit int, stdout, stderr, jsonVariable string) string {
	call, found := k.find(words)
	if !found {
		return ""
	}
	set := func(name string, fallback bool) bool {
		if on, given := call.line.Switch(name); given {
			return on
		}
		return fallback
	}
	d, runs := call.command, call.runs
	if set("help", false) || set("version", false) || !set("json", jsonVariable == "1") || runs && exit == 0 && d.Result.Kind == contract.ResultText {
		return ""
	}
	kind, data := contract.ResultObject, map[string]any(nil)
	if runs {
		kind, data = cmp.Or(d.Result.Kind, contract.ResultObject), d.Result.Schema
	}
	switch {
	case kind == contract.ResultText:
		kind = contract.ResultObject
	case runs && d.Modes.DryRun && set("dry-run", false):
		kind, data = contract.ResultObject, d.Result.DryRunSchema
	case !runs && exit == 0:
		return ""
	}
	stream, answer := "stdout", stdout
	if !(contract.Reply{Exit: exit, Stdout: []byte(stdout)}).OnStdout(kind) {
		lines := strings.SplitAfter(strings.TrimRight(stderr, "\n"), "\n")
		stream, answer = "stderr", lines[len(lines)-1]+"\n"
	}
	line, ok := strings.CutSuffix(answer, "\n")
	if !ok || line == "" || strings.Contains(line, "\n") {
		return fmt.Sprintf("want one line of answer on %s, got %q", stream, answer)
	}
	if err := validate(contract.AnswerSchema(kind, data), line); err != nil {
		return fmt.Sprintf("the answer on %s breaks the schema that manifest.toml declares for `ncly %s`: %v. Fix: make the handler answer what the schema declares, or correct the schema", stream, call.path, err)
	}
	return ""
}

// parsed is what the parser of ncly finds in a call of the extension.
type parsed struct {
	// path is the command path, such as "hello run".
	path    string
	command contract.Command
	runs    bool
	// line holds the switches of the call, which command.Scan reads as the
	// parser does, past an unknown flag or an invalid value.
	line command.Line
}

// find returns what the parser of ncly finds in a call, given as the words
// after ncly. A global flag may come before the domain. It returns false
// for a call of core or outside the root.
func (k *kit) find(words []string) (parsed, bool) {
	root := &cobra.Command{Use: "ncly"}
	command.AddGlobalFlags(root, k.printer)
	command.AddManifest(root, k.printer, k.printer, k.manifest, func(contract.Command) func(*cobra.Command, []string) error {
		return func(*cobra.Command, []string) error { return nil }
	})
	c, _, err := root.Find(words)
	if err != nil || c == root {
		return parsed{}, false
	}
	call := parsed{path: strings.TrimPrefix(c.CommandPath(), "ncly "), line: command.Scan(c, words)}
	var found bool
	call.command, found = k.manifest.Command(call.path)
	call.runs = found && k.manifest.Runs(call.command)
	return call, true
}

// validate checks an answer against a schema of draft 2020-12, with a
// loader that refuses every remote schema.
func validate(schema map[string]any, answer string) error {
	raw, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(raw)))
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(noRemote{})
	if err := compiler.AddResource("answer.json", doc); err != nil {
		return err
	}
	compiled, err := compiler.Compile("answer.json")
	if err != nil {
		return err
	}
	value, err := jsonschema.UnmarshalJSON(strings.NewReader(answer))
	if err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return compiled.Validate(value)
}

// noRemote refuses to load any schema, so a reference outside the schema
// fails instead of reaching the network.
type noRemote struct{}

func (noRemote) Load(url string) (any, error) { return nil, fmt.Errorf("no remote schema: %s", url) }
