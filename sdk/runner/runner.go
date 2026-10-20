// Package runner runs the programs that ncly and its extensions start, as
// the Programs that ncly runs section of contract.md defines. It owns their
// environment, their session or process group, their cancellation, their
// captured output, and their exit code.
//
// Start a program without the terminal with Setsid alone, never with
// Setpgid too, because setpgid refuses a session leader (D047). Give such a
// program no descriptor of the terminal, because a process outside the
// terminal's session reads it without a stop. Bound what a caller keeps of
// noisy output with Capture.
//
// Read a program's stdout and stderr through pipes that the runner owns: a
// writer that exec.Cmd copies can keep Wait blocked until descendants close
// the pipe. Send cleanup signals through the runner to the whole process
// tree, because exec.CommandContext kills only the direct child. Keep the
// signal handler registered until cleanup ends, because stopping
// signal.NotifyContext first lets a second Ctrl-C end ncly at once. The test
// binary shortens the 15 seconds of cleanup, so a scenario proves the
// SIGKILL path without waiting.
package runner

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/pascalandy/nicely/sdk/platform"
)

// Program is one program to run.
type Program struct {
	Path string
	// Name, when set, is the program's argv[0] instead of Path, as ncly
	// starts itself under the name of a bundled extension.
	Name string
	Args []string
	Env  []string
	// Stdin, Stdout, and Stderr reach the program. An *os.File passes as it
	// is, so the program sees the terminal when it is one.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Terminal keeps the program in the process group of ncly, because it
	// may read the terminal, as an extension does. Any other program starts
	// in a new session, with no controlling terminal, so an open of /dev/tty
	// fails at once. Such a program also gets no descriptor of the terminal.
	// A terminal on its stdin fails Run. A terminal on its stdout or stderr
	// reaches it through a pipe that the runner reads. No program inherits a
	// descriptor of the terminal above stderr.
	Terminal bool
	// Limit, when set, bounds execution. If it expires before the program
	// exits, Run sends SIGKILL to every descendant of ncly and closes the
	// streams at once. Normal exit and cancellation keep their ordinary
	// drain and cleanup, even when the limit expires during them.
	Limit time.Duration
	// Completion also bounds output reading by Limit and skips cleanup on
	// cancellation, as a list command that completion runs requires.
	Completion bool
}

// Result is how a program ended.
type Result struct {
	// Exit is the program's exit code, or 128 plus the number of the signal
	// that ended it, as a shell reports it. It is -1 when ctx had ended
	// before the program started, so it never ran.
	Exit int
	// Signal is the signal that ended ctx before the program exited, or
	// SIGTERM when another cause, such as a deadline, ended it. A signal that
	// ends ctx within signalSettle after the exit counts too, because it may
	// have reached the program first. Signal is nil otherwise.
	Signal os.Signal
	// TimedOut reports that the program still ran at its Limit.
	TimedOut bool
}

// Cleanup bounds the time that a program and its descendants get to stop
// after a signal, before SIGKILL ends what still runs. Only the test binary
// shortens it.
var Cleanup = 15 * time.Second

// DrainLimit bounds how long the runner reads a program's streams after the
// program exits, because a descendant may hold them open. It runs on its
// own clock, apart from Cleanup.
const DrainLimit = 2 * time.Second

// Run starts the program and waits for it to exit. When ctx ends first, Run
// sends the signal that ended it, or SIGTERM for any other end, to every
// descendant of ncly that it can reach, the program included, and sends
// SIGKILL to what still runs once Cleanup has passed. A signal that ends
// ctx within signalSettle after the exit does the same. Any other later end
// of ctx, such as a deadline, changes nothing. Run reads a stream that is
// not a file for at most DrainLimit after the program exits.
func Run(ctx context.Context, p Program) (Result, error) {
	cmd := exec.Command(p.Path, p.Args...)
	if p.Name != "" {
		cmd.Args[0] = p.Name
	}
	cmd.Env = p.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: !p.Terminal}
	if ctx.Err() != nil {
		return Result{Exit: -1, Signal: SignalOf(ctx)}, nil
	}
	copyStdin, err := feed(cmd, p.Stdin, p.Terminal)
	var streams []*stream
	if err == nil {
		cmd.Stdout, err = attach(p.Stdout, p.Terminal, &streams)
	}
	if err == nil {
		cmd.Stderr, err = attach(p.Stderr, p.Terminal, &streams)
	}
	if err == nil {
		// A descendant whose parent exits becomes a child of ncly, so
		// cleanup still reaches it. Only Linux offers this.
		_ = platform.BecomeSubreaper()
		closeInheritedTerminals()
		err = cmd.Start()
	}
	for _, s := range streams {
		// Closing the runner's copy lets EOF arrive once the program and
		// every descendant that inherited the write end have closed theirs.
		_ = s.w.Close()
	}
	if err != nil {
		drain(streams, time.Now())
		return Result{}, err
	}
	copyStdin()
	var limit <-chan time.Time
	limitEnds := time.Now().Add(p.Limit)
	if p.Limit > 0 {
		timer := time.NewTimer(p.Limit)
		defer timer.Stop()
		limit = timer.C
	}
	exited := make(chan struct{})
	var exitedAt time.Time
	go func() {
		_ = cmd.Wait()
		exitedAt = time.Now()
		close(exited)
	}()
	var result Result
	// stopped reports that an end of ctx stops the program.
	var stopped bool
	select {
	case <-exited:
		// A Ctrl-C at the terminal reaches the program and ncly together, so
		// the program may exit, whatever its code, before ncly sees its own
		// copy. Any other end of ctx, such as a deadline, came after the
		// exit, and changes nothing (#188).
		select {
		case <-ctx.Done():
			_, stopped = signalCause(ctx)
		case <-time.After(signalSettle):
		}
	case <-ctx.Done():
		stopped = true
	case <-limit:
		stopped = ctx.Err() != nil
		result.TimedOut = !stopped
	}
	if result.TimedOut || p.Completion && stopped {
		// Expiry kills at once. Only completion does the same on a signal.
		if stopped {
			result.Signal = SignalOf(ctx)
		}
		signalTree(syscall.SIGKILL)
		<-exited
		drain(streams, time.Now())
		result.Exit = exitCode(cmd.ProcessState)
		return result, nil
	}
	var cleanupEnds time.Time
	if stopped {
		result.Signal = SignalOf(ctx)
		cleanupEnds = time.Now().Add(Cleanup)
		signalTree(result.Signal)
		select {
		case <-exited:
		case <-time.After(Cleanup):
			signalTree(syscall.SIGKILL)
			<-exited
		}
	}
	var stopping sync.WaitGroup
	if result.Signal != nil {
		stopping.Go(func() { stopDescendants(cleanupEnds) })
	}
	// The drain limit counts from the exit, never from the end of cleanup,
	// so a descendant that outlives the program cannot write into its
	// answer once the limit has passed. Only completion bounds the reading
	// by its execution limit and then ends what still runs.
	if p.Completion && p.Limit > 0 && result.Signal == nil {
		drain(streams, limitEnds)
		signalTree(syscall.SIGKILL)
	} else {
		drain(streams, exitedAt.Add(DrainLimit))
	}
	stopping.Wait()
	result.Exit = exitCode(cmd.ProcessState)
	return result, nil
}

// signalSettle is how long Run waits, once the program has exited, for a
// signal that ncly received at the same time, which the Go runtime delivers
// within a millisecond. A signal that arrives later changes nothing.
const signalSettle = 25 * time.Millisecond

// stopDescendants waits until no descendant of ncly runs, and sends SIGKILL
// to those that still run once cleanup ends.
func stopDescendants(cleanupEnds time.Time) {
	for len(platform.Descendants(os.Getpid())) > 0 {
		if time.Now().After(cleanupEnds) {
			signalTree(syscall.SIGKILL)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Stop sends sig to every descendant of ncly, then waits until none runs,
// and sends SIGKILL to those that still run once cleanup ends, as
// Cancellation in contract.md defines. It reaches a program that starts
// outside Run, such as the security program that go-keyring starts on
// macOS, so a command that a signal ends while it waits for one calls it.
func Stop(sig os.Signal) {
	cleanupEnds := time.Now().Add(Cleanup)
	signalTree(sig)
	stopDescendants(cleanupEnds)
}

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

func exitCode(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}
