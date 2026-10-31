package extensiontest

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/rogpeppe/go-internal/testscript"
)

func (c *Coverage) signal(ts *testscript.TestScript, neg bool, args, survives, protected []string, registeredHost string) {
	if neg || len(args) < 3 || (args[0] != "INT" && args[0] != "TERM") {
		ts.Fatalf("usage: signal INT|TERM <marker> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, _ := c.pairedCall(ts, args[2], args[3:], registeredHost)
	marker := ts.MkAbs(args[1])
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		ts.Fatalf("signal: marker %s must not exist before the call", args[1])
	}
	before, err := snapshotFiles(protected, false)
	ts.Check(err)
	sig, wantExit, wantCode := syscall.SIGINT, 130, contract.Interrupted
	if args[0] == "TERM" {
		sig, wantExit, wantCode = syscall.SIGTERM, 143, contract.Terminated
	}
	exit, stdout, stderr := foreground(ts, "signal", program, args[3:], func(done <-chan struct{}, host *os.Process) error {
		return sendSignal(done, host, marker, survives, sig)
	})
	code, problem := c.checkedCode(words, exit, stdout, stderr)
	if exit != wantExit {
		ts.Fatalf("signal: %s exited with %d, want %d", args[2], exit, wantExit)
	}
	if problem != "" {
		ts.Fatalf("signal: %s", problem)
	}
	if code != wantCode {
		ts.Fatalf("signal: %s returned %s, want %s", args[0], code, wantCode)
	}
	same, err := unchanged(protected, before, marker)
	ts.Check(err)
	if !same {
		ts.Fatalf("signal: %s changed protected paths: %s", args[0], fmt.Sprint(protected))
	}
}

// foreground runs a resolved call in its own process group, with the scenario's stdin, environment, and folder.
// act runs beside the host until it returns. foreground returns the exit and both streams of the host.
func foreground(ts *testscript.TestScript, name, program string, args []string, act func(done <-chan struct{}, host *os.Process) error) (int, string, string) {
	dir, err := os.MkdirTemp("", "ncly-"+name+"-")
	ts.Check(err)
	defer func() { _ = os.RemoveAll(dir) }()
	// Testscript exposes neither stdin nor the complete scenario environment to custom commands.
	// Capture both before starting a command whose process group and stream drain the kit controls.
	envfile, inputfile := filepath.Join(dir, "env"), filepath.Join(dir, "stdin")
	ts.Check(ts.Exec("/bin/sh", "-c", `/usr/bin/env -0 > "$1" && /bin/cat > "$2"`, name, envfile, inputfile))
	env, err := os.ReadFile(envfile)
	ts.Check(err)
	input, err := os.Open(inputfile)
	ts.Check(err)
	defer func() { _ = input.Close() }()
	cmd := exec.Command(program, args...)
	cmd.Dir = ts.MkAbs(".")
	cmd.Env = strings.Split(strings.TrimSuffix(string(env), "\x00"), "\x00")
	cmd.Stdin = input
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = runner.DrainLimit
	ts.Check(cmd.Start())
	defer func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }()
	done, result := make(chan struct{}), make(chan error, 1)
	go func() { result <- act(done, cmd.Process) }()
	exit, runErr := exitCode(cmd.Wait())
	close(done)
	_, _ = ts.Stdout().Write(stdout.Bytes())
	_, _ = ts.Stderr().Write(stderr.Bytes())
	ts.Check(<-result)
	ts.Check(runErr)
	return exit, stdout.String(), stderr.String()
}

// waitMarker waits up to 10 seconds for the host's new regular marker file.
func waitMarker(name string, done <-chan struct{}, host *os.Process, marker string) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if info, err := os.Lstat(marker); err == nil && info.Mode().IsRegular() {
			return nil
		}
		select {
		case <-done:
			return fmt.Errorf("%s: host exited before marker %s", name, marker)
		case <-deadline.C:
			kill(host)
			return fmt.Errorf("%s: timed out waiting for marker %s", name, marker)
		case <-tick.C:
		}
	}
}

// waitExit waits up to limit for the host to exit, and kills its process group after that.
func waitExit(done <-chan struct{}, host *os.Process, limit time.Duration) bool {
	deadline := time.NewTimer(limit)
	defer deadline.Stop()
	select {
	case <-done:
		return true
	case <-deadline.C:
		kill(host)
		return false
	}
}

func kill(host *os.Process) { _ = syscall.Kill(-host.Pid, syscall.SIGKILL) }

// unchanged compares protected paths with their state before the call. The kit's own files are skipped.
func unchanged(protected []string, before map[string]fileState, skip ...string) (bool, error) {
	after, err := snapshotFiles(protected, false)
	if err != nil {
		return false, err
	}
	for _, path := range skip {
		delete(before, path)
		delete(after, path)
	}
	return maps.Equal(before, after), nil
}

// sendSignal waits for the host's new marker, captures finished files, and bounds both waits.
func sendSignal(done <-chan struct{}, host *os.Process, marker string, survives []string, sig syscall.Signal) error {
	if err := waitMarker("signal", done, host, marker); err != nil {
		return err
	}
	before, err := snapshotFiles(survives, true)
	if err != nil {
		kill(host)
		return err
	}
	if err := host.Signal(sig); err != nil {
		kill(host)
		return fmt.Errorf("signal: send %s to host: %w", sig, err)
	}
	if !waitExit(done, host, runner.Cleanup+runner.DrainLimit+3*time.Second) {
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
