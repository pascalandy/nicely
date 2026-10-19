package transcript

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/pascalandy/nicely/sdk/runner"
)

// terminal applies the one terminal step of the engine. It checks the step
// against the table and waits for the engine to exit 0. Then it records the
// completion of verified files, or the failure that the step reports.
func (a *attempt) terminal(ctx context.Context, m message) contract.Outcome {
	acked := a.dir != ""
	broken := record.StepPending
	if acked {
		broken = record.StepUnknown
	}
	status, defined := stepStatus(m, acked)
	if !defined {
		return a.end(broken, a.s.broke(a.p, "transcript.protocol.code"))
	}
	a.codes = m.Warnings
	if m.Detail != "" && os.Getenv("NCLY_VERBOSE") == "1" {
		_, _ = fmt.Fprintln(a.call.Stderr, m.Detail)
	}
	if problem := a.s.finish(ctx, a.p, a.j); problem != nil || ctx.Err() != nil {
		if ctx.Err() != nil {
			return a.stop(ctx)
		}
		return a.end(broken, problem)
	}
	if status != record.StepCompleted {
		return a.end(status, itemProblem(a.p, *m.Code, a.j.url, a.run.ID, acked))
	}
	artifacts, err := collect(a.dir)
	if err != nil {
		return a.end(record.StepUnknown, inspect(a.p, contract.Runtime, a.run.ID, "transcript.run.artifacts_invalid", map[string]any{"Dir": a.dir}))
	}
	if err := a.run.Complete(0, 0, artifacts, evidence{FormatVersion: 1, Warnings: append([]string{}, a.codes...)}); err != nil {
		return a.answer(unsaved(a.p, filepath.Dir(a.run.File())))
	}
	o := success(a.call, a.run.ID, a.j.url, a.dir, a.codes)
	o.Effect = a.run.Effect()
	return o
}

// end records how the attempt ended without completion. It saves the step
// status that the evidence supports, with the problem, and the run status
// that follows from it. Then it answers the item.
func (a *attempt) end(status record.StepStatus, problem *contract.Problem) contract.Outcome {
	run := record.Failed
	if status == record.StepUnknown {
		run = record.Unknown
	}
	_ = a.run.End(0, 0, status, *problem, run)
	return a.answer(problem)
}

// stop ends an attempt that Ctrl-C, SIGTERM, or the timeout stopped. Its
// step stays pending before the intent and unknown after it.
func (a *attempt) stop(ctx context.Context) contract.Outcome {
	a.s.close()
	problem := stopCause(ctx, a.p, a.j)
	status, run := record.StepPending, record.Interrupted
	if problem != nil {
		run = record.Failed
	} else {
		problem = &contract.Problem{Code: stopCode(ctx), Message: a.p.T("transcript.run.stopped"), Hint: "ncly run view " + a.run.ID, Next: []string{"ncly", "run", "view", a.run.ID}}
	}
	if a.begun {
		status, run = record.StepUnknown, record.Unknown
	}
	_ = a.run.End(0, 0, status, *problem, run)
	return a.answer(problem)
}

// answer is the outcome of an attempted item that did not complete. It
// holds the run, the item with its problem and the result folder once
// reserved, and the effect so far.
func (a *attempt) answer(problem *contract.Problem) contract.Outcome {
	item := itemData{URL: a.j.url}
	if a.dir != "" {
		item.OutputDir = &a.dir
	}
	if !a.call.JSON {
		label := a.p.T("error.label", map[string]any{"Code": problem.Code})
		_, _ = fmt.Fprintf(a.call.Stderr, "%s %s\n\n  %s %s\n\n", label, problem.Message, a.p.T("error.hint"), problem.Hint)
	}
	return contract.Outcome{
		Data:     runData{RunID: a.run.ID},
		Results:  []contract.Item{{Data: item, Errors: []contract.Problem{*problem}}},
		Warnings: warnings(a.p, a.codes),
		Effect:   a.run.Effect(),
	}
}

// refuse answers a failure before the record of this attempt: no item, and
// the run ID only when an earlier attempt recorded the run. A nil problem
// is a signal, which the SDK answers.
func (a *attempt) refuse(problem *contract.Problem) contract.Outcome {
	if a.saved {
		return failed(problem, a.run.ID)
	}
	return failed(problem, "")
}

// savedID is the ID of a run that a record already holds, which an answer
// names, and empty for a new run.
func savedID(run *record.Run) string {
	if run.Record == nil {
		return ""
	}
	return run.ID
}

// stopCause is the problem of a call that its timeout stopped, or nil for a
// signal, which the SDK answers.
func stopCause(ctx context.Context, p *i18n.Printer, j job) *contract.Problem {
	if errors.Is(context.Cause(ctx), errTimedOut) {
		return timedOut(p, j)
	}
	return nil
}

// stopCode is the code of the signal that ended ctx: INTERRUPTED for
// Ctrl-C, and TERMINATED for SIGTERM.
func stopCode(ctx context.Context) contract.Code {
	if contract.StopSignal(runner.SignalOf(ctx)) == contract.Interrupt {
		return contract.Interrupted
	}
	return contract.Terminated
}
