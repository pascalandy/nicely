package main

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){"ncly": main})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("..", "..", "testdata", "script"),
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup:               setupHome,
		Cmds:                map[string]func(*testscript.TestScript, bool, []string){"exits": cmdExits},
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
