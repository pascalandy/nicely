package runner_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/runner"
)

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
