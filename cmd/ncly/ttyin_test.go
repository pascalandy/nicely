package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

// ttyinRun gives only stdin a controlling terminal: ttyin <file> <command> [args...].
// Stdout and stderr keep the pipes that testscript captures.
//
// testscript's builtin opens a nonblocking master on macOS before the slave exists.
// Its reader stops on EAGAIN, so pending echo can block the child's exit (#162).
// creack/pty opens a blocking master, as the full-terminal helper uses.
func ttyinRun() {
	os.Exit(runTTYin(os.Args[1:]))
}

func runTTYin(args []string) int {
	fail := func(err error) int {
		fmt.Fprintln(os.Stderr, "ttyin:", err)
		return 1
	}
	if len(args) < 2 {
		return fail(errors.New("usage: ttyin <file> <command> [args...]"))
	}
	input, err := os.ReadFile(args[0])
	if err != nil {
		return fail(err)
	}
	ptmx, tty, err := pty.Open()
	if err != nil {
		return fail(err)
	}
	defer func() { _ = ptmx.Close() }()
	defer func() { _ = tty.Close() }()
	drained := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, ptmx)
		close(drained)
	}()
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	_ = tty.Close()
	written := make(chan struct{})
	go func() {
		// A command that rejects terminal input can exit before this write.
		_, _ = ptmx.Write(append(input, 4 /* EOT */))
		close(written)
	}()
	code, err := processExitCode(cmd.Wait())
	_ = ptmx.Close()
	<-written
	<-drained
	if err != nil {
		return fail(err)
	}
	return code
}

// processExitCode preserves a command's exit code, including 128 plus its signal.
func processExitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		return 0, err
	}
	if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal()), nil
	}
	return exit.ExitCode(), nil
}
