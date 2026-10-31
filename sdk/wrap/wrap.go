// Package wrap runs one program for a handler, such as a worker, as Wrapped
// programs in contract.md defines. It checks the program as doctor does,
// and starts it without the terminal, with an empty stdin. The program gets
// only the keys that it names. Run keeps its output up to a limit, or drops
// the stdout that the handler does not read. Run turns each failure into
// the result that the contract defines for it, and the handler's decoder
// maps the rest. A handler that runs more than one program uses sdk/runner
// directly.
//
// Pass Run the ctx of the handler, which only a signal ends, so that a
// deadline of the caller never reads as a signal. Program.Timeout bounds
// the program instead. A probe that uses up that deadline gives the
// timeout, not PREREQ_MISSING, because the program is there.
//
// Set Program.Input to the item or path that the call works on. Each
// RUNTIME message that Run makes then names it. The answer to a signal
// names it too. The handler cannot add the input to these outcomes.
package wrap

import (
	"cmp"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// OutputLimit is the default bound of what Run keeps of stdout, and of
// stderr: 1 MB, in the decimal units that the SDK shows.
const OutputLimit = 1_000_000

// ProbeLimit bounds the version probe of a program, within its Timeout.
const ProbeLimit = 10 * time.Second

// Program is one program that a handler runs through Run.
type Program struct {
	// Name is a program that the command declares in requires.bins, such as
	// bash. Run looks it up in PATH and probes its version as doctor does.
	Name string
	// Args follow the name, such as the path that Call.Worker returns.
	Args []string
	// Keys holds the services whose keys the program receives, each one that
	// the command declares. Run reads each key through Call.Key, and a key
	// that Call.Key does not return stops the call before any work.
	Keys []string
	// Env sets more variables of the program. Run drops each name that ends
	// in _API_KEY, so a key reaches the program only through Keys.
	Env map[string]string
	// Timeout bounds the probe and the run together. Zero means no bound.
	Timeout time.Duration
	// OutputLimit bounds what Run keeps of stdout, and of stderr, in bytes.
	// Zero means the OutputLimit of the package.
	OutputLimit int
	// DiscardStdout drops all that the program writes on stdout, so stdout
	// never passes OutputLimit and Ended.Stdout stays empty. Set it when the
	// decoder reads no stdout, such as for an encoder that prints its
	// progress there. The program still writes to a pipe, as without it.
	DiscardStdout bool
	// Input is the item or path that the call works on, such as the file
	// that the handler checked. When set, the messages of the timeout, of a
	// signal, and of each other RUNTIME outcome of Run and Ended name it.
	// Empty keeps the messages without an input.
	Input string
}

// Ended is how a program ended on its own, which the decoder of Run maps.
type Ended struct {
	// Exit is the exit code of the program.
	Exit int
	// Stdout holds all that the program wrote on stdout, and nothing when
	// Program.DiscardStdout is set.
	Stdout []byte
	// Stderr holds what the program wrote on stderr, up to the limit.
	Stderr  []byte
	name    string
	input   string
	printer *i18n.Printer
}

// Failed returns the RUNTIME outcome of an end that the decoder does not
// map, such as an exit other than 0. A decoder returns it for every exit
// that it does not know, so no exit code of the program passes through.
func (e Ended) Failed() contract.Outcome {
	data := map[string]any{"Name": e.name, "Exit": e.Exit, "Input": e.input}
	return runtime(e.printer.T(named("wrap.failed", e.input), data), e.printer)
}

// Malformed returns the RUNTIME outcome of output that the decoder cannot
// read, such as missing or invalid JSON. A decoder returns it, so it never
// turns a failure into a success.
func (e Ended) Malformed() contract.Outcome {
	data := map[string]any{"Name": e.name, "Input": e.input}
	return runtime(e.printer.T(named("wrap.malformed", e.input), data), e.printer)
}

// Run checks the program, then runs it from the caller's folder, and
// returns what decode makes of its end. A failure that the contract defines
// never reaches decode. Run answers PREREQ_MISSING for a program that is
// missing, too old, or whose probe fails or reaches ProbeLimit. It answers
// RUNTIME for a timeout, also one during the probe, or a start that fails.
// Without DiscardStdout, stdout over the limit gives RUNTIME too. A signal
// gives the outcome of that signal.
func Run(ctx context.Context, call extension.Call, p Program, decode func(Ended) contract.Outcome) contract.Outcome {
	printer := call.Printer
	data := map[string]any{"Name": p.Name, "Path": call.Path, "Input": p.Input}
	fail := func(id string) contract.Outcome { return runtime(printer.T(named(id, p.Input), data), printer) }
	r, declared := call.Requirement(p.Name)
	if !declared {
		return runtime(printer.T("wrap.undeclared", data), printer)
	}
	set := map[string]string{}
	for name, value := range p.Env {
		if !strings.HasSuffix(name, "_API_KEY") {
			set[name] = value
		}
	}
	for _, service := range p.Keys {
		key, problem := call.Key(service)
		if problem != nil {
			return contract.Outcome{Errors: []contract.Problem{*problem}}
		}
		set[strings.ToUpper(service)+"_API_KEY"] = key
	}
	env := runner.Environ(os.Environ(), set)
	work, probe := ctx, ProbeLimit
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		work, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
		probe = min(probe, p.Timeout)
	}
	found, problem := check(ctx, printer, p.Name, r, env, probe)
	switch {
	case found.Signal != nil:
		return stopped(found.Signal, p.Input, printer)
	case found.Verdict == runner.ProgramTimedOut && p.Timeout > 0 && expired(work):
		// The probe used up the deadline of Timeout, so the program is there.
		data["Timeout"] = printer.Duration(p.Timeout)
		return fail("wrap.timeout")
	case problem != nil:
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	limit := cmp.Or(p.OutputLimit, OutputLimit)
	stdout, stderr := runner.Capture{Limit: limit}, runner.Capture{Limit: limit}
	var out io.Writer = &stdout
	if p.DiscardStdout {
		out = io.Discard
	}
	result, err := runner.Run(work, runner.Program{Path: found.Path, Args: p.Args, Env: env, Stdout: out, Stderr: &stderr})
	switch {
	case ctx.Err() != nil:
		return stopped(runner.SignalOf(ctx), p.Input, printer)
	case err != nil:
		return fail("wrap.start_failed")
	case result.Signal != nil:
		// Only the deadline of Timeout stops the program when no signal
		// ended ctx, and the runner then stopped its whole process tree.
		data["Timeout"] = printer.Duration(p.Timeout)
		return fail("wrap.timeout")
	case stdout.Over():
		data["Limit"] = printer.Size(int64(limit))
		return fail("wrap.output_over")
	}
	return decode(Ended{Exit: result.Exit, Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), name: p.Name, input: p.Input, printer: printer})
}

// runtime returns the RUNTIME outcome of a message, whose hint warns that
// a repeat is not safe, because the program may have made its effects.
func runtime(message string, printer *i18n.Printer) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: message, Hint: printer.T("outcome.check_hint")}}}
}

// named returns the catalog ID of a message that names the input, or id
// itself when the handler set no input.
func named(id, input string) string {
	if input == "" {
		return id
	}
	return id + "_input"
}

// expired reports whether the deadline of ctx has passed. It reads the
// clock, because ctx may end a moment after its deadline.
func expired(ctx context.Context) bool {
	deadline, ok := ctx.Deadline()
	return ok && !time.Now().Before(deadline)
}

// stopped returns the outcome of a signal. With an input, its error names
// that input, and Finish puts this error first, because its code leads.
func stopped(sig os.Signal, input string, printer *i18n.Printer) contract.Outcome {
	o := contract.Outcome{Signal: contract.StopSignal(sig)}
	if input == "" {
		return o
	}
	code, id := contract.Interrupted, "wrap.interrupted"
	if o.Signal == contract.Terminate {
		code, id = contract.Terminated, "wrap.terminated"
	}
	message := printer.T(id, map[string]any{"Input": input})
	o.Errors = []contract.Problem{{Code: code, Message: message, Hint: printer.T("outcome.check_hint")}}
	return o
}
