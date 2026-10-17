package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/pascalandy/nicely/sdk/runner"
)

// answer is the answer of a completed run, as ncly agent in spec.md
// defines it.
type answer struct {
	RunID      string `json:"run_id"`
	Answer     string `json:"answer"`
	Harness    string `json:"harness"`
	Model      string `json:"model"`
	DurationMS int64  `json:"duration_ms"`
}

// evidence is what the step codex keeps beside its artifact, so a repeated
// call answers what the first call answered.
type evidence struct {
	FormatVersion int   `json:"format_version"`
	DurationMS    int64 `json:"duration_ms"`
}

// The one item and the one step of a run, as Records and replay in
// spec.md names them.
const (
	itemKey  = "task"
	stepID   = "codex"
	answerTo = "answer.txt"
)

// store is the runs folder of the state folder.
func store() record.Store {
	home, _ := os.UserHomeDir()
	return record.Store{Dir: record.Folder(config.Locate(home, os.LookupEnv).State)}
}

// execute runs the prepared task under its run ID, the one that the call
// checked or a new one when id is empty. It takes the run lock, decides on
// a saved record, as Records and replay in spec.md defines, and for a new
// task persists the record and the intent of the step codex before Codex
// starts, then publishes the verified answer and the completion before it
// answers success.
func execute(ctx context.Context, call extension.Call, t task, id string) contract.Outcome {
	p := call.Printer
	runs := store()
	if id == "" {
		id = record.NewID(time.Now())
	}
	keep := map[string]string{"run_id": id}
	data := map[string]any{"ID": id}
	unsaved := func() contract.Outcome {
		dir := map[string]any{"Dir": runs.Dir}
		return failure(&contract.Problem{Code: contract.Runtime, Message: p.T("agent.run.record_failed", dir), Hint: p.T("agent.run.record_failed_hint", dir)}, keep)
	}
	run, err := runs.Open(ctx, id)
	var bad *record.UnreadableError
	switch {
	case errors.Is(err, record.ErrBusy):
		return failure(inspect(p, contract.Temporary, id, "agent.run.busy", data), keep)
	case errors.As(err, &bad):
		return failure(unreadable(p, bad.File, id), keep)
	case ctx.Err() != nil:
		return stopped(ctx, p, t, keep)
	case err != nil:
		return unsaved()
	}
	defer run.Close()
	var resolved map[string]any
	if run.Record != nil {
		completed, problem := decide(p, call, t, run.Record, run.File(), id)
		if problem != nil {
			return failure(problem, keep)
		}
		if completed {
			return savedAnswer(call, run.Record, run.File(), id, contract.NoEffect)
		}
		resolved = run.Record.Resolved
	}
	problem := pick(ctx, call, &t, resolved)
	if ctx.Err() != nil {
		return stopped(ctx, p, t, keep)
	}
	if problem != nil {
		return failure(problem, keep)
	}
	tools := map[string]string{"codex": t.version}
	if run.Record != nil {
		err = run.Again(nclyVersion(), tools)
	} else {
		folder := run.ArtifactDir()
		err = run.Create(record.Record{
			Path:        call.Path,
			NclyVersion: nclyVersion(),
			Inputs:      inputs(call, t),
			Resolved:    map[string]any{"harness": "codex", "model": t.model, "effort": t.effort},
			Tools:       tools,
			Items:       []record.Item{{Key: itemKey, OutputDir: &folder, Steps: []record.Step{{ID: stepID, Effect: record.EffectPaid}}}},
		})
	}
	if err != nil {
		return unsaved()
	}
	if err := run.Begin(0, 0); err != nil {
		return unsaved()
	}
	text, duration, problem, started := launch(ctx, call, t, id)
	if problem == nil {
		var artifact record.Artifact
		if artifact, err = run.Publish(answerTo, text); err == nil {
			err = run.Complete(0, 0, []record.Artifact{artifact}, evidence{FormatVersion: 1, DurationMS: duration.Milliseconds()})
		}
		if err != nil {
			problem = inspect(p, contract.Runtime, id, "agent.run.publish_failed", data)
		}
	}
	if problem != nil {
		// Once Codex started, its paid work may have happened, so nothing
		// proves that it had no effect: the step stays unknown. A start that
		// failed, or that a stop came before, ran no task.
		step, status := record.StepUnknown, record.Unknown
		if !started {
			step, status = record.StepFailed, record.Failed
			if problem.Code == contract.Interrupted || problem.Code == contract.Terminated {
				status = record.Interrupted
			}
		}
		_ = run.End(0, 0, step, *problem, status)
		o := failure(problem, keep)
		o.Effect = run.Effect()
		return o
	}
	o := savedAnswer(call, run.Record, run.File(), id, run.Effect())
	if ctx.Err() != nil && !timedOut(ctx) && o.Errors == nil {
		// A signal came once Codex had answered: the answer and the
		// completion stay saved, and the signal still decides the exit code.
		o.Errors = []contract.Problem{*inspect(p, stopCode(ctx), id, "agent.run.stopped_after", data)}
	}
	return o
}

// decide applies Records and replay in spec.md to the record that holds
// the ID of the call, once the record could be read: the problem of cases
// 2 and 5, true for a completed run, whose saved answer the call returns,
// and false for a run whose step codex is still pending, which may run
// again.
func decide(p *i18n.Printer, call extension.Call, t task, saved *record.Record, file, id string) (bool, *contract.Problem) {
	data := map[string]any{"ID": id}
	if !resolvable(saved.Resolved) {
		return false, unreadable(p, file, id)
	}
	if saved.Path != call.Path || !sameValues(saved.Inputs, inputs(call, t)) || len(saved.Items) != 1 || saved.Items[0].Key != itemKey {
		return false, inspect(p, contract.UsageInvalid, id, "agent.run.other_call", data)
	}
	steps := saved.Items[0].Steps
	switch {
	case len(steps) != 1 || steps[0].ID != stepID:
	case saved.Status == record.Completed && steps[0].Status == record.StepCompleted:
		return true, nil
	case steps[0].Status == record.StepPending:
		return false, nil
	}
	return false, inspect(p, contract.ResumeUnsafe, id, "agent.run.effect_unknown", data)
}

// resolvable reports whether the resolved values of a record hold what a
// repeated call reads from them, as case 1 of Records and replay in spec.md
// lists them: the harness codex, a model, and a known effort.
func resolvable(resolved map[string]any) bool {
	model, _ := resolved["model"].(string)
	effort, _ := resolved["effort"].(string)
	return resolved["harness"] == "codex" && model != "" && slices.Contains(efforts, effort)
}

// sameValues reports whether two sets of inputs hold the same values, as
// JSON compares them, so the values of a record match those of a call.
func sameValues(a, b map[string]any) bool {
	x, errA := json.Marshal(a)
	y, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(x, y)
}

// savedAnswer answers a completed run from its committed record: the
// verified artifact of its step codex, the model of its resolved values,
// and the duration of its evidence, so a repeated call answers what the
// first call answered, and never from what Codex wrote before completion.
func savedAnswer(call extension.Call, saved *record.Record, file, id string, effect contract.Effect) contract.Outcome {
	p := call.Printer
	text, ev, problem := verified(p, saved, file, id)
	if problem != nil {
		return failure(problem, map[string]string{"run_id": id})
	}
	model, _ := saved.Resolved["model"].(string)
	if !call.JSON {
		_, _ = call.Stdout.Write(text)
		if !bytes.HasSuffix(text, []byte("\n")) {
			_, _ = io.WriteString(call.Stdout, "\n")
		}
	}
	return contract.Outcome{Data: answer{RunID: id, Answer: string(text), Harness: "codex", Model: model, DurationMS: ev.DurationMS}, Effect: effect}
}

// verified reads the saved answer of a completed run, with the evidence of
// its step codex, and checks that the answer still matches the size and the
// SHA-256 of its artifact, as case 3 of Records and replay in spec.md asks.
// A dry run and execution share it, so they decide alike.
func verified(p *i18n.Printer, saved *record.Record, file, id string) ([]byte, evidence, *contract.Problem) {
	step := saved.Items[0].Steps[0]
	var ev struct {
		FormatVersion int    `json:"format_version"`
		DurationMS    *int64 `json:"duration_ms"`
	}
	if len(step.Artifacts) != 1 || json.Unmarshal(step.Evidence, &ev) != nil || ev.FormatVersion != 1 || ev.DurationMS == nil || *ev.DurationMS < 0 {
		return nil, evidence{}, unreadable(p, file, id)
	}
	text, err := step.Artifacts[0].Read()
	if err != nil {
		return nil, evidence{}, inspect(p, contract.ResumeUnsafe, id, "agent.run.answer_changed", map[string]any{"File": step.Artifacts[0].Path})
	}
	return text, evidence{FormatVersion: 1, DurationMS: *ev.DurationMS}, nil
}

// unreadable is the RESUME_UNSAFE problem of a record that the call cannot
// use, which never counts as an absent one.
func unreadable(p *i18n.Printer, file, id string) *contract.Problem {
	return inspect(p, contract.ResumeUnsafe, id, "agent.run.record_unreadable", map[string]any{"File": file})
}

// failure is an outcome with one problem, which keeps the run_id of a
// recorded call, as Output in contract.md asks.
func failure(p *contract.Problem, keep map[string]string) contract.Outcome {
	o := contract.Outcome{Errors: []contract.Problem{*p}}
	if keep != nil {
		o.Data = keep
	}
	return o
}

// inputs are what the call states, as Records and replay in spec.md lists
// them: the task by its SHA-256, the absolute folder, tools, and the model
// and the effort only when a flag gives them, so a change of the config
// never changes what the call states.
func inputs(call extension.Call, t task) map[string]any {
	in := map[string]any{"task_sha256": t.sha256, "cwd": t.cwd, "tools": true}
	for _, name := range []string{"model", "effort"} {
		if value, given := call.Flags[name]; given {
			in[name] = value
		}
	}
	return in
}

// nclyVersion is the release that writes the record: the one that core
// passes, or the build of this extension when it runs alone.
func nclyVersion() string {
	if v := os.Getenv("NCLY_VERSION"); v != "" {
		return v
	}
	return command.ModuleVersion()
}

// launch runs Codex with the prepared invocation in the run's own
// temporary folder, sends the task on stdin, and returns the answer that
// Codex wrote, with how long it ran. On failure, it returns the problem,
// and whether Codex started, so its paid work may have happened, as Stops
// and failures in spec.md defines.
func launch(ctx context.Context, call extension.Call, t task, id string) ([]byte, time.Duration, *contract.Problem, bool) {
	p := call.Printer
	data := map[string]any{"ID": id, "Limit": limitText(t.timeout)}
	tmp, err := os.MkdirTemp("", "ncly-run-"+id+"-")
	if err != nil {
		return nil, 0, inspect(p, contract.Runtime, id, "agent.run.start_failed", data), false
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	file := filepath.Join(tmp, answerTo)
	// Codex gets no pipe for the output that the call never reads, so a
	// descendant that keeps it open never holds the call. Its stderr reaches
	// the stderr of ncly only with --verbose.
	var stderr io.Writer
	if os.Getenv("NCLY_VERBOSE") == "1" {
		stderr = call.Stderr
	}
	start := time.Now()
	result, err := runner.Run(ctx, runner.Program{Path: t.codex, Args: t.args(file), Env: t.env, Stdin: bytes.NewReader(t.text), Stderr: stderr})
	duration := time.Since(start)
	switch {
	case err != nil:
		return nil, duration, inspect(p, contract.Runtime, id, "agent.run.start_failed", data), false
	case result.Exit == -1 && timedOut(ctx):
		// The stop came before the start, so Codex never ran.
		return nil, duration, inspect(p, contract.Runtime, id, "agent.run.timed_out_early", data), false
	case result.Exit == -1:
		return nil, duration, inspect(p, stopCode(ctx), id, "agent.run.stopped_early", data), false
	case result.Signal != nil && timedOut(ctx):
		return nil, duration, inspect(p, contract.Runtime, id, "agent.run.timed_out", data), true
	case result.Signal != nil:
		return nil, duration, inspect(p, stopCode(ctx), id, "agent.run.stopped", data), true
	case result.Exit != 0:
		return nil, duration, inspect(p, contract.Runtime, id, "agent.run.codex_failed", data), true
	}
	text, err := os.ReadFile(file)
	if err != nil || strings.TrimSpace(string(text)) == "" {
		return nil, duration, inspect(p, contract.Runtime, id, "agent.run.answer_missing", data), true
	}
	return text, duration, nil, true
}

// stopped answers a call that a signal or the timeout stopped before its
// record, with keep: the signal, which the SDK answers, or RUNTIME, whose
// hint asks for a longer timeout.
func stopped(ctx context.Context, p *i18n.Printer, t task, keep map[string]string) contract.Outcome {
	if !timedOut(ctx) {
		var o contract.Outcome
		if keep != nil {
			o.Data = keep
		}
		return o
	}
	message := p.T("agent.run.timed_out_early", map[string]any{"Limit": limitText(t.timeout)})
	return failure(&contract.Problem{Code: contract.Runtime, Message: message, Hint: p.T("agent.run.timed_out_early_hint")}, keep)
}

// timedOut reports whether the timeout of the call ended ctx, rather than a
// signal.
func timedOut(ctx context.Context) bool { return errors.Is(context.Cause(ctx), errTimedOut) }

// stopCode is the code of the signal that ended ctx: INTERRUPTED for
// Ctrl-C, and TERMINATED for SIGTERM.
func stopCode(ctx context.Context) contract.Code {
	if contract.StopSignal(runner.SignalOf(ctx)) == contract.Interrupt {
		return contract.Interrupted
	}
	return contract.Terminated
}

// limitText writes a timeout as a caller writes it, such as 30s, 5m, or
// 1h30m, without the zero units that time.Duration adds.
func limitText(d time.Duration) string {
	text := d.String()
	if strings.HasSuffix(text, "m0s") {
		text = strings.TrimSuffix(text, "0s")
	}
	if strings.HasSuffix(text, "h0m") {
		text = strings.TrimSuffix(text, "0m")
	}
	return text
}

// inspect is a problem whose hint and next inspect the run before anything
// else, such as a failure after the launch of Codex, whose effect stays
// unknown.
func inspect(p *i18n.Printer, code contract.Code, id, message string, data map[string]any) *contract.Problem {
	view := []string{"ncly", "run", "view", id}
	return &contract.Problem{Code: code, Message: p.T(message, data), Hint: strings.Join(view, " "), Next: view}
}
