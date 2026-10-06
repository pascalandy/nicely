package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){"ncly": main, "closedpipe": closedPipe})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("..", "..", "testdata", "script"),
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup:               setupHome,
		Condition:           condition,
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"exits":     cmdExits,
			"answer":    cmdAnswer,
			"snapshot":  cmdSnapshot,
			"unchanged": cmdUnchanged,
			"pseudo":    cmdPseudo,
		},
	})
}

// setupHome keeps every file that ncly reads or writes inside $WORK.
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
