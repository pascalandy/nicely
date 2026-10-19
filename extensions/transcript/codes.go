package transcript

import (
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
)

// downloadFailed is the own code of the manifest for audio that yt-dlp
// could not fetch.
const downloadFailed contract.Code = "TRANSCRIPT_DOWNLOAD_FAILED"

// programCode is how Go reads one program code of a failed step, as the
// table of spec/protocol.md maps it. Its catalog ID is transcript. and the
// program code.
type programCode struct {
	code contract.Code
	// early is true for a code that may arrive before the intent, where it
	// leaves the step pending.
	early bool
	// late is the step status of the code after the ack, or empty for a
	// code that cannot arrive then.
	late record.StepStatus
}

var programCodes = map[string]programCode{
	"download_temporary":    {contract.Temporary, true, ""},
	"download_failed":       {downloadFailed, true, ""},
	"deepgram_unreachable":  {contract.Temporary, false, record.StepFailed},
	"deepgram_key_rejected": {contract.AuthRejected, false, record.StepFailed},
	"deepgram_failed":       {contract.Runtime, false, record.StepFailed},
	"deepgram_unknown":      {contract.Runtime, false, record.StepUnknown},
	"transcript_unusable":   {contract.Runtime, false, record.StepUnknown},
	"write_failed":          {contract.Runtime, true, record.StepUnknown},
}

// programWarnings maps the one program warning to its public code.
var programWarnings = map[string]contract.Code{
	"browser_cookies_skipped": "TRANSCRIPT_BROWSER_COOKIES_SKIPPED",
}

// stepStatus checks a step message against the table. Before the ack, it
// accepts a failure that may arrive early, with the step pending. After the
// ack, it accepts the completed step or a failure with its late status. It
// returns the status that the step takes, and false for a step that the
// protocol does not define.
func stepStatus(m message, acked bool) (record.StepStatus, bool) {
	for _, w := range m.Warnings {
		if _, known := programWarnings[w]; !known {
			return "", false
		}
	}
	if m.Code == nil {
		return record.StepCompleted, acked && m.Status == string(record.StepCompleted)
	}
	c, known := programCodes[*m.Code]
	switch {
	case !known:
		return "", false
	case !acked:
		return record.StepPending, c.early && m.Status == string(record.StepPending)
	}
	return c.late, c.late != "" && m.Status == string(c.late)
}

// itemProblem is the problem of the item for a program code. Before the
// ack, its hint repeats or checks the call. After it, the hint inspects the
// run first, except for a refused key, which a human replaces.
func itemProblem(p *i18n.Printer, code, url, id string, acked bool) *contract.Problem {
	c := programCodes[code]
	data := map[string]any{"URL": url, "ID": id}
	problem := &contract.Problem{Code: c.code, Message: p.T("transcript."+code, data)}
	switch {
	case c.code == contract.AuthRejected:
		login := []string{"ncly", "auth", "login", "deepgram"}
		problem.Hint, problem.Next = strings.Join(login, " "), login
	case acked:
		view := []string{"ncly", "run", "view", id}
		problem.Hint, problem.Next = strings.Join(view, " "), view
	default:
		problem.Hint = p.T("transcript."+code+"_hint", data)
	}
	return problem
}

// warnings are the public warnings of the program's warning codes.
func warnings(p *i18n.Printer, codes []string) []contract.Problem {
	var out []contract.Problem
	for _, w := range codes {
		out = append(out, contract.Problem{Code: programWarnings[w], Message: p.T("transcript." + w), Hint: p.T("transcript." + w + "_hint")})
	}
	return out
}
