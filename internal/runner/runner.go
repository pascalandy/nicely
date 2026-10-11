// Package runner runs the programs that ncly and its extensions start, as
// the Programs that ncly runs section of contract.md defines: their
// environment, their process group, their cancellation, and their exit
// code.
package runner

import (
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/internal/platform"
)

// Program is one program to run.
type Program struct {
	Path string
	Args []string
	Env  []string
	// Stdin, Stdout, and Stderr reach the program. An *os.File passes as it
	// is, so the program sees the terminal when it is one.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Terminal keeps the program in the process group of ncly, because it
	// may read the terminal, as an extension does. Any other program starts
	// in its own process group.
	Terminal bool
}

// Result is how a program ended.
type Result struct {
	// Exit is the program's exit code, or 128 plus the number of the signal
	// that ended it, as a shell reports it. It is -1 when ctx had ended
	// before the program started, so it never ran.
	Exit int
	// Signal is the signal that ended ctx before the program exited, or nil
	// when the program exited first.
	Signal os.Signal
}

// Run starts the program and waits for it to exit. When ctx ends first, Run
// sends the signal that ended it, or SIGTERM for any other end, to every
// descendant of ncly that it can reach, the program included, then keeps
// waiting. A later end of ctx changes nothing.
func Run(ctx context.Context, p Program) (Result, error) {
	cmd := exec.Command(p.Path, p.Args...)
	cmd.Env = p.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = p.Stdin, p.Stdout, p.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: !p.Terminal}
	// A descendant whose parent exits becomes a child of ncly, so cleanup
	// still reaches it. Only Linux offers this.
	_ = platform.BecomeSubreaper()
	if ctx.Err() != nil {
		return Result{Exit: -1, Signal: SignalOf(ctx)}, nil
	}
	if err := cmd.Start(); err != nil {
		return Result{}, err
	}
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	var result Result
	select {
	case <-exited:
		// A Ctrl-C at the terminal reaches the program and ncly together, so
		// the program may exit, whatever its code, before ncly sees its own
		// copy.
		select {
		case <-ctx.Done():
		case <-time.After(signalSettle):
		}
	case <-ctx.Done():
	}
	if ctx.Err() != nil {
		result.Signal = SignalOf(ctx)
		signalTree(result.Signal)
		<-exited
	}
	result.Exit = exitCode(cmd.ProcessState)
	return result, nil
}

// signalSettle is how long Run waits, once the program has exited, for a
// signal that ncly received at the same time, which the Go runtime delivers
// within a millisecond. A signal that arrives later changes nothing.
const signalSettle = 25 * time.Millisecond

// signalTree sends sig to every descendant of ncly.
func signalTree(sig os.Signal) {
	for _, pid := range platform.Descendants(os.Getpid()) {
		if process, err := os.FindProcess(pid); err == nil {
			_ = process.Signal(sig)
			// On Linux, FindProcess opens a pidfd that only Release or the
			// garbage collector closes.
			_ = process.Release()
		}
	}
}

// SignalCause is the cause of a context that a signal ended.
type SignalCause struct{ Signal os.Signal }

func (c SignalCause) Error() string { return "stopped by " + c.Signal.String() }

// SignalOf returns the signal that ended ctx, and SIGTERM when ctx ended
// for another reason, such as a timeout.
func SignalOf(ctx context.Context) os.Signal {
	var cause SignalCause
	if errors.As(context.Cause(ctx), &cause) {
		return cause.Signal
	}
	return syscall.SIGTERM
}

// OnSignal returns a context that the first Ctrl-C or SIGTERM ends, with
// that signal as its cause. The handler stays registered for the life of
// the process, so a later signal never ends it.
func OnSignal() context.Context {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { cancel(SignalCause{<-signals}) }()
	return ctx
}

func exitCode(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}

// Environ returns the environment of a program: its parent's, without the
// key of any service, which needs a grant, and with each variable of set
// given its value, or removed when that value is empty.
func Environ(parent []string, set map[string]string) []string {
	env := slices.DeleteFunc(slices.Clone(parent), func(entry string) bool {
		name, _, _ := strings.Cut(entry, "=")
		_, changed := set[name]
		return changed || strings.HasSuffix(name, "_API_KEY")
	})
	for _, name := range slices.Sorted(maps.Keys(set)) {
		if value := set[name]; value != "" {
			env = append(env, name+"="+value)
		}
	}
	return env
}
