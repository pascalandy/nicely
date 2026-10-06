// Package tui holds the look that every command shares: the styles, the
// help page, and the error block. A form, a spinner, or a progress bar
// joins it with the first command that needs one.
package tui

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/internal/contract"
	"golang.org/x/term"
)

// The styles follow variant D of the T4 mockups: the quiet help of variant
// A, with bold uppercase headings, and the errors of variant B, which lead
// with their code. They are plain SGR sequences rather than Lip Gloss,
// whose package detects the terminal as it loads and can wait on
// `tmux info` without a time limit.
var (
	heading      = ansi.Style{}.Bold()
	errorLabel   = ansi.Style{}.Bold().ForegroundColor(ansi.Red)
	warningLabel = ansi.Style{}.Bold().ForegroundColor(ansi.Yellow)
	hintLabel    = ansi.Style{}.Bold().ForegroundColor(ansi.Cyan)
)

// Output wraps a stream so that styles become plain text when color is off.
// noColor turns color off even with CLICOLOR_FORCE. Otherwise a stream that
// is not a terminal gets plain text unless CLICOLOR_FORCE asks for color.
// The decision reads only the environment and the stream, and never queries
// the terminal.
func Output(w io.Writer, noColor bool) io.Writer {
	force := os.Getenv("CLICOLOR_FORCE")
	profile := colorprofile.NoTTY
	if !noColor && (isTerminal(w) || (force != "" && force != "0")) {
		profile = colorprofile.ANSI
	}
	return &colorprofile.Writer{Forward: w, Profile: profile}
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
	Summary       string
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
		block(s.Title, columns(s))
	}
	var examples []string
	for _, e := range p.Examples {
		examples = append(examples, "$ "+e)
	}
	block(p.ExamplesTitle, examples)
	return b.String()
}

func columns(s Section) []string {
	names := make([]string, len(s.Entries))
	width := 0
	for i, e := range s.Entries {
		names[i] = e.Name
		if s.Colon {
			names[i] += ":"
		}
		width = max(width, ansi.StringWidth(names[i]))
	}
	lines := make([]string, len(s.Entries))
	for i, e := range s.Entries {
		lines[i] = names[i] + strings.Repeat(" ", width-ansi.StringWidth(names[i])+2) + e.Summary
	}
	return lines
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
