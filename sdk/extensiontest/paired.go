package extensiontest

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/rogpeppe/go-internal/testscript"
)

// scriptFiles holds the selections of one scenario.
// protect lasts until the next protect call, and next until the next signal or appears call.
type scriptFiles struct {
	protect, next, survives []string
}

// ScenarioCommands supplies signal, appears, refuses, protect, and survives to a testscript runner.
// host is the registered ncly's absolute path, independent of the scenario's PATH.
func (c *Coverage) ScenarioCommands(host string) map[string]func(*testscript.TestScript, bool, []string) {
	var mu sync.Mutex
	files := map[*testscript.TestScript]*scriptFiles{}
	state := func(ts *testscript.TestScript) *scriptFiles {
		mu.Lock()
		defer mu.Unlock()
		if files[ts] == nil {
			files[ts] = &scriptFiles{}
			ts.Defer(func() {
				mu.Lock()
				defer mu.Unlock()
				delete(files, ts)
			})
		}
		return files[ts]
	}
	selectFiles := func(survives bool) func(*testscript.TestScript, bool, []string) {
		return func(ts *testscript.TestScript, neg bool, args []string) {
			if neg || len(args) == 0 {
				ts.Fatalf("usage: protect <path>... or survives <file>...")
			}
			paths := make([]string, len(args))
			for i, path := range args {
				paths[i] = ts.MkAbs(path)
			}
			if survives {
				state(ts).survives = paths
			} else {
				state(ts).protect, state(ts).next = paths, paths
			}
		}
	}
	return map[string]func(*testscript.TestScript, bool, []string){
		"protect": selectFiles(false), "survives": selectFiles(true),
		"signal": func(ts *testscript.TestScript, neg bool, args []string) {
			files := state(ts)
			next := files.next
			files.next = nil
			defer func() { files.survives = nil }()
			c.signal(ts, neg, args, files.survives, next, host)
		},
		"appears": func(ts *testscript.TestScript, neg bool, args []string) {
			files := state(ts)
			next := files.next
			files.next = nil
			c.appears(ts, neg, args, next, host)
		},
		"refuses": func(ts *testscript.TestScript, neg bool, args []string) {
			c.refuses(ts, neg, args, state(ts).protect, host)
		},
	}
}

// pairedCall resolves only registered calls and requires a structured JSON answer.
func (c *Coverage) pairedCall(ts *testscript.TestScript, program string, args []string, host string) (string, []string, parsed) {
	path, words := program, slices.Clone(args)
	switch program {
	case "ncly":
		if !filepath.IsAbs(host) {
			ts.Fatalf("register ncly with its absolute executable path")
		}
		path = host
	case "ncly-" + c.manifest.Slug():
		path = filepath.Join(ts.Getenv("XDG_DATA_HOME"), "nicely", "extensions", program)
		words = append(c.manifest.Root(), words...)
	default:
		ts.Fatalf("use the registered command ncly or ncly-%s instead of %s", c.manifest.Slug(), program)
	}
	call, found := c.find(words)
	if !found || !call.runs {
		ts.Fatalf("%s: expected a runnable command of %s", program, c.manifest.Slug())
	}
	if on, given := call.line.Switch("json"); given && !on {
		ts.Fatalf("%s: signal, appears, and refuses require JSON", program)
	}
	if on, given := call.line.Switch("no-input"); given && !on {
		ts.Fatalf("%s: signal, appears, and refuses require non-interactive mode", program)
	}
	if on, _ := call.line.Switch("help"); on {
		ts.Fatalf("%s: signal, appears, and refuses cannot check help", program)
	}
	if on, _ := call.line.Switch("version"); on {
		ts.Fatalf("%s: signal, appears, and refuses cannot check version", program)
	}
	return path, words, call
}

func machineMode(ts *testscript.TestScript) func() {
	json, noInput := ts.Getenv("NCLY_JSON"), ts.Getenv("NCLY_NO_INPUT")
	ts.Setenv("NCLY_JSON", "1")
	ts.Setenv("NCLY_NO_INPUT", "1")
	return func() {
		ts.Setenv("NCLY_JSON", json)
		ts.Setenv("NCLY_NO_INPUT", noInput)
	}
}

func (c *Coverage) checkedCode(words []string, exit int, stdout, stderr string) (contract.Code, string) {
	if problem := c.Observe(words, exit, stdout, stderr, "1"); problem != "" {
		return "", fmt.Sprintf("`ncly %s`: %s", strings.Join(words, " "), problem)
	}
	if exit == 0 {
		return "", ""
	}
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	var answer struct{ Errors []contract.Problem }
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &answer); err != nil || len(answer.Errors) == 0 {
		return "", fmt.Sprintf("want a refusal answer on stderr, got %q", stderr)
	}
	return answer.Errors[0].Code, ""
}

func (c *Coverage) refuses(ts *testscript.TestScript, neg bool, args, protected []string, host string) {
	if neg || len(args) < 2 {
		ts.Fatalf("usage: refuses <code> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, call := c.pairedCall(ts, args[1], args[2:], host)
	if _, given := call.line.Switch("dry-run"); given || !call.command.Modes.DryRun {
		ts.Fatalf("refuses: give a command with a dry run, without a dry-run flag")
	}
	before, err := snapshotFiles(protected, false)
	ts.Check(err)
	// Exec consumes testscript's stdin. Save it once so both calls receive identical bytes.
	dir, err := os.MkdirTemp("", "ncly-refuses-")
	ts.Check(err)
	defer func() { _ = os.RemoveAll(dir) }()
	input := filepath.Join(dir, "stdin")
	ts.Check(ts.Exec("/bin/sh", "-c", `/bin/cat > "$1"`, "refuses", input))
	// The explicit value lets command discovery skip this flag before any command word.
	// Prefixing it preserves separators, including a literal -- used as another flag's value.
	preview := append([]string{"--dry-run=true"}, args[2:]...)
	wantExit := 0
	for i, arguments := range [][]string{preview, args[2:]} {
		mode := []string{"dry run", "run"}[i]
		invocation := append([]string{"-c", `input=$1; shift; exec "$@" < "$input"`, "refuses", input, program}, arguments...)
		exit, err := exitCode(ts.Exec("/bin/sh", invocation...))
		ts.Check(err)
		observed := words
		if i == 0 {
			observed = slices.Clone(preview)
			if args[1] == "ncly-"+c.manifest.Slug() {
				observed = append(c.manifest.Root(), preview...)
			}
		}
		code, problem := c.checkedCode(observed, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"))
		if problem != "" {
			ts.Fatalf("refuses: %s: %s", mode, problem)
		}
		if exit == 0 {
			ts.Fatalf("refuses: %s exited with 0, want a refusal with %s", mode, args[0])
		}
		if string(code) != args[0] {
			ts.Fatalf("refuses: %s returned %s, want %s", mode, code, args[0])
		}
		if i == 0 {
			wantExit = exit
		} else if exit != wantExit {
			ts.Fatalf("refuses: run exited with %d, dry run exited with %d", exit, wantExit)
		}
		after, err := snapshotFiles(protected, false)
		ts.Check(err)
		if !maps.Equal(before, after) {
			ts.Fatalf("refuses: %s changed protected paths: %s", mode, fmt.Sprint(protected))
		}
	}
}
