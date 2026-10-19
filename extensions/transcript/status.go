package transcript

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
)

// finding is one check of ncly transcript status, the check object of ncly
// doctor in core-spec.md.
type finding struct {
	ID        string        `json:"id"`
	Component string        `json:"component"`
	Status    string        `json:"status"`
	Code      contract.Code `json:"code,omitempty"`
	Message   string        `json:"message"`
	Hint      string        `json:"hint,omitempty"`
}

// report is the answer of ncly transcript status.
type report struct {
	Checks []finding `json:"checks"`
}

// arcProfile is Arc's default profile, from the home folder, where
// youtube.py reads the cookies of browser = "arc".
var arcProfile = filepath.Join("Library", "Application Support", "Arc", "User Data", "Default")

// cookieFiles are the places of the cookie database in a Chromium profile,
// the current one first.
var cookieFiles = []string{filepath.Join("Network", "Cookies"), "Cookies"}

// status checks the cookies of the configured browser, which a run needs
// beside the programs that doctor probes. ncly transcript status in spec.md
// defines it. It reads no key and starts no program. It fails only for a
// broken [transcript] table, as preparation does.
func status(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	home, _ := os.UserHomeDir()
	var j job
	if _, problem := readSettings(p, home, config.Locate(home, os.LookupEnv), &j); problem != nil {
		return refused(problem)
	}
	checks := []finding{cookies(p, home, j.browser)}
	if !call.JSON {
		writeChecks(call.Stdout, p, checks)
	}
	return contract.Outcome{Data: report{Checks: checks}}
}

// cookies is the transcript.cookies check. It passes without a browser, and
// with arc when Go can open a cookie file of Arc's default profile.
// Otherwise it gives the warning of a run that skipped the cookies.
func cookies(p *i18n.Printer, home, browser string) finding {
	c := finding{ID: "transcript.cookies", Component: "transcript", Status: "pass"}
	if browser == "" {
		c.Message = p.T("transcript.cookies_anonymous")
		return c
	}
	profile := filepath.Join(home, arcProfile)
	for _, name := range cookieFiles {
		if path := filepath.Join(profile, name); readable(path) {
			c.Message = p.T("transcript.cookies_readable", map[string]any{"File": path})
			return c
		}
	}
	c.Status, c.Code = "warn", programWarnings["browser_cookies_skipped"]
	c.Message = p.T("transcript.cookies_unreadable", map[string]any{"Profile": profile})
	c.Hint = p.T("transcript.browser_cookies_skipped_hint")
	return c
}

// readable reports whether path is a regular file that Go can open for
// reading. It checks the type first, since opening a pipe blocks.
func readable(path string) bool {
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// writeChecks prints the report as ncly doctor does, in the look of sdk/tui.
func writeChecks(w io.Writer, p *i18n.Printer, checks []finding) {
	labels := map[string]string{"pass": p.T("transcript.status_pass"), "warn": p.T("transcript.status_warn")}
	rows := make([]tui.Check, len(checks))
	for i, c := range checks {
		rows[i] = tui.Check{Status: c.Status, Label: labels[c.Status], ID: c.ID, Message: c.Message, Code: string(c.Code), Hint: c.Hint}
	}
	// Core passes the resolved --no-color as NO_COLOR, as Global flags in
	// contract.md says.
	_, _ = io.WriteString(tui.Output(w, os.Getenv("NO_COLOR") != ""), tui.Checks(rows, p.T("error.hint")))
}
