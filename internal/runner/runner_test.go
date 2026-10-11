package runner_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/pascalandy/nicely/internal/runner"
)

// TestRunNeverStartsAProgramOnceCtxEnded proves that a signal that came
// before the start stops the program from running at all.
func TestRunNeverStartsAProgramOnceCtxEnded(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(runner.SignalCause{Signal: os.Interrupt})
	started := filepath.Join(t.TempDir(), "started")
	result, err := runner.Run(ctx, runner.Program{Path: "/bin/sh", Args: []string{"-c", "touch " + started}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Exit != -1 || result.Signal != os.Interrupt {
		t.Fatalf("Run returned %+v, want exit -1 and the interrupt", result)
	}
	if _, err := os.Stat(started); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the program ran: %v", err)
	}
}
