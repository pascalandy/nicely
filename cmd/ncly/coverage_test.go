package main

import (
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
	checks *extensiontest.Coverage
	slug   string
	root   []string
}

// setupCoverage registers the kit for first-party extension scenarios.
// Partial scenario selections check answers but do not require every command.
func setupCoverage(t *testing.T, dir string, setup func(*testscript.Env) error) func(*testscript.Env) error {
	t.Helper()
	slug := filepath.Base(filepath.Dir(filepath.Dir(dir)))
	if slug != "agent" && slug != "transcript" {
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
	t.Cleanup(func() {
		if started.Load() != int64(len(files)) {
			return
		}
		for _, problem := range checks.Missing() {
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
		env.Values[coverageKey{}] = scenarioCoverage{checks: checks, slug: slug, root: m.Root()}
		return nil
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
	if name != program {
		ts.Fatalf("use the registered command %s instead of its path, so the kit checks its answer", name)
	}
	if problem := c.checks.Observe(words, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"), ts.Getenv("NCLY_JSON")); problem != "" {
		ts.Fatalf("`%s %s`: %s", program, strings.Join(args, " "), problem)
	}
}
