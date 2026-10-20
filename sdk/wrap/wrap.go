// Package wrap runs one program for a handler, such as a worker, as Wrapped
// programs in contract.md defines. It checks the program as doctor does,
// and starts it without the terminal, with an empty stdin. The program gets
// only the keys that it names, and wrap keeps its output up to a limit. Run
// turns each failure into the result that the contract defines for it, and
// the handler's decoder maps the rest. A handler that runs more than one
// program uses sdk/runner directly.
//
// Pass Run the ctx of the handler, which only a signal ends, so that a
// deadline of the caller never reads as a signal. Program.Timeout bounds
// the program instead.
package wrap

import (
	"cmp"
	"context"
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
}

// Ended is how a program ended on its own, which the decoder of Run maps.
type Ended struct {
	// Exit is the exit code of the program.
	Exit int
	// Stdout holds all that the program wrote on stdout.
	Stdout []byte
	// Stderr holds what the program wrote on stderr, up to the limit.
	Stderr  []byte
	name    string
	printer *i18n.Printer
}

// Failed returns the RUNTIME outcome of an end that the decoder does not
// map, such as an exit other than 0. A decoder returns it for every exit
// that it does not know, so no exit code of the program passes through.
func (e Ended) Failed() contract.Outcome {
	data := map[string]any{"Name": e.name, "Exit": e.Exit}
	return runtime(e.printer.T("wrap.failed", data), e.printer)
}

// Malformed returns the RUNTIME outcome of output that the decoder cannot
// read, such as missing or invalid JSON. A decoder returns it, so it never
// turns a failure into a success.
func (e Ended) Malformed() contract.Outcome {
	return runtime(e.printer.T("wrap.malformed", map[string]any{"Name": e.name}), e.printer)
}

// Run checks the program, then runs it from the caller's folder, and
// returns what decode makes of its end. A failure that the contract defines
// never reaches decode. Run answers PREREQ_MISSING for a program that is
// missing, too old, or whose probe fails or times out. It answers RUNTIME
// for a timeout, a start that fails, and stdout over the limit. A signal
// gives the outcome of that signal.
func Run(ctx context.Context, call extension.Call, p Program, decode func(Ended) contract.Outcome) contract.Outcome {
	printer := call.Printer
	data := map[string]any{"Name": p.Name, "Path": call.Path}
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
	path, failed := check(ctx, printer, p.Name, r, env, probe)
	if failed != nil {
		return *failed
	}
	limit := cmp.Or(p.OutputLimit, OutputLimit)
	stdout, stderr := runner.Capture{Limit: limit}, runner.Capture{Limit: limit}
	result, err := runner.Run(work, runner.Program{Path: path, Args: p.Args, Env: env, Stdout: &stdout, Stderr: &stderr})
	switch {
	case ctx.Err() != nil:
		return contract.Outcome{Signal: contract.StopSignal(runner.SignalOf(ctx))}
	case err != nil:
		return runtime(printer.T("wrap.start_failed", data), printer)
	case result.Signal != nil:
		// Only the deadline of Timeout stops the program when no signal
		// ended ctx, and the runner then stopped its whole process tree.
		data["Timeout"] = printer.Duration(p.Timeout)
		return runtime(printer.T("wrap.timeout", data), printer)
	case stdout.Over():
		data["Limit"] = printer.Size(int64(limit))
		return runtime(printer.T("wrap.output_over", data), printer)
	}
	return decode(Ended{Exit: result.Exit, Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), name: p.Name, printer: printer})
}

// check checks the program with the rule of Required programs in
// core-spec.md. It returns where PATH found it, or the outcome that stops
// the call. That outcome is the one of a signal, or PREREQ_MISSING, whose
// hint and next hold the install command of this machine.
func check(ctx context.Context, printer *i18n.Printer, name string, r contract.Requirement, env []string, limit time.Duration) (string, *contract.Outcome) {
	found, err := runner.CheckProgram(ctx, name, r, env, limit)
	data := map[string]any{"Name": name, "Found": found.Version, "Version": r.Version, "Exit": found.Exit, "Probe": strings.Join(append([]string{name}, r.ProbeArgs()...), " ")}
	var message, hint string
	switch {
	case err != nil:
		data["Exit"] = -1
		message = printer.T("wrap.probe_failed", data)
	case found.Signal != nil:
		return "", &contract.Outcome{Signal: contract.StopSignal(found.Signal)}
	case found.Verdict == runner.ProgramFound:
		return found.Path, nil
	case found.Verdict == runner.ProgramMissing:
		message = printer.T("wrap.missing", data)
	case found.Verdict == runner.ProgramTooOld:
		message = printer.T("wrap.too_old", data)
	case found.Verdict == runner.ProgramProbeFailed:
		message = printer.T("wrap.probe_failed", data)
	case found.Verdict == runner.ProgramNoVersion:
		message = printer.T("wrap.no_version", data)
	default:
		// ProgramCheck.Failed leaves a probe that timed out to its caller.
		message, hint = printer.T("wrap.probe_timeout", data), printer.T("wrap.probe_timeout_hint", data)
	}
	problem := contract.Problem{Code: contract.PrereqMissing, Message: message, Hint: cmp.Or(hint, printer.T("wrap.install", data))}
	if found.Install != nil {
		problem.Hint, problem.Next = strings.Join(found.Install, " "), found.Install
	}
	return "", &contract.Outcome{Errors: []contract.Problem{problem}}
}

// runtime returns the RUNTIME outcome of a message, whose hint warns that
// a repeat is not safe, because the program may have made its effects.
func runtime(message string, printer *i18n.Printer) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: message, Hint: printer.T("outcome.check_hint")}}}
}
