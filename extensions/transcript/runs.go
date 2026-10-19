package transcript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
)

// runData holds the key of an answer that concerns a run: its ID.
type runData struct {
	RunID string `json:"run_id"`
}

// itemData is the one item of an attempted or replayed transcription, as
// ncly transcript in spec.md defines it. OutputDir and Transcript stay null
// until they exist.
type itemData struct {
	URL        string  `json:"url"`
	OutputDir  *string `json:"output_dir"`
	Transcript *string `json:"transcript"`
}

// evidence is what the step transcribe keeps beside its artifacts, so a
// replay answers the warnings of the first answer.
type evidence struct {
	FormatVersion int      `json:"format_version"`
	Warnings      []string `json:"warnings"`
}

// store is the runs folder of the state folder.
func store() record.Store {
	home, _ := os.UserHomeDir()
	return record.Store{Dir: record.Folder(config.Locate(home, os.LookupEnv).State)}
}

// inputs are what the call states, as Records and replay in
// spec/records-and-replay.md lists them. They hold the source, and the
// output parent only when --output gives it. A change of the config
// therefore never changes what the call states.
func inputs(j job) map[string]any {
	var output any
	if j.given {
		output = j.output
	}
	return map[string]any{"source": "youtube", "output": output}
}

// resolved are the values that preparation selected, which a repeated call
// never compares.
func resolved(j job) map[string]any {
	var browser any
	if j.browser != "" {
		browser = j.browser
	}
	return map[string]any{"output": j.output, "browser": browser}
}

// restore gives the job the values that the record of a pending run
// resolved, since a new attempt reads them there instead of picking again.
// An output parent that no longer exists stops the attempt before any
// effect, so a human can restore it.
func restore(p *i18n.Printer, saved *record.Record, file, id string, j *job) *contract.Problem {
	output, ok := saved.Resolved["output"].(string)
	browser, named := saved.Resolved["browser"].(string)
	if !ok || !filepath.IsAbs(output) || named && browser != "arc" || !named && saved.Resolved["browser"] != nil {
		return unreadable(p, file, id)
	}
	if _, err := folder(output); err != nil {
		return inspect(p, contract.ResumeUnsafe, id, "transcript.run.output_gone", map[string]any{"ID": id, "Dir": output})
	}
	j.output, j.browser, j.recorded = output, browser, true
	return nil
}

// decide applies Records and replay to the record that holds the ID of the
// call. It returns the problem of a run that the call may not continue. It
// returns true for a completed run, which the call replays. It returns
// false for a run whose step is still pending, which runs again.
func decide(p *i18n.Printer, call extension.Call, j job, saved *record.Record, file, id string) (bool, *contract.Problem) {
	data := map[string]any{"ID": id}
	if len(saved.Items) != 1 || len(saved.Items[0].Steps) != 1 || saved.Items[0].Steps[0].ID != stepID {
		return false, unreadable(p, file, id)
	}
	if saved.Path != call.Path || saved.Items[0].Key != j.url || !sameValues(saved.Inputs, inputs(j)) {
		return false, inspect(p, contract.UsageInvalid, id, "transcript.run.other_call", data)
	}
	switch step := saved.Items[0].Steps[0]; {
	case saved.Status == record.Completed && step.Status == record.StepCompleted:
		return true, nil
	case step.Status == record.StepPending:
		return false, nil
	}
	return false, inspect(p, contract.ResumeUnsafe, id, "transcript.run.effect_unknown", data)
}

// sameValues reports whether two sets of inputs hold the same values, as
// JSON compares them.
func sameValues(a, b map[string]any) bool {
	x, errA := json.Marshal(a)
	y, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(x, y)
}

// replay answers a completed run from its record, once each artifact still
// holds what the record says, without the engine or Deepgram. A changed or
// missing file is RESUME_UNSAFE, and evidence that this build cannot read
// counts as an unreadable record.
func replay(call extension.Call, saved *record.Record, file, id string) contract.Outcome {
	p := call.Printer
	item := saved.Items[0]
	step := item.Steps[0]
	var ev evidence
	if item.OutputDir == nil || len(step.Artifacts) != len(artifactNames) || json.Unmarshal(step.Evidence, &ev) != nil || ev.FormatVersion != 1 {
		return refusal(unreadable(p, file, id), id)
	}
	for i, a := range step.Artifacts {
		if !intact(a, *item.OutputDir, artifactNames[i]) {
			return refusal(inspect(p, contract.ResumeUnsafe, id, "transcript.run.artifact_changed", map[string]any{"File": a.Path, "ID": id}), id)
		}
	}
	return success(call, id, item.Key, *item.OutputDir, ev.Warnings)
}

// success is the answer of a completed transcription: its run, and the one
// item with its result folder and transcript.
func success(call extension.Call, id, url, dir string, codes []string) contract.Outcome {
	p := call.Printer
	text := filepath.Join(dir, artifactNames[0])
	if !call.JSON {
		_, _ = fmt.Fprintln(call.Stdout, p.T("transcript.run.done", map[string]any{"URL": url, "Dir": dir}))
	}
	return contract.Outcome{
		Data:     runData{RunID: id},
		Results:  []contract.Item{{Data: itemData{URL: url, OutputDir: &dir, Transcript: &text}}},
		Warnings: warnings(p, codes),
	}
}

// refusal is the outcome of a call that attempts no item: one problem, and
// the run ID when the refusal concerns a run.
func refusal(p *contract.Problem, id string) contract.Outcome {
	o := refused(p)
	if id != "" {
		o.Data = runData{RunID: id}
	}
	return o
}

// inspect is a problem whose hint and next inspect the run before anything
// else.
func inspect(p *i18n.Printer, code contract.Code, id, message string, data map[string]any) *contract.Problem {
	view := []string{"ncly", "run", "view", id}
	return &contract.Problem{Code: code, Message: p.T(message, data), Hint: strings.Join(view, " "), Next: view}
}

// unreadable is the RESUME_UNSAFE problem of a record that the call cannot
// use, which never counts as an absent one.
func unreadable(p *i18n.Printer, file, id string) *contract.Problem {
	return inspect(p, contract.ResumeUnsafe, id, "transcript.run.record_unreadable", map[string]any{"File": file})
}

// unsaved is the RUNTIME problem of a record that cannot be written.
func unsaved(p *i18n.Printer, dir string) *contract.Problem {
	data := map[string]any{"Dir": dir}
	return &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.record_failed", data), Hint: p.T("transcript.run.record_failed_hint", data)}
}

// nclyVersion is the release that writes the record: the one that core
// passes, or the build of this extension when it runs alone.
func nclyVersion() string {
	if v := os.Getenv("NCLY_VERSION"); v != "" {
		return v
	}
	return command.ModuleVersion()
}
