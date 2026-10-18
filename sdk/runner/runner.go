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
	// in its own process group.
	Terminal bool
	// Limit, when set, bounds the program and the reading of its output, as
	// completion bounds a list command. At the limit, or when ctx ends
	// first, Run sends SIGKILL to every descendant of ncly at once, without
	// the cleanup of a signal, and closes the streams instead of starting
	// the drain limit.
	Limit time.Duration
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: !p.Terminal}
	if ctx.Err() != nil {
		return Result{Exit: -1, Signal: SignalOf(ctx)}, nil
	}
	copyStdin, err := feed(cmd, p.Stdin)
	var streams []*stream
	if err == nil {
		cmd.Stdout, err = attach(p.Stdout, &streams)
	}
	if err == nil {
		cmd.Stderr, err = attach(p.Stderr, &streams)
	}
	if err == nil {
		// A descendant whose parent exits becomes a child of ncly, so
		// cleanup still reaches it. Only Linux offers this.
		_ = platform.BecomeSubreaper()
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
		result.TimedOut = true
		stopped = ctx.Err() != nil
	}
	if p.Limit > 0 && (result.TimedOut || stopped) {
		// A program that has a limit gets no cleanup: at the limit or on a
		// signal, it ends at once, and so does every descendant that holds
		// its streams.
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
	// answer once the limit has passed. A program with a limit gets no drain
	// limit: its own limit bounds the reading, and then ends what still runs.
	if p.Limit > 0 && result.Signal == nil {
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

// feed connects r to what the program reads: a file as it is, and
// otherwise a pipe, and returns the function that starts filling that pipe
// from r once the program has started. Run never waits for that copy,
// because a Read from r may block until its owner closes r, as the
// documentation of exec.Cmd warns of its own copy. Wait closes the pipe
// once the program exits.
func feed(cmd *exec.Cmd, r io.Reader) (func(), error) {
	if _, isFile := r.(*os.File); isFile || r == nil {
		cmd.Stdin = r
		return func() {}, nil
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	return func() {
		go func() {
			_, _ = io.Copy(w, r)
			_ = w.Close()
		}()
	}, nil
}

// stream copies what a program writes on a pipe that the runner owns, so
// that the runner, not a descendant that holds the pipe, decides when the
// copy ends.
type stream struct {
	to   io.Writer
	r, w *os.File
	done chan struct{}
}

// attach returns what the program writes to: a file as it is, so that the
// program sees a terminal when it is one, and otherwise the write end of a
// pipe whose copy to w the runner controls. A writer that wraps a file, such
// as the stdout of an SDK handler, passes its file. Two streams to the same
// writer share one pipe, as exec.Cmd does, so their writes never overlap.
func attach(w io.Writer, streams *[]*stream) (io.Writer, error) {
	if wrapper, wraps := w.(interface{ File() *os.File }); wraps {
		w = wrapper.File()
	}
	if _, isFile := w.(*os.File); isFile || w == nil {
		return w, nil
	}
	for _, s := range *streams {
		if sameWriter(s.to, w) {
			return s.w, nil
		}
	}
	r, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	s := &stream{to: w, r: r, w: pw, done: make(chan struct{})}
	go func() {
		_, _ = io.Copy(w, r)
		close(s.done)
	}()
	*streams = append(*streams, s)
	return pw, nil
}

// sameWriter reports whether a and b are the same writer. Comparing two
// values of a type that is not comparable panics, and then they differ.
func sameWriter(a, b io.Writer) (same bool) {
	defer func() { _ = recover() }()
	return a == b
}

// drain lets each copy run until its stream ends or until the deadline,
// then stops it. Output written after the deadline is lost.
func drain(streams []*stream, deadline time.Time) {
	for _, s := range streams {
		select {
		case <-s.done:
		case <-time.After(time.Until(deadline)):
		}
		_ = s.r.Close()
		<-s.done
	}
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

// SignalCause is the cause of a context that a signal ended.
type SignalCause struct{ Signal os.Signal }

func (c SignalCause) Error() string { return "stopped by " + c.Signal.String() }

// SignalOf returns the signal that ended ctx, and SIGTERM when ctx ended
// for another reason, such as a timeout.
func SignalOf(ctx context.Context) os.Signal {
	if sig, ok := signalCause(ctx); ok {
		return sig
	}
	return syscall.SIGTERM
}

// signalCause returns the signal that ended ctx, and false when a signal
// did not end it.
func signalCause(ctx context.Context) (os.Signal, bool) {
	var cause SignalCause
	ok := errors.As(context.Cause(ctx), &cause)
	return cause.Signal, ok
}

// OnSignal returns a context that the first Ctrl-C or SIGTERM ends, with
// that signal as its cause. The handler stays registered for the life of
// the process, so a later signal never ends it.
func OnSignal() context.Context {
	ctx, _ := ListenForSignals()
	return ctx
}

// ListenForSignals returns a context that the first Ctrl-C or SIGTERM ends,
// with that signal as its cause, as OnSignal does, and a function that stops
// listening, after which a signal reaches the process as before, such as
// once a form has its answer.
func ListenForSignals() (context.Context, func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		select {
		case s := <-signals:
			cancel(SignalCause{s})
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(signals)
		cancel(nil)
	}
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
