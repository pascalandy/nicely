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
	"sync/atomic"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
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
// exited, only a signal that ends ctx during the settle wait stops it, as a
// Ctrl-C that reaches the program and ncly together does. A deadline or a
// plain cancellation that ends ctx then keeps the program's own result
// (#188). synctest makes the order certain: its clock stands still while the
// program runs, because the wait for the exit is a system call, so ctx ends
// a millisecond after the exit, while Run waits 25 ms for a signal.
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
// ends ctx before the program exits stops it with SIGTERM, or with SIGKILL
// under a limit. The program ends only on a signal, so the deadline came
// first.
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

// TestRunNeverWaitsForAStdinThatBlocks proves that Run returns once the
// program exits, although its stdin, a reader that is not a file, still
// blocks in Read, as an open pipe does, and that Run keeps no file open for
// it.
func TestRunNeverWaitsForAStdinThatBlocks(t *testing.T) {
	exit := runner.Program{Path: "/bin/sh", Args: []string{"-c", "exit 0"}}
	// The first pipe of the process opens the files of the Go poller.
	if _, err := runner.Run(context.Background(), exit); err != nil {
		t.Fatal(err)
	}
	before := openFiles(t)
	r, w := io.Pipe()
	defer func() { _ = w.Close() }()
	exit.Stdin = r
	done := make(chan error, 1)
	go func() {
		_, err := runner.Run(context.Background(), exit)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run still waits for the copy of stdin after the program exited")
	}
	if after := openFiles(t); after != before {
		t.Fatalf("%d files are open after Run, want %d", after, before)
	}
}

// TestRunReadsNoStdinWhenTheStartFails proves that a program that never
// started leaves nothing that reads its stdin.
func TestRunReadsNoStdinWhenTheStartFails(t *testing.T) {
	r, w := io.Pipe()
	defer func() { _ = r.Close() }()
	if _, err := runner.Run(context.Background(), runner.Program{Path: "/nonexistent/program", Stdin: r}); err == nil {
		t.Fatal("Run started a program that does not exist")
	}
	read := make(chan struct{})
	go func() {
		_, _ = w.Write([]byte("x"))
		close(read)
	}()
	select {
	case <-read:
		t.Fatal("a copy still reads the stdin of a program that never started")
	case <-time.After(100 * time.Millisecond):
	}
}

func openFiles(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
}

// TestRunCopiesStdoutAndStderrToOneWriterOneWriteAtATime proves that a
// writer that receives both streams never sees two writes at once.
func TestRunCopiesStdoutAndStderrToOneWriterOneWriteAtATime(t *testing.T) {
	var w overlapWriter
	script := `i=0; while [ $i -lt 50 ]; do echo out; echo err >&2; i=$((i+1)); done`
	result, err := runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", script}, Stdout: &w, Stderr: &w})
	if err != nil || result.Exit != 0 {
		t.Fatalf("Run returned %+v and %v", result, err)
	}
	if w.overlap.Load() {
		t.Fatal("two writes reached the writer at once")
	}
	if got := w.bytes.Load(); got != 50*8 {
		t.Fatalf("the writer got %d bytes, want %d", got, 50*8)
	}
}

// overlapWriter records whether a Write starts while another runs.
type overlapWriter struct {
	busy, overlap atomic.Bool
	bytes         atomic.Int64
}

func (w *overlapWriter) Write(b []byte) (int, error) {
	if w.busy.Swap(true) {
		w.overlap.Store(true)
	}
	time.Sleep(time.Millisecond)
	w.bytes.Add(int64(len(b)))
	w.busy.Store(false)
	return len(b), nil
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

// TestAProgramWritesToTheFileThatAnOutputWraps proves that a program started
// with the stdout of an SDK handler writes to the file itself, never to a
// pipe, so it sees a terminal when the file is one.
func TestAProgramWritesToTheFileThatAnOutputWraps(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	result, err := runner.Run(context.Background(), runner.Program{
		Path:   "/bin/sh",
		Args:   []string{"-c", "test ! -p /dev/stdout"},
		Stdout: command.NewOutput(f),
	})
	if err != nil || result.Exit != 0 {
		t.Errorf("the program wrote to a pipe: exit %d, error %v", result.Exit, err)
	}
}
