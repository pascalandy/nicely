package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheActionsOfTheFakeChild runs each action that returns. The scenarios
// of the standalone fixture run the fake child itself, exit included.
func TestTheActionsOfTheFakeChild(t *testing.T) {
	cases := []struct {
		name, actions, stdin, stdout, stderr, err string
	}{
		{name: "streams", actions: "# a comment\n\nstdout one two\nstderr three\n", stdout: "one two\n", stderr: "three\n"},
		{name: "stdin", actions: "stdin\nstdout end\n", stdin: "fed\n", stdout: "fed\nend\n"},
		{name: "sleep", actions: "sleep 1ms\nstdout woke\n", stdout: "woke\n"},
		{name: "unknown action", actions: "stdout before\nshout loud\n", stdout: "before\n", err: `:2: unknown action "shout"`},
		{name: "bad duration", actions: "sleep soon\n", err: `:1: time: invalid duration "soon"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tool.fake")
			if err := os.WriteFile(path, []byte(c.actions), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			err := run(path, strings.NewReader(c.stdin), &stdout, &stderr)
			if c.err == "" && err != nil || c.err != "" && (err == nil || !strings.HasSuffix(err.Error(), c.err)) {
				t.Errorf("run = %v, want an error ending with %q", err, c.err)
			}
			if stdout.String() != c.stdout || stderr.String() != c.stderr {
				t.Errorf("stdout %q and stderr %q, want %q and %q", stdout.String(), stderr.String(), c.stdout, c.stderr)
			}
		})
	}
}

// TestTheFakeChildPrintsNamesNeverValues proves that env shows which
// variables reach the child, such as a key, without the key.
func TestTheFakeChildPrintsNamesNeverValues(t *testing.T) {
	t.Setenv("FAKE_TEST_KEY", "never printed")
	path := filepath.Join(t.TempDir(), "tool.fake")
	if err := os.WriteFile(path, []byte("env\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := run(path, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatal(err)
	}
	names := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	if !strings.Contains("\n"+stdout.String(), "\nFAKE_TEST_KEY\n") || strings.Contains(stdout.String(), "never printed") {
		t.Errorf("env printed %q, want the name FAKE_TEST_KEY without its value", stdout.String())
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("env printed %q before %q, want the names sorted", names[i-1], names[i])
		}
	}
}

func TestAMissingFileFails(t *testing.T) {
	err := run(filepath.Join(t.TempDir(), "tool.fake"), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if !os.IsNotExist(err) {
		t.Errorf("run = %v, want a missing file", err)
	}
}
