package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A scenario cannot place another process's write between the read of
// preparation and the lock, so these tests call changeGrants on a file that
// changed after the caller read it.

func TestChangeGrantsGivesChangeTheFileAsTheLockFindsIt(t *testing.T) {
	file := filepath.Join(t.TempDir(), "grants.json")
	written := `{"format_version":1,"grants":[{"service":"deepgram","domain":"hello","path":null,"granted_at":"2026-10-17T04:00:00Z"}]}` + "\n"
	if err := os.WriteFile(file, []byte(written), 0o600); err != nil {
		t.Fatal(err)
	}
	var seen []grant
	err := changeGrants(context.Background(), file, func(current []grant) ([]grant, error) {
		seen = current
		return nil, nil
	})
	if err != nil || findGrant(seen, "deepgram", "hello") == nil {
		t.Fatalf("change saw %v, %v; want the grant that the file holds", seen, err)
	}
	if data, _ := os.ReadFile(file); string(data) != written {
		t.Errorf("a change that returned nil rewrote the file: %s", data)
	}
}

func TestChangeGrantsChangesNothingInAFileThatBecameInvalid(t *testing.T) {
	file := filepath.Join(t.TempDir(), "grants.json")
	if err := os.WriteFile(file, []byte(`{"format_version":2,"grants":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	err := changeGrants(context.Background(), file, func(current []grant) ([]grant, error) {
		called = true
		return current, nil
	})
	var invalid *grantsError
	if !errors.As(err, &invalid) || !invalid.otherVersion || invalid.version != 2 {
		t.Fatalf("changeGrants returned %v; want the grantsError of format_version 2", err)
	}
	if called {
		t.Error("change ran on a file that this ncly does not read")
	}
	if data, _ := os.ReadFile(file); string(data) != `{"format_version":2,"grants":[]}` {
		t.Errorf("the file changed: %s", data)
	}
}

func TestChangeGrantsChangesNothingOnceItsContextEnded(t *testing.T) {
	// A signal can arrive after the human confirmed and before the lock,
	// which is free.
	file := filepath.Join(t.TempDir(), "grants.json")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := changeGrants(ctx, file, func(current []grant) ([]grant, error) {
		called = true
		return append(current, grant{Service: "deepgram", Domain: "hello"}), nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("changeGrants returned %v; want the error of its context", err)
	}
	if called {
		t.Error("change ran after the context ended")
	}
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the grants file exists: %v", err)
	}
}
