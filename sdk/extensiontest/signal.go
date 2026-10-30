package extensiontest

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/rogpeppe/go-internal/testscript"
)

func (c *Coverage) signal(ts *testscript.TestScript, neg bool, args, survives []string) {
	if neg || len(args) < 3 || (args[0] != "INT" && args[0] != "TERM") {
		ts.Fatalf("usage: signal INT|TERM <marker> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, _ := c.pairedCall(ts, args[2], args[3:])
	marker := ts.MkAbs(args[1])
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		ts.Fatalf("signal: marker %s must not exist before the call", args[1])
	}
	dir, err := os.MkdirTemp("", "ncly-signal-")
	ts.Check(err)
	defer func() { _ = os.RemoveAll(dir) }()
	pidfile := filepath.Join(dir, "host.pid")
	sig, wantExit, wantCode := syscall.SIGINT, 130, contract.Interrupted
	if args[0] == "TERM" {
		sig, wantExit, wantCode = syscall.SIGTERM, 143, contract.Terminated
	}
	done, result := make(chan struct{}), make(chan error, 1)
	go func() { result <- sendSignal(done, pidfile, marker, survives, sig) }()
	// The shell immediately becomes the host. Its PID and exit are the host's, and no cleanup command replaces either.
	arguments := append([]string{"-c", `echo "$$" > "$1"; shift; exec "$@"`, "signal", pidfile, program}, args[3:]...)
	exit, runErr := exitCode(ts.Exec("/bin/sh", arguments...))
	close(done)
	ts.Check(<-result)
	ts.Check(runErr)
	code, problem := c.checkedCode(ts, words, exit)
	if exit != wantExit {
		ts.Fatalf("signal: %s exited with %d, want %d", args[2], exit, wantExit)
	}
	if problem != "" {
		ts.Fatalf("signal: %s", problem)
	}
	if code != wantCode {
		ts.Fatalf("signal: %s returned %s, want %s", args[0], code, wantCode)
	}
}

// sendSignal waits for the host's new marker, captures finished files, and bounds both waits.
func sendSignal(done <-chan struct{}, pidfile, marker string, survives []string, sig syscall.Signal) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	var host *os.Process
	defer func() {
		if host != nil {
			_ = host.Release()
		}
	}()
	for {
		if host == nil {
			if data, err := os.ReadFile(pidfile); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
					host, _ = os.FindProcess(pid)
				}
			}
		}
		if info, err := os.Lstat(marker); err == nil && info.Mode().IsRegular() && host != nil {
			break
		}
		select {
		case <-done:
			return fmt.Errorf("signal: host exited before marker %s", marker)
		case <-deadline.C:
			if host != nil {
				_ = host.Kill()
			}
			return fmt.Errorf("signal: timed out waiting for marker %s", marker)
		case <-tick.C:
		}
	}
	before, err := snapshotFiles(survives, true)
	if err != nil {
		_ = host.Kill()
		return err
	}
	if err := host.Signal(sig); err != nil {
		_ = host.Kill()
		return fmt.Errorf("signal: send %s to host: %w", sig, err)
	}
	deadline.Reset(15 * time.Second)
	select {
	case <-done:
	case <-deadline.C:
		_ = host.Kill()
		return fmt.Errorf("signal: host did not exit after %s", sig)
	}
	after, err := snapshotFiles(survives, true)
	if err != nil {
		return err
	}
	if !maps.Equal(before, after) {
		return fmt.Errorf("signal: finished files changed after %s: %v", sig, survives)
	}
	return nil
}
