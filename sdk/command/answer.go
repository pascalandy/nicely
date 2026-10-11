package command

import (
	"fmt"
	"io"
	"os"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
)

// Output records the first failed write to a file, so lost output never
// reports success. Beyond Write, it keeps only Fd, for terminal detection,
// so that no other method of the file can write around Write.
type Output struct {
	file *os.File
	err  error
}

// NewOutput wraps a file, such as stdout.
func NewOutput(file *os.File) *Output { return &Output{file: file} }

func (o *Output) Write(b []byte) (int, error) {
	n, err := o.file.Write(b)
	if err != nil && o.err == nil {
		o.err = err
	}
	return n, err
}

func (o *Output) Fd() uintptr { return o.file.Fd() }

// Err returns the first write error, or nil.
func (o *Output) Err() error { return o.err }

// Failure carries the outcome of a command that failed, as the error that
// the command returns.
type Failure struct{ Outcome contract.Outcome }

func (f *Failure) Error() string { return string(f.Outcome.Errors[0].Code) }

// Fail returns the failure of a command whose errors are problems.
func Fail(problems ...contract.Problem) *Failure {
	return &Failure{contract.Outcome{Errors: problems}}
}

// Report prints an outcome and returns its exit code. An answer that
// cannot be written is a runtime failure, whatever the verdict: when stderr
// itself refuses it, the exit code is all that remains to tell.
func Report(p *i18n.Printer, g Globals, o contract.Outcome, stdout, stderr io.Writer) int {
	a := contract.Finish(o, Summarizer(p))
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
		return int(contract.ExitRuntime)
	}
	return int(a.Exit)
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
// explains, in the language of p, as contract.Finish asks.
func Summarizer(p *i18n.Printer) func(contract.Code) contract.Problem {
	return func(code contract.Code) contract.Problem {
		s := summaries[code]
		return contract.Problem{Code: code, Message: p.T(s.message), Hint: p.T(s.hint)}
	}
}
