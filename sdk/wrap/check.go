package wrap

import (
	"cmp"
	"context"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// check checks the program with the rule of Required programs in
// core-spec.md. It returns what the check found, and the PREREQ_MISSING
// problem that stops the call, whose hint and next hold the install command
// of this machine. A signal and a found program give no problem. The
// caller decides if a probe that timed out used up the deadline of Timeout.
func check(ctx context.Context, printer *i18n.Printer, name string, r contract.Requirement, env []string, limit time.Duration) (runner.ProgramCheck, *contract.Problem) {
	found, err := runner.CheckProgram(ctx, name, r, env, limit)
	data := map[string]any{"Name": name, "Found": found.Version, "Version": r.Version, "Exit": found.Exit, "Probe": strings.Join(append([]string{name}, r.ProbeArgs()...), " ")}
	var message, hint string
	switch {
	case err != nil:
		data["Exit"] = -1
		message = printer.T("wrap.probe_failed", data)
	case found.Signal != nil, found.Verdict == runner.ProgramFound:
		return found, nil
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
	return found, &problem
}
