// Package tui holds the look that every command shares: the styles, the
// help page, the error block, the masked field of a key, and the
// confirmation. A spinner or a progress bar joins it with the first command
// that needs one.
//
// Build every interactive screen from this package, so every command looks
// the same. A component joins it after Pascal picks its look from a published
// mockup, as dev-preferences.md asks, and the spec of each command holds the
// look that Pascal picked in M07. A command whose spec gives no look reuses
// the help and error styles of M00 and prints the rest as plain text, so it
// waits on no mockup.
package tui

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/sdk/contract"
	"golang.org/x/term"
)

// The styles follow variant D of the T4 mockups: the quiet help of variant
// A, with bold uppercase headings, and the errors of variant B, which lead
// with their code. They are plain SGR sequences, which Output turns into
// plain text when color is off.
var (
	heading      = ansi.Style{}.Bold()
	errorLabel   = ansi.Style{}.Bold().ForegroundColor(ansi.Red)
	warningLabel = ansi.Style{}.Bold().ForegroundColor(ansi.Yellow)
	hintLabel    = ansi.Style{}.Bold().ForegroundColor(ansi.Cyan)
	muted        = ansi.Style{}.ForegroundColor(ansi.BrightBlack)
)

// Bold, ErrorLabel, HintLabel, and Muted style a word of a report as the help
// and the error block do: a name in bold, a problem's label in bold red,
// hint: in bold cyan, and a label of lesser weight in bright black.
func Bold(s string) string       { return heading.Styled(s) }
func ErrorLabel(s string) string { return errorLabel.Styled(s) }
func HintLabel(s string) string  { return hintLabel.Styled(s) }
func Muted(s string) string      { return muted.Styled(s) }

// Output wraps a stream so that styles become plain text when color is off.
// noColor turns color off even with CLICOLOR_FORCE. Otherwise a stream that
// is not a terminal gets plain text unless CLICOLOR_FORCE asks for color.
// The decision reads only the environment and the stream, and never queries
// the terminal.
func Output(w io.Writer, noColor bool) io.Writer {
	return &colorprofile.Writer{Forward: w, Profile: Profile(w, noColor)}
}

// Profile returns the color profile of styles written to w, by the rule of
// Output: the 16 ANSI colors, or plain text when color is off. A form takes
// this profile, so Bubble Tea never detects one, which queries tmux.
func Profile(w io.Writer, noColor bool) colorprofile.Profile {
	force := os.Getenv("CLICOLOR_FORCE")
	if !noColor && (isTerminal(w) || (force != "" && force != "0")) {
		return colorprofile.ANSI
	}
	return colorprofile.NoTTY
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
}

// Entry is one line of a list: a command or a flag, then its summary.
type Entry struct {
	Name    string
	Summary string
}

// Section is a titled list of entries.
type Section struct {
	Title   string
	Entries []Entry
	// Colon follows each name with a colon, as for commands.
	Colon bool
}

// HelpPage is a help page, already in the active language.
type HelpPage struct {
	Summary string
	// Note is one plain line under the summary, such as the effects of a
	// command, or empty.
	Note          string
	UsageTitle    string
	Usage         string
	Sections      []Section
	ExamplesTitle string
	Examples      []string
}

// Help renders a help page, which ends with its examples. Columns take the
// width of the translated text.
func Help(p HelpPage) string {
	var b strings.Builder
	b.WriteString(p.Summary + "\n")
	if p.Note != "" {
		b.WriteString("\n" + p.Note + "\n")
	}
	block := func(title string, lines []string) {
		if len(lines) == 0 {
			return
		}
		b.WriteString("\n" + heading.Styled(strings.ToUpper(title)) + "\n")
		for _, line := range lines {
			b.WriteString("  " + line + "\n")
		}
	}
	block(p.UsageTitle, []string{p.Usage})
	for _, s := range p.Sections {
		block(s.Title, columns(s.Entries, s.Colon, ansi.Style{}))
	}
	var examples []string
	for _, e := range p.Examples {
		examples = append(examples, "$ "+e)
	}
	block(p.ExamplesTitle, examples)
	return b.String()
}

// Names renders entries as the help renders its commands: each name and a
// colon, in bold, then its summary, aligned after the widest name, so the
// terminal wraps a long summary.
func Names(entries []Entry) string {
	return strings.Join(columns(entries, true, heading), "\n") + "\n"
}

func columns(entries []Entry, colon bool, style ansi.Style) []string {
	names := make([]string, len(entries))
	width := 0
	for i, e := range entries {
		names[i] = e.Name
		if colon {
			names[i] += ":"
		}
		width = max(width, ansi.StringWidth(names[i]))
	}
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = style.Styled(names[i]) + strings.Repeat(" ", width-ansi.StringWidth(names[i])+2) + e.Summary
	}
	return lines
}

// Check is one finding of a report, such as a check of ncly doctor, with its
// text in the active language.
type Check struct {
	// Status is pass, warn, or fail, and Label its word in the active
	// language.
	Status, Label string
	ID, Message   string
	// Code and Hint stay empty for a check that passes.
	Code, Hint string
}

// The status styles of a check, as ncly doctor in core-spec.md describes
// them.
var statusColors = map[string]ansi.BasicColor{"pass": ansi.Green, "warn": ansi.Yellow, "fail": ansi.Red}

// Checks renders a report in three aligned columns: each check's status, in
// bold and in its color, its id, and its message. A check with a problem adds
// a line under it, with its code, in the color of its status, in the column
// of the id, then the hint label, such as hint:, and its hint in the column
// of the message.
func Checks(checks []Check, hint string) string {
	status, id := 0, 0
	for _, c := range checks {
		status = max(status, ansi.StringWidth(c.Label))
		id = max(id, ansi.StringWidth(c.ID), ansi.StringWidth(c.Code))
	}
	pad := func(s string, width int) string { return s + strings.Repeat(" ", width-ansi.StringWidth(s)+2) }
	var b strings.Builder
	for _, c := range checks {
		color := statusColors[c.Status]
		b.WriteString(ansi.Style{}.Bold().ForegroundColor(color).Styled(c.Label) + strings.Repeat(" ", status-ansi.StringWidth(c.Label)+2) + pad(c.ID, id) + c.Message + "\n")
		if c.Hint != "" {
			b.WriteString(strings.Repeat(" ", status+2) + ansi.Style{}.ForegroundColor(color).Styled(c.Code) + strings.Repeat(" ", id-ansi.StringWidth(c.Code)+2) + hintLabel.Styled(hint) + " " + c.Hint + "\n")
		}
	}
	return b.String()
}

// Problems renders errors, then warnings. Each one leads with a label that
// carries its code, such as error[USAGE_INVALID]:, then shows its hint after
// the hint label. label returns the translated label of a code.
func Problems(errs, warnings []contract.Problem, label func(code contract.Code, warning bool) string, hint string) string {
	var blocks []string
	add := func(style ansi.Style, p contract.Problem, warning bool) {
		blocks = append(blocks, style.Styled(label(p.Code, warning))+" "+p.Message+"\n\n  "+hintLabel.Styled(hint)+" "+p.Hint+"\n")
	}
	for _, p := range errs {
		add(errorLabel, p, false)
	}
	for _, p := range warnings {
		add(warningLabel, p, true)
	}
	return strings.Join(blocks, "\n")
}
