package cli

import (
	"os"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/runner"
)

// programs checks each program that the manifest of an extension requires
// and does not mark conditional, in the order of their names. A broken
// manifest declares nothing that doctor can read.
func (pr *prober) programs(e *extension) []check {
	if e.broken != nil {
		return nil
	}
	var checks []check
	env := runner.Environ(os.Environ(), nil)
	for _, name := range e.manifest.Programs {
		r := e.manifest.Requirements[name]
		if r.Conditional || pr.signal != nil {
			continue
		}
		c := check{ID: e.slug + "." + name, Component: e.slug}
		data := map[string]any{"Name": name, "Version": r.Version, "Probe": strings.Join(append([]string{name}, r.ProbeArgs()...), " ")}
		timedOut := func(message string) {
			c.Status, c.Code, c.Message, c.Hint = statusFail, contract.CheckTimeout, message, pr.p.T("doctor.timeout_hint")
		}
		left := time.Until(pr.deadline)
		if pr.spent || left <= 0 {
			pr.spent = true
			timedOut(pr.p.T("doctor.not_started", data))
			checks = append(checks, c)
			continue
		}
		found, err := runner.CheckProgram(pr.ctx, name, r, env, left)
		data["Path"], data["Found"], data["Exit"] = found.Path, found.Version, found.Exit
		switch {
		case err != nil:
			found.Verdict = runner.ProgramProbeFailed
			data["Exit"] = -1
			found.Install = nil
		case found.Signal != nil:
			pr.signal = found.Signal
			return checks
		}
		c.Status = statusFail
		c.Code = contract.PrereqMissing
		switch found.Verdict {
		case runner.ProgramFound:
			c.Status, c.Code, c.Message = statusPass, "", pr.p.T("doctor.program_found", data)
			if found.Version != "" {
				c.Message = pr.p.T("doctor.program_version", data)
			}
		case runner.ProgramTimedOut:
			pr.spent = true
			timedOut(pr.p.T("doctor.program_timeout", data))
		case runner.ProgramMissing:
			c.Message = pr.p.T("doctor.program_missing", data)
		case runner.ProgramTooOld:
			c.Message = pr.p.T("doctor.program_old", data)
		case runner.ProgramProbeFailed:
			c.Message = pr.p.T("doctor.program_probe_failed", data)
		case runner.ProgramNoVersion:
			c.Message = pr.p.T("doctor.program_no_version", data)
		}
		if found.Failed() {
			c.Hint = pr.p.T("doctor.program_install", data)
			if found.Install != nil {
				c.Hint, c.Next = strings.Join(found.Install, " "), found.Install
			}
		}
		checks = append(checks, c)
	}
	return checks
}
