package extension

import (
	"io"
	"os"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"golang.org/x/term"
)

// workersMargin is how long before the end of cleanup the workers of a
// signaled run go, at the latest.
const workersMargin = time.Second

// handle runs the handler of a validated call and prints its result by the
// kind of the command's result. The first Ctrl-C or SIGTERM ends the
// handler's context and decides the verdict. A handler that has not
// returned once cleanup and the drain limit have passed gets no more time.
func (l loaded) handle(s selection, p *i18n.Printer, g command.Globals, codes contract.Registry, out *command.Output, stdout, stderr *os.File) int {
	kind := s.command.Result.Kind
	call := s.call
	call.JSON, call.Printer = g.JSON, p
	call.domain, call.services = l.manifest.Slug(), s.command.Requires.Keys
	call.programs = map[string]contract.Requirement{}
	for _, name := range s.command.Requires.Bins {
		call.programs[name] = l.manifest.Requirements[name]
	}
	call.workers = &workers{resources: l.Resources, domain: l.manifest.Domain}
	call.Mode = contract.DetectMode(term.IsTerminal(int(os.Stdin.Fd())), term.IsTerminal(int(stdout.Fd())), g.NoInput, g.JSON, os.LookupEnv)
	call.Stdin, call.Stdout, call.Stderr = os.Stdin, io.Writer(out), io.Writer(stderr)
	if g.JSON && kind != contract.ResultText {
		call.Stdout = io.Discard
	}
	// Run directly, the extension resolves the global flags itself, so that
	// its handler and the handler's programs see the variables that ncly
	// would set.
	for name, value := range g.Vars() {
		if value == "" {
			_ = os.Unsetenv(name)
		} else {
			_ = os.Setenv(name, value)
		}
	}
	ctx := runner.OnSignal()
	done := make(chan contract.Outcome, 1)
	go func() { done <- l.Handlers[call.Path](ctx, call) }()
	var o contract.Outcome
	select {
	case o = <-done:
	case <-ctx.Done():
		// The runner ends the handler's programs with SIGKILL when cleanup
		// ends, then reads their streams for the drain limit, so the handler
		// gets that long to return what it has. Under ncly, core kills the
		// extension when its own cleanup ends. The workers therefore go a
		// second before that, even if a program still runs.
		early := time.AfterFunc(max(runner.Cleanup-workersMargin, 0), call.workers.remove)
		defer early.Stop()
		select {
		case o = <-done:
		case <-time.After(runner.Cleanup + runner.DrainLimit):
			o = contract.Outcome{}
		}
	}
	// The workers go once the handler returns, or once it gets no more time.
	call.workers.remove()
	if ctx.Err() != nil {
		o.Signal = contract.StopSignal(runner.SignalOf(ctx))
	}
	if out.Err() != nil {
		return command.Report(p, g, codes, command.WriteFailure(p), stdout, stderr)
	}
	if kind == contract.ResultText && codes.Decide(o).Exit == contract.ExitOK {
		// A text result prints its text, never an answer, so its warnings
		// print as text too.
		g.JSON = false
	}
	return command.Report(p, g, codes, o, stdout, stderr)
}
