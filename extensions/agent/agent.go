// Package agent is the external extension that runs one trusted task
// through Codex, as extensions/agent/spec.md defines. It imports only sdk/,
// and installs as ncly-agent with its manifest, catalogs, and skill beside it.
package agent

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

//go:embed skill
var skill embed.FS

// Extension is the agent extension, which cmd/ncly-agent runs.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{
		"agent run": run,
	},
}

// answerFile stands for the file that execution creates for Codex's last
// message, which a dry run never creates.
const answerFile = "<answer-file>"

// dryRun is the answer of ncly agent run --dry-run, as Dry run in spec.md
// defines it.
type dryRun struct {
	DryRun         bool              `json:"dry_run"`
	Harness        string            `json:"harness"`
	HarnessVersion string            `json:"harness_version"`
	Model          string            `json:"model"`
	Effort         string            `json:"effort"`
	Cwd            string            `json:"cwd"`
	TaskSHA256     string            `json:"task_sha256"`
	Invocation     []string          `json:"invocation"`
	Environment    map[string]string `json:"environment"`
	Plan           []plan            `json:"plan"`
	Pending        []check           `json:"pending_checks"`
}

// plan is the one item of a dry run, as Dry-run answer in contract.md
// defines it.
type plan struct {
	Key string `json:"key"`
	// Output is the folder that will hold the result of a new task, and
	// OutputDir the folder of a saved run that the call reuses.
	Output    string          `json:"output,omitempty"`
	OutputDir string          `json:"output_dir,omitempty"`
	Steps     []step          `json:"steps"`
	Effects   map[string]bool `json:"effects"`
}

type step struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// check is a pending check of a dry run: one that only execution makes.
type check struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// defaultTimeout bounds a call without --timeout, as the manifest declares.
const defaultTimeout = "5m"

// errTimedOut is the cause of a context that the timeout of the call ended.
var errTimedOut = errors.New("the call reached its timeout")

// run prepares the task, then runs it under its run ID. A dry run shows
// what execution would do, and starts no task. The timeout covers the whole
// call, from the start of preparation.
func run(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	value, given := call.Flags["timeout"]
	if !given {
		value = defaultTimeout
	}
	limit, valid := command.Duration(value)
	if !valid {
		return contract.Outcome{Errors: []contract.Problem{*usage(p, "agent.run.timeout_invalid", "agent.run.timeout_invalid_hint", map[string]any{"Value": i18n.Clip(value)})}}
	}
	// signals ends with Ctrl-C or SIGTERM alone, so it still tells a signal
	// once the timeout has ended ctx.
	signals := ctx
	ctx, cancel := context.WithTimeoutCause(signals, limit, errTimedOut)
	defer cancel()
	t, problem := prepare(call)
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	t.timeout = limit
	id, given := call.Flags["run-id"]
	if given && !record.ValidID(id) {
		return contract.Outcome{Errors: []contract.Problem{*usage(p, "agent.run.run_id_invalid", "agent.run.run_id_invalid_hint", map[string]any{"ID": i18n.Clip(id)})}}
	}
	if call.Flags["dry-run"] != "true" {
		return execute(ctx, signals, call, t, id)
	}
	return dryRunOf(ctx, call, t, id)
}

// dryRunOf answers what execution would do, as Dry run and Records and
// replay in spec.md define: the error that execution would answer, the
// reuse of a completed run, or the invocation of a new task. It reads the
// record of id without holding its lock, and writes nothing.
func dryRunOf(ctx context.Context, call extension.Call, t task, id string) contract.Outcome {
	p := call.Printer
	runs := store()
	var saved *record.Record
	if id != "" {
		keep := map[string]string{"run_id": id}
		active, err := runs.Active(id)
		if err == nil && active {
			return failure(inspect(p, contract.Temporary, id, "agent.run.busy", map[string]any{"ID": id}), keep)
		}
		if err == nil {
			saved, err = runs.Read(id)
		}
		if err != nil {
			return failure(unreadable(p, runs.File(id), id), keep)
		}
		if saved != nil {
			completed, problem := decide(p, call, t, saved, runs.File(id), id)
			if problem != nil {
				return failure(problem, keep)
			}
			if completed {
				if _, _, problem := verified(p, saved, runs.File(id), id); problem != nil {
					return failure(problem, keep)
				}
				return reused(call, t, saved, id)
			}
		}
	}
	var resolved map[string]any
	if saved != nil {
		resolved = saved.Resolved
	}
	problem := pick(ctx, call, &t, resolved)
	if ctx.Err() != nil {
		return stopped(ctx, p, t, nil)
	}
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	d := dryRun{
		DryRun:         true,
		Harness:        "codex",
		HarnessVersion: t.version,
		Model:          t.model,
		Effort:         t.effort,
		Cwd:            t.cwd,
		TaskSHA256:     t.sha256,
		Invocation:     append([]string{t.codex}, t.args(answerFile)...),
		Environment:    map[string]string{depthVariable: strconv.Itoa(t.depth)},
		Plan: []plan{{
			Key:     itemKey,
			Output:  runs.Dir,
			Steps:   []step{{ID: stepID, Action: "run"}},
			Effects: map[string]bool{"user_writes": true, "network": true, "paid": true},
		}},
		Pending: []check{{ID: "agent.login", Message: p.T("agent.run.pending_login")}},
	}
	if !call.JSON {
		writeDryRun(call.Stdout, p, d)
	}
	return contract.Outcome{Data: d}
}

// reused is the dry run of a completed run, which execution answers from
// its record without starting Codex: the values of the record, no
// invocation, and no pending check.
func reused(call extension.Call, t task, saved *record.Record, id string) contract.Outcome {
	p := call.Printer
	model, _ := saved.Resolved["model"].(string)
	effort, _ := saved.Resolved["effort"].(string)
	folder := ""
	if dir := saved.Items[0].OutputDir; dir != nil {
		folder = *dir
	}
	d := dryRun{
		DryRun:         true,
		Harness:        "codex",
		HarnessVersion: saved.Tools["codex"],
		Model:          model,
		Effort:         effort,
		Cwd:            t.cwd,
		TaskSHA256:     t.sha256,
		Invocation:     []string{},
		Environment:    map[string]string{},
		Plan: []plan{{
			Key:       itemKey,
			OutputDir: folder,
			Steps:     []step{{ID: stepID, Action: "reuse"}},
			Effects:   map[string]bool{"user_writes": false, "network": false, "paid": false},
		}},
		Pending: []check{},
	}
	if !call.JSON {
		for _, line := range []string{
			p.T("agent.run.planned_reuse", map[string]any{"ID": id}),
			p.T("agent.run.planned_model", map[string]any{"Model": d.Model}),
			p.T("agent.run.planned_effort", map[string]any{"Effort": d.Effort}),
			p.T("agent.run.planned_cwd", map[string]any{"Dir": d.Cwd}),
		} {
			_, _ = fmt.Fprintln(call.Stdout, line)
		}
	}
	return contract.Outcome{Data: d}
}

// writeDryRun prints a dry run as plain text: the invocation as a shell
// reads it back, the settings, and each pending check.
func writeDryRun(w io.Writer, p *i18n.Printer, d dryRun) {
	words := make([]string, len(d.Invocation))
	for i, word := range d.Invocation {
		words[i] = command.Quote(word)
	}
	_, _ = fmt.Fprintln(w, p.T("agent.run.planned", map[string]any{"Version": d.HarnessVersion}))
	_, _ = fmt.Fprintln(w, "  "+strings.Join(words, " "))
	for _, line := range []string{
		p.T("agent.run.planned_model", map[string]any{"Model": d.Model}),
		p.T("agent.run.planned_effort", map[string]any{"Effort": d.Effort}),
		p.T("agent.run.planned_cwd", map[string]any{"Dir": d.Cwd}),
		p.T("agent.run.planned_environment", map[string]any{"Variable": depthVariable, "Value": d.Environment[depthVariable]}),
	} {
		_, _ = fmt.Fprintln(w, line)
	}
	for _, c := range d.Pending {
		_, _ = fmt.Fprintln(w, p.T("agent.run.pending", map[string]any{"ID": c.ID, "Message": c.Message}))
	}
}
