package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/runner"
	hello "github.com/pascalandy/nicely/testdata/extensions/hello"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"ncly":           testNcly,
		"closedpipe":     closedPipe,
		"exitcode":       exitCode,
		"hello-external": func() { shortCleanup(); extension.Main(hello.Extension) },
	})
}

// fixtures holds the Go fixture extensions of testdata/extensions/, by
// domain. The setup of a fixture's scenarios installs it in the data folder.
var fixtures = map[string]extension.Extension{"hello": hello.Extension}

// testNcly runs ncly with a short cleanup.
func testNcly() {
	shortCleanup()
	main()
}

// shortCleanup gives cleanup 4 seconds instead of 15, so that a scenario
// reaches the SIGKILL of cleanup quickly. The drain limit keeps its 2
// seconds, so the scenarios still tell the two clocks apart.
func shortCleanup() { runner.Cleanup = 4 * time.Second }

// TestScripts runs core's scenarios, then those of each extension and of
// each fixture extension, which keep theirs in their own folder.
func TestScripts(t *testing.T) {
	root := filepath.Join("..", "..")
	t.Run("core", func(t *testing.T) { runScripts(t, filepath.Join(root, "testdata", "script"), setupHome) })
	extensions, _ := filepath.Glob(filepath.Join(root, "extensions", "*", "testdata", "script"))
	fixtureDirs, _ := filepath.Glob(filepath.Join(root, "testdata", "extensions", "*", "testdata", "script"))
	for _, dir := range append(extensions, fixtureDirs...) {
		domain := filepath.Base(filepath.Dir(filepath.Dir(dir)))
		t.Run(domain, func(t *testing.T) {
			runScripts(t, dir, func(env *testscript.Env) error {
				if err := setupHome(env); err != nil {
					return err
				}
				if e, isFixture := fixtures[domain]; isFixture {
					return install(env, domain, e)
				}
				return nil
			})
		})
	}
}

func runScripts(t *testing.T, dir string, setup func(*testscript.Env) error) {
	testscript.Run(t, testscript.Params{
		Dir:                 dir,
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup:               setup,
		Condition:           condition,
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"exits":     cmdExits,
			"answer":    cmdAnswer,
			"snapshot":  cmdSnapshot,
			"unchanged": cmdUnchanged,
			"pseudo":    cmdPseudo,
			"waitfile":  cmdWaitfile,
			"gone":      cmdGone,
			"term":      cmdTerm,
		},
	})
}

// install puts a Go fixture in the data folder as an external extension:
// the script ncly-<domain> that runs it, with its manifest and catalogs
// beside it. The script runs the test binary under a name other than
// ncly-<domain>, so the fixture never runs as a bundled extension here.
func install(env *testscript.Env, domain string, e extension.Extension) error {
	dir := filepath.Join(env.WorkDir, ".local", "share", "nicely", "extensions")
	executable := filepath.Join(dir, "ncly-"+domain)
	locales, err := fs.Sub(e.Locales, "locales")
	if err == nil {
		err = os.CopyFS(executable+".locales", locales)
	}
	if err == nil {
		err = os.WriteFile(executable+".toml", e.Manifest, 0o644)
	}
	if err == nil {
		err = os.WriteFile(executable, []byte("#!/bin/sh\nexec "+domain+"-external \"$@\"\n"), 0o755)
	}
	env.Setenv("DATA", dir)
	return err
}

// setupHome points HOME and the XDG folders at $WORK, so ncly never touches
// the host's own config, data, state, or cache.
func setupHome(env *testscript.Env) error {
	home := env.WorkDir
	env.Setenv("HOME", home)
	env.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	env.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	env.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	env.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	return nil
}

// condition answers [bash-completion], which holds when the bash-completion
// package is installed, as the generated bash script needs it.
func condition(cond string) (bool, error) {
	if cond != "bash-completion" {
		return false, fmt.Errorf("unknown condition %q", cond)
	}
	for _, path := range []string{
		"/usr/share/bash-completion/bash_completion",
		"/opt/homebrew/share/bash-completion/bash_completion",
	} {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}
	return false, nil
}
