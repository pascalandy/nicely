package runner_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/pascalandy/nicely/sdk/runner"
	"golang.org/x/sys/unix"
)

// TestRunStartsAProgramWithoutTheTerminalInANewSession proves that a
// program without the terminal leads a new session. A program with the
// terminal stays in the session of ncly (D047).
func TestRunStartsAProgramWithoutTheTerminalInANewSession(t *testing.T) {
	own, err := unix.Getsid(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name     string
		terminal bool
	}{{"without", false}, {"with", true}} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			ready, done := filepath.Join(dir, "pid"), filepath.Join(dir, "done")
			script := `echo $$ > "$1.tmp"; mv "$1.tmp" "$1"; while [ ! -f "$2" ]; do sleep 0.01; done`
			ended := make(chan error, 1)
			go func() {
				_, err := runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", script, "sh", ready, done}, Terminal: c.terminal})
				ended <- err
			}()
			pid := waitForPID(t, ready)
			sid, err := unix.Getsid(pid)
			if err := os.WriteFile(done, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := <-ended; err != nil {
				t.Fatal(err)
			}
			if want := map[bool]int{false: pid, true: own}[c.terminal]; sid != want {
				t.Fatalf("the program's session is %d, want %d", sid, want)
			}
		})
	}
}

// TestRunGivesAProgramWithoutTheTerminalNoDescriptorOfIt proves that a
// terminal on stdout reaches a program without the terminal through a pipe.
// The copy of that pipe still shows the output there. A terminal on stdin
// fails Run. A program with the terminal gets the terminal itself.
func TestRunGivesAProgramWithoutTheTerminalNoDescriptorOfIt(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ptmx.Close() }()
	defer func() { _ = tty.Close() }()
	shown := make(chan string, 1)
	go func() {
		var screen bytes.Buffer
		buf := make([]byte, 256)
		for !strings.Contains(screen.String(), "terminal: yes") {
			n, err := ptmx.Read(buf)
			screen.Write(buf[:n])
			if err != nil {
				break
			}
		}
		shown <- screen.String()
	}()
	script := `if [ -t 1 ]; then echo "$1: yes"; else echo "$1: no"; fi`
	for _, terminal := range []bool{false, true} {
		label := map[bool]string{false: "pipe", true: "terminal"}[terminal]
		_, err := runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", script, "sh", label}, Stdout: tty, Terminal: terminal})
		if err != nil {
			t.Fatal(err)
		}
	}
	select {
	case screen := <-shown:
		if !strings.Contains(screen, "pipe: no") || !strings.Contains(screen, "terminal: yes") {
			t.Fatalf("the terminal showed %q, want pipe: no and terminal: yes", screen)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the terminal never showed both lines")
	}
	started := filepath.Join(t.TempDir(), "started")
	if _, err := runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", "touch " + started}, Stdin: tty}); err == nil {
		t.Fatal("Run gave a program without the terminal the terminal on stdin")
	}
	if _, err := os.Stat(started); err == nil {
		t.Fatal("the program started")
	}
}

// TestCaptureKeepsOnlyTheFirstBytes proves that Capture keeps at most its
// limit and reports what it dropped. It never fails a write, so a program
// that writes more than the limit still runs to its end.
func TestCaptureKeepsOnlyTheFirstBytes(t *testing.T) {
	c := runner.Capture{Limit: 1000}
	result, err := runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", "i=0; while [ $i -lt 500 ]; do echo 123456789; i=$((i+1)); done; echo end >&2"}, Stdout: &c, Stderr: io.Discard})
	if err != nil || result.Exit != 0 {
		t.Fatalf("Run returned %+v and %v", result, err)
	}
	if len(c.Bytes()) != 1000 || !c.Over() || !bytes.HasPrefix(c.Bytes(), []byte("123456789\n")) {
		t.Fatalf("Capture kept %d bytes, over %t", len(c.Bytes()), c.Over())
	}
	small := runner.Capture{Limit: 10}
	if n, err := small.Write([]byte("0123456789")); n != 10 || err != nil || small.Over() {
		t.Fatalf("a write that fits returned %d and %v, over %t", n, err, small.Over())
	}
	if n, err := small.Write([]byte("x")); n != 1 || err != nil || !small.Over() || string(small.Bytes()) != "0123456789" {
		t.Fatalf("a write past the limit returned %d and %v, kept %q", n, err, small.Bytes())
	}
}

// waitForPID returns the process ID that a program writes to path.
func waitForPID(t *testing.T, path string) int {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if data, err := os.ReadFile(path); err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				t.Fatal(err)
			}
			return pid
		}
	}
	t.Fatalf("%s never appeared", path)
	return 0
}

// TestRunPassesNoInheritedTerminal proves that no program gets a descriptor
// of the terminal that ncly inherited above stderr without close-on-exec. A
// shell can leave one open for its children.
func TestRunPassesNoInheritedTerminal(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	// The master stays open: once it closes, the slave is no terminal.
	defer func() { _ = ptmx.Close() }()
	defer func() { _ = tty.Close() }()
	script := `for path in /dev/fd/*; do fd=${path##*/}; if [ "$fd" -gt 2 ] && [ -t "$fd" ]; then echo "terminal on $fd"; fi; done`
	for _, terminal := range []bool{false, true} {
		// Dup gives a descriptor without close-on-exec, as a shell passes
		// one. Each case gets its own, because Run marks the one it finds.
		fd, err := unix.Dup(int(tty.Fd()))
		if err != nil {
			t.Fatal(err)
		}
		found := runner.Capture{Limit: 1000}
		_, err = runner.Run(context.Background(), runner.Program{Path: "/bin/sh", Args: []string{"-c", script}, Stdout: &found, Terminal: terminal})
		_ = unix.Close(fd)
		if err != nil {
			t.Fatal(err)
		}
		if len(found.Bytes()) > 0 {
			t.Fatalf("a program with Terminal %t inherited the terminal: %s", terminal, found.Bytes())
		}
	}
}

// TestRunNeverWaitsOnATerminalThatNobodyReads proves that the copy to a
// terminal ends Run at its limit, as a direct write did. A full terminal,
// such as one that nobody reads, blocks the copy of a program without it.
func TestRunNeverWaitsOnATerminalThatNobodyReads(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tty.Close() }()
	ended := make(chan time.Duration, 1)
	start := time.Now()
	go func() {
		_, _ = runner.Run(context.Background(), runner.Program{Path: "/usr/bin/yes", Stdout: tty, Limit: 100 * time.Millisecond})
		ended <- time.Since(start)
	}()
	select {
	case took := <-ended:
		if took > runner.DrainLimit+time.Second {
			t.Fatalf("Run took %v, more than the drain limit", took)
		}
	case <-time.After(runner.DrainLimit + 3*time.Second):
		// Reading the terminal lets the blocked copy end the run.
		go func() { _, _ = io.Copy(io.Discard, ptmx) }()
		<-ended
		t.Fatal("Run waited until the terminal was read")
	}
	_ = ptmx.Close()
}
