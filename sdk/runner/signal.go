package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

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
