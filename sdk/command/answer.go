package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
)

// Output records the first failed write to a file, so lost output never
// reports success. Beyond Write, it keeps Fd, for terminal detection, and
// File, for a program that writes to the file itself. Write and Err are safe
// for concurrent use, because a handler that outlives its time may still
// write while the extension answers.
type Output struct {
	file *os.File
	// mu guards err alone, so a write that blocks never delays Err.
	mu  sync.Mutex
	err error
}

// NewOutput wraps a file, such as stdout.
func NewOutput(file *os.File) *Output { return &Output{file: file} }

func (o *Output) Write(b []byte) (int, error) {
	n, err := o.file.Write(b)
	if err != nil {
		o.mu.Lock()
		if o.err == nil {
			o.err = err
		}
		o.mu.Unlock()
	}
	return n, err
}

func (o *Output) Fd() uintptr { return o.file.Fd() }

// File returns the file, for a program that the runner starts, which then
// sees a terminal when the file is one. The program reports its own failed
// writes through its exit code.
func (o *Output) File() *os.File { return o.file }

// Err returns the first write error, or nil.
func (o *Output) Err() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.err
}

// Failure carries the outcome of a command that failed or that a signal
// stopped, as the error that the command returns.
type Failure struct{ Outcome contract.Outcome }

func (f *Failure) Error() string {
	if len(f.Outcome.Errors) == 0 {
		return "stopped by a signal"
	}
	return string(f.Outcome.Errors[0].Code)
}

// Fail returns the failure of a command whose errors are problems.
func Fail(problems ...contract.Problem) *Failure {
	return &Failure{contract.Outcome{Errors: problems}}
}

// Report prints an outcome and returns its exit code. An answer that
// cannot be written is a runtime failure, whatever the verdict: when stderr
// itself refuses it, the exit code is all that remains to tell.
func Report(p *i18n.Printer, g Globals, codes contract.Registry, o contract.Outcome, stdout, stderr io.Writer) int {
	a := codes.Finish(o, Summarizer(p))
	var err error
	if g.JSON {
		err = a.WriteJSON(stdout, stderr)
	} else {
		label := func(code contract.Code, warning bool) string {
			if warning {
				return p.T("warning.label", map[string]any{"Code": code})
			}
			return p.T("error.label", map[string]any{"Code": code})
		}
		_, err = fmt.Fprint(tui.Output(stderr, g.NoColor), tui.Problems(a.Errors(), a.Warnings(), label, p.T("error.hint")))
	}
	if err != nil {
		if g.JSON && len(a.Errors()) == 0 {
			// The answer belonged on stdout, so stderr may still take the
			// failure to write it, with the keys that still apply.
			lost := o
			lost.Errors, lost.HumanNeeded = WriteFailure(p).Errors, false
			return Report(p, g, codes, lost, stdout, stderr)
		}
		return int(contract.ExitRuntime)
	}
	return int(a.Exit)
}

// Finish returns the exit code of a command line whose run ended with err,
// after printing its failure: RUNTIME when out lost a write, the failure
// that err carries, or USAGE_INVALID with hint for any other error of the
// parser.
func Finish(p *i18n.Printer, g Globals, codes contract.Registry, out *Output, err error, hint string, stdout, stderr io.Writer) int {
	if out.Err() != nil {
		err = &Failure{Outcome: WriteFailure(p)}
	}
	if err == nil {
		return int(contract.ExitOK)
	}
	var failed *Failure
	if !errors.As(err, &failed) {
		failed = Fail(Usage(p.T("usage.invalid"), hint))
	}
	return Report(p, g, codes, failed.Outcome, stdout, stderr)
}

// WriteFailure is the outcome of an answer that could not be written.
func WriteFailure(p *i18n.Printer) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: p.T("output.write_failed"), Hint: p.T("output.write_failed_hint")}}}
}

// summaries hold the catalog IDs of the leading error of an outcome that no
// single failure explains, such as an interrupted run or an incomplete batch.
var summaries = map[contract.Code]struct{ message, hint string }{
	contract.Interrupted: {"outcome.interrupted", "outcome.check_hint"},
	contract.Terminated:  {"outcome.terminated", "outcome.check_hint"},
	contract.Temporary:   {"outcome.temporary", "outcome.temporary_hint"},
	contract.Runtime:     {"outcome.runtime", "outcome.check_hint"},
}

// Summarizer returns the leading error of an outcome that no single failure
// explains, in the language of p, as Registry.Finish asks.
func Summarizer(p *i18n.Printer) func(contract.Code) contract.Problem {
	return func(code contract.Code) contract.Problem {
		s := summaries[code]
		return contract.Problem{Code: code, Message: p.T(s.message), Hint: p.T(s.hint)}
	}
}
