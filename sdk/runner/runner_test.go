package runner_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pascalandy/nicely/sdk/runner"
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

// TestRunStopsAfterTheExitOnlyForASignal proves that, once the program has
// exited, only a signal that ends ctx during the settle wait stops it. Such a
// signal is a Ctrl-C that reaches the program and ncly together. A deadline
// or a plain cancellation that ends ctx then keeps the program's own result
// (#188). synctest makes the order certain. Its clock stands still while the
// program runs, because the wait for the exit is a system call. So ctx ends a
// millisecond after the exit, while Run waits 25 ms for a signal.
func TestRunStopsAfterTheExitOnlyForASignal(t *testing.T) {
	cases := []struct {
		name string
		// ctx returns a context that ends a millisecond after Run starts.
		ctx  func(t *testing.T) context.Context
		want os.Signal
	}{
		{"deadline", func(t *testing.T) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			t.Cleanup(cancel)
			return ctx
		}, nil},
		{"cancel", func(t *testing.T) context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			time.AfterFunc(time.Millisecond, cancel)
			return ctx
		}, nil},
		{"signal", func(t *testing.T) context.Context {
			ctx, cancel := context.WithCancelCause(context.Background())
			time.AfterFunc(time.Millisecond, func() { cancel(runner.SignalCause{Signal: os.Interrupt}) })
			return ctx
		}, os.Interrupt},
	}
	for _, limit := range []time.Duration{0, time.Hour} {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s/limit=%v", c.name, limit), func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx := c.ctx(t)
					result, err := runner.Run(ctx, runner.Program{Path: "/bin/sh", Args: []string{"-c", "exit 0"}, Limit: limit})
					if err != nil {
						t.Fatal(err)
					}
					if ctx.Err() == nil {
						t.Fatal("Run returned before ctx ended")
					}
					if result.Exit != 0 || result.Signal != c.want || result.TimedOut {
						t.Fatalf("Run returned %+v, want exit 0 and the signal %v", result, c.want)
					}
				})
			})
		}
	}
}

// TestRunStopsAProgramThatRunsAtTheDeadline proves that a deadline that
// ends ctx before the program exits stops it: Run sends SIGTERM, or SIGKILL
// under a limit, and reports SIGTERM. The program ends only on a signal, so
// the deadline came first.
func TestRunStopsAProgramThatRunsAtTheDeadline(t *testing.T) {
	for limit, killed := range map[time.Duration]syscall.Signal{0: syscall.SIGTERM, time.Hour: syscall.SIGKILL} {
		t.Run(fmt.Sprintf("limit=%v", limit), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			result, err := runner.Run(ctx, runner.Program{Path: "/bin/sh", Args: []string{"-c", "exec sleep 30"}, Limit: limit})
			if err != nil {
				t.Fatal(err)
			}
			if result.Signal != syscall.SIGTERM || result.Exit != 128+int(killed) {
				t.Fatalf("Run returned %+v, want SIGTERM and exit %d", result, 128+int(killed))
			}
		})
	}
}

// TestRunKillsWhatHoldsAStreamWhenCleanupEnds proves that the drain never
// delays SIGKILL: a descendant that ignores the signal and holds stdout
// open dies once cleanup ends, before the drain limit.
func TestRunKillsWhatHoldsAStreamWhenCleanupEnds(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("only Linux reaches the descendant once the program exits")
	}
	cleanup := runner.Cleanup
	runner.Cleanup = 300 * time.Millisecond
	t.Cleanup(func() { runner.Cleanup = cleanup })
	ready := filepath.Join(t.TempDir(), "ready")
	script := `trap 'exit 0' TERM; (trap '' TERM; exec sleep 30) & touch "$1"; wait`
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		for _, err := os.Stat(ready); err != nil; _, err = os.Stat(ready) {
			time.Sleep(10 * time.Millisecond)
		}
		cancel(runner.SignalCause{Signal: syscall.SIGTERM})
	}()
	start := time.Now()
	result, err := runner.Run(ctx, runner.Program{Path: "/bin/sh", Args: []string{"-c", script, "sh", ready}, Stdout: io.Discard})
	if err != nil || result.Signal != syscall.SIGTERM {
		t.Fatalf("Run returned %+v and %v", result, err)
	}
	if took := time.Since(start); took > 1500*time.Millisecond {
		t.Fatalf("Run took %v, so SIGKILL waited for the drain limit", took)
	}
}
