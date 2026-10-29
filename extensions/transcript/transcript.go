// Package transcript is the external extension that transcribes one YouTube
// video with Deepgram, as extensions/transcript/spec.md defines. It embeds
// the Python engine of python/, which it extracts into Nicely's cache and
// runs through uv, and it imports only sdk/. It installs as ncly-transcript
// with its manifest, catalogs, and skill beside it.
package transcript

import (
	"context"
	"embed"
	"errors"
	"fmt"

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

// Extension is the transcript extension, which cmd/ncly-transcript runs.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{
		"transcript run youtube": run,
		"transcript status":      status,
	},
}

// The one step of a run, as Records and replay in spec/records-and-replay.md
// names it.
const stepID = "transcribe"

// defaultTimeout bounds a call without --timeout, as the manifest declares.
const defaultTimeout = "570s"

// errTimedOut is the cause of a context that the timeout of the call ended.
var errTimedOut = errors.New("the call reached its timeout")

// run prepares the call, then answers its dry run or executes it. The
// timeout covers the whole call, from the start of preparation.
func run(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	value, given := call.Flags["timeout"]
	if !given {
		value = defaultTimeout
	}
	limit, valid := command.Duration(value)
	if !valid {
		return refused(usage(p, "transcript.run.timeout_invalid", "transcript.run.timeout_invalid_hint", map[string]any{"Value": i18n.Clip(value)}))
	}
	ctx, cancel := context.WithTimeoutCause(ctx, limit, errTimedOut)
	defer cancel()
	if id, given := call.Flags["run-id"]; given && !record.ValidID(id) {
		return refused(usage(p, "transcript.run.run_id_invalid", "transcript.run.run_id_invalid_hint", map[string]any{"ID": i18n.Clip(id)}))
	}
	job, problem := prepare(call)
	if problem != nil {
		return refused(problem)
	}
	job.timeout = limit
	if call.Flags["dry-run"] != "true" {
		return execute(ctx, call, job, call.Flags["run-id"])
	}
	return dryRun(ctx, call, job)
}

// refused is the outcome of a call that attempts no item: one problem, and
// no results.
func refused(p *contract.Problem) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{*p}}
}

// failed is the outcome of a call that stopped before it attempted the
// item: its problem, or none for a signal, which the SDK answers. Either
// names the run of id, which an earlier attempt recorded, unless id is
// empty.
func failed(p *contract.Problem, id string) contract.Outcome {
	if p != nil {
		return refusal(p, id)
	}
	o := contract.Outcome{}
	if id != "" {
		o.Data = runData{RunID: id}
	}
	return o
}

// usage is a USAGE_INVALID problem whose message and hint come from the
// catalog.
func usage(p *i18n.Printer, message, hint string, data map[string]any) *contract.Problem {
	return &contract.Problem{Code: contract.UsageInvalid, Message: p.T(message, data), Hint: p.T(hint, data)}
}

// dryRunAnswer is the answer of a dry run, as Dry-run answer in the contract
// defines it.
type dryRunAnswer struct {
	DryRun  bool       `json:"dry_run"`
	Plan    []planItem `json:"plan"`
	Pending []check    `json:"pending_checks"`
}

// planItem is the one item of a dry run. Output is the parent that will hold
// the result folder of a new run. OutputDir is the result folder of a
// completed run, which the call reuses.
type planItem struct {
	Key       string          `json:"key"`
	Output    string          `json:"output,omitempty"`
	OutputDir string          `json:"output_dir,omitempty"`
	Steps     []planStep      `json:"steps"`
	Effects   map[string]bool `json:"effects"`
}

type planStep struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// check is a pending check of a dry run: one that only execution makes.
type check struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// dryRun answers what execution would do. A run ID that a record holds
// decides as execution decides, from a read without the lock. A completed
// run plans reuse, and a run that the call may not continue refuses alike.
// Otherwise it resolves the job and asks the engine for its plan, as
// execution does. It reads no key and creates no record, no result folder,
// and no user file. Only the extraction of the engine and the cache of uv
// may write.
func dryRun(ctx context.Context, call extension.Call, j job) contract.Outcome {
	p := call.Printer
	held := ""
	if id, given := call.Flags["run-id"]; given {
		runs := store()
		saved, err := runs.Read(id)
		var bad *record.UnreadableError
		switch {
		case errors.As(err, &bad):
			return refusal(unreadable(p, bad.File, id), id)
		case err != nil:
			return refusal(unreadable(p, runs.File(id), id), id)
		case saved != nil:
			held = id
			completed, problem := decide(p, call, j, saved, runs.File(id), id)
			if problem != nil {
				return refusal(problem, id)
			}
			if completed {
				return reuse(call, saved, runs.File(id), id)
			}
			if problem := restore(p, saved, runs.File(id), id, &j); problem != nil {
				return refusal(problem, id)
			}
		}
	}
	if problem := resolve(ctx, call, &j); problem != nil {
		return refusal(problem, held)
	}
	s, problem := start(ctx, call, j, true, nil)
	if problem != nil {
		return refusal(problem, held)
	}
	defer s.close()
	problem = s.plan(ctx, p, j)
	if problem == nil && ctx.Err() == nil {
		problem = s.finish(ctx, p, j)
	}
	if problem != nil || ctx.Err() != nil {
		return failed(problem, held)
	}
	d := dryRunAnswer{
		DryRun: true,
		Plan: []planItem{{
			Key:     j.url,
			Output:  j.output,
			Steps:   []planStep{{ID: stepID, Action: "run"}},
			Effects: map[string]bool{"user_writes": true, "network": true, "paid": true},
		}},
		Pending: []check{{ID: "auth.deepgram", Message: p.T("transcript.run.pending_auth")}},
	}
	if !call.JSON {
		_, _ = fmt.Fprintln(call.Stdout, p.T("transcript.run.planned", map[string]any{"URL": j.url, "Output": j.output}))
		for _, c := range d.Pending {
			_, _ = fmt.Fprintln(call.Stdout, p.T("transcript.run.pending", map[string]any{"ID": c.ID, "Message": c.Message}))
		}
	}
	return contract.Outcome{Data: d}
}
