package platform

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteFileReplacesTheFileAndLeavesNoTemporaryFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state", "nicely")
	path := filepath.Join(dir, "grants.json")
	for _, data := range []string{"first\n", "second\n"} {
		if err := WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != data {
			t.Fatalf("the file holds %q, %v; want %q", got, err, data)
		}
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("the file has the mode %v, %v; want 0600", info.Mode(), err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("the folder holds %d entries; want the file alone", len(entries))
	}
}

func TestWriteFileKeepsTheFileWhenTheWriteFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grants.json")
	if err := os.WriteFile(path, []byte("kept\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A folder in place of the file makes the rename fail.
	blocked := filepath.Join(dir, "blocked")
	if err := os.MkdirAll(filepath.Join(blocked, "inside"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(blocked, []byte("lost\n"), 0o600); err == nil {
		t.Fatal("WriteFile replaced a folder that holds a file")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("the folder holds %d entries; want the file and the folder alone", len(entries))
	}
}

func TestLockWaitsForTheOwnerThenFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "grants.lock")
	unlock, err := Lock(context.Background(), path, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Each Lock opens the file again, and flock locks an open file, so a
	// second Lock in the same process competes as another process would.
	start := time.Now()
	if _, err := Lock(context.Background(), path, 100*time.Millisecond); !errors.Is(err, ErrLocked) {
		t.Fatalf("a second Lock returned %v; want ErrLocked", err)
	}
	if waited := time.Since(start); waited < 100*time.Millisecond {
		t.Errorf("a second Lock gave up after %v; want at least the wait", waited)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Lock(ctx, path, time.Second); !errors.Is(err, context.Canceled) {
		t.Errorf("a Lock whose ctx ended returned %v; want its error", err)
	}
	unlock()
	again, err := Lock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("Lock after the release returned %v", err)
	}
	again()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the lock file is gone: %v", err)
	}
}

func TestLockTakesNoFreeLockOnceItsContextEnded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "grants.lock")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Lock(ctx, path, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("a Lock whose ctx ended took the free lock: %v", err)
	}
	unlock, err := Lock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("the refused Lock still holds the lock: %v", err)
	}
	unlock()
}
