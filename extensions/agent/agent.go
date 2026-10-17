// Package agent is the external extension that runs one trusted task
// through Codex, as extensions/agent/spec.md defines. It imports only sdk/,
// and installs as ncly-agent with its manifest, catalogs, and skill beside it.
package agent

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
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
	Key     string          `json:"key"`
	Output  string          `json:"output"`
	Steps   []step          `json:"steps"`
	Effects map[string]bool `json:"effects"`
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

// run prepares the task. A dry run shows what execution would run, and
// starts no task. This build runs no task yet: M10 adds execution.
func run(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	t, problem := prepare(ctx, call)
	if ctx.Err() != nil {
		// A signal stopped preparation, and the SDK answers it alone.
		return contract.Outcome{}
	}
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	if call.Flags["dry-run"] != "true" {
		return contract.Outcome{Errors: []contract.Problem{{
			Code:    contract.CapabilityUnsupported,
			Message: p.T("agent.run.execution_unavailable"),
			Hint:    p.T("agent.run.execution_unavailable_hint"),
		}}}
	}
	home, _ := os.UserHomeDir()
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
			Key:     "task",
			Output:  filepath.Join(config.Locate(home, os.LookupEnv).State, "runs"),
			Steps:   []step{{ID: "codex", Action: "run"}},
			Effects: map[string]bool{"user_writes": true, "network": true, "paid": true},
		}},
		Pending: []check{{ID: "agent.login", Message: p.T("agent.run.pending_login")}},
	}
	if !call.JSON {
		writeDryRun(call.Stdout, p, d)
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
