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

func (c *Coverage) signal(ts *testscript.TestScript, neg bool, args, survives []string, registeredHost string) {
	if neg || len(args) < 3 || (args[0] != "INT" && args[0] != "TERM") {
		ts.Fatalf("usage: signal INT|TERM <marker> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, _ := c.pairedCall(ts, args[2], args[3:], registeredHost)
	marker := ts.MkAbs(args[1])
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		ts.Fatalf("signal: marker %s must not exist before the call", args[1])
	}
	dir, err := os.MkdirTemp("", "ncly-signal-")
	ts.Check(err)
	defer func() { _ = os.RemoveAll(dir) }()
	// Testscript exposes neither stdin nor the complete scenario environment to custom commands.
	// Capture both before starting a command whose process group and stream drain the kit controls.
	envfile, inputfile := filepath.Join(dir, "env"), filepath.Join(dir, "stdin")
	ts.Check(ts.Exec("/bin/sh", "-c", `/usr/bin/env -0 > "$1" && /bin/cat > "$2"`, "signal", envfile, inputfile))
	env, err := os.ReadFile(envfile)
	ts.Check(err)
	input, err := os.Open(inputfile)
	ts.Check(err)
	defer func() { _ = input.Close() }()
	cmd := exec.Command(program, args[3:]...)
	cmd.Dir = ts.MkAbs(".")
	cmd.Env = strings.Split(strings.TrimSuffix(string(env), "\x00"), "\x00")
	cmd.Stdin = input
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = runner.DrainLimit
	ts.Check(cmd.Start())
	defer func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }()
	sig, wantExit, wantCode := syscall.SIGINT, 130, contract.Interrupted
	if args[0] == "TERM" {
		sig, wantExit, wantCode = syscall.SIGTERM, 143, contract.Terminated
	}
	done, result := make(chan struct{}), make(chan error, 1)
	go func() { result <- sendSignal(done, cmd.Process, marker, survives, sig) }()
	exit, runErr := exitCode(cmd.Wait())
	close(done)
	_, _ = ts.Stdout().Write(stdout.Bytes())
	_, _ = ts.Stderr().Write(stderr.Bytes())
	ts.Check(<-result)
	ts.Check(runErr)
	code, problem := c.checkedCode(words, exit, stdout.String(), stderr.String())
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
func sendSignal(done <-chan struct{}, host *os.Process, marker string, survives []string, sig syscall.Signal) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	kill := func() { _ = syscall.Kill(-host.Pid, syscall.SIGKILL) }
	for {
		if info, err := os.Lstat(marker); err == nil && info.Mode().IsRegular() {
			break
		}
		select {
		case <-done:
			return fmt.Errorf("signal: host exited before marker %s", marker)
		case <-deadline.C:
			kill()
			return fmt.Errorf("signal: timed out waiting for marker %s", marker)
		case <-tick.C:
		}
	}
	before, err := snapshotFiles(survives, true)
	if err != nil {
		kill()
		return err
	}
	if err := host.Signal(sig); err != nil {
		kill()
		return fmt.Errorf("signal: send %s to host: %w", sig, err)
	}
	deadline.Reset(runner.Cleanup + runner.DrainLimit + 3*time.Second)
	select {
	case <-done:
	case <-deadline.C:
		kill()
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
