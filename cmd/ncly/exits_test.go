package main

import (
	"errors"
	"os/exec"
	"strconv"

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
	observeCall(ts, args[1], args[2:], got)
}
