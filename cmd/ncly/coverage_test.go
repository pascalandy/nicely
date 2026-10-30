package main

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extensiontest"
	"github.com/rogpeppe/go-internal/testscript"
)

type coverageKey struct{}

type scenarioCoverage struct {
	checks   *extensiontest.Coverage
	slug     string
	root     []string
	commands map[string]func(*testscript.TestScript, bool, []string)
}

// setupCoverage registers the kit for first-party extension scenarios.
// Partial scenario selections check answers but do not require every command.
func setupCoverage(t *testing.T, dir string, setup func(*testscript.Env) error) func(*testscript.Env) error {
	t.Helper()
	slug := filepath.Base(filepath.Dir(filepath.Dir(dir)))
	if slug != "agent" && slug != "transcript" && slug != "hello" {
		return setup
	}
	e := fixtures[slug]
	checks, err := extensiontest.NewCoverage(e)
	if err != nil {
		t.Fatal(err)
	}
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.txt*"))
	if err != nil {
		t.Fatal(err)
	}
	files = slices.DeleteFunc(files, func(file string) bool {
		return filepath.Ext(file) != ".txt" && filepath.Ext(file) != ".txtar"
	})
	var started atomic.Int64
	host, err := exec.LookPath("ncly")
	if err != nil {
		t.Fatal(err)
	}
	commands := checks.ScenarioCommands(host)
	t.Cleanup(func() {
		for _, reservation := range checks.Reservations() {
			t.Log(reservation)
		}
		if started.Load() != int64(len(files)) {
			return
		}
		missing := checks.Missing
		if slug == "hello" {
			// hello exercises signal commands in the background, which earns no foreground call coverage.
			missing = checks.MissingCodes
		}
		for _, problem := range missing() {
			t.Error(problem)
		}
	})
	return func(env *testscript.Env) error {
		started.Add(1)
		if err := setup(env); err != nil {
			return err
		}
		if env.Values == nil {
			env.Values = map[any]any{}
		}
		env.Values[coverageKey{}] = scenarioCoverage{checks: checks, slug: slug, root: m.Root(), commands: commands}
		return nil
	}
}

// kitCommand shares the kit's paired commands and file helpers with first-party extension scenarios.
func kitCommand(name string) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		c, ok := ts.Value(coverageKey{}).(scenarioCoverage)
		if !ok {
			ts.Fatalf("%s: this scenario has no extension coverage adapter", name)
		}
		c.commands[name](ts, neg, args)
	}
}

// observeCall checks a completed foreground call without changing its streams or execution.
func observeCall(ts *testscript.TestScript, program string, args []string, exit int) {
	c, ok := ts.Value(coverageKey{}).(scenarioCoverage)
	if !ok {
		return
	}
	name := filepath.Base(program)
	words := args
	switch name {
	case "ncly":
	case "ncly-" + c.slug:
		words = append(append([]string{}, c.root...), args...)
	default:
		return
	}
	if problem := c.checks.Observe(words, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"), ts.Getenv("NCLY_JSON")); problem != "" {
		ts.Fatalf("`%s %s`: %s", program, strings.Join(args, " "), problem)
	}
}
