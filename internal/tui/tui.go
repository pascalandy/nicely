// Package tui holds the look that every command shares: the styles, the
// help page, and the error block. A form, a spinner, or a progress bar
// joins it with the first command that needs one.
package tui

import (
	"io"
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/pascalandy/nicely/internal/contract"
)

// Styles are the styles of variant A in the T4 mockups: bold uppercase
// headings, a red mark on errors, and dim labels.
var (
	heading     = lipgloss.NewStyle().Bold(true)
	errorMark   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red)
	warningMark = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Yellow)
	dim         = lipgloss.NewStyle().Faint(true)
)

// Output wraps a stream so that styles become plain text when color is off.
// noColor holds the --no-color flag when the user gave it, and it wins.
// Otherwise NO_COLOR or TERM=dumb turns color off, even with CLICOLOR_FORCE,
// and a stream that is not a terminal gets plain text unless CLICOLOR_FORCE
// asks for color.
func Output(w io.Writer, noColor *bool) io.Writer {
	environ := os.Environ()
	plain := os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"
	if noColor != nil {
		plain = *noColor
		environ = slices.DeleteFunc(environ, func(v string) bool { return strings.HasPrefix(v, "NO_COLOR=") })
	}
	out := colorprofile.NewWriter(w, environ)
	if plain {
		out.Profile = colorprofile.NoTTY
	}
	return out
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
		b.WriteString("\n" + heading.Render(strings.ToUpper(title)) + "\n")
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
		width = max(width, lipgloss.Width(names[i]))
	}
	lines := make([]string, len(s.Entries))
	for i, e := range s.Entries {
		lines[i] = names[i] + strings.Repeat(" ", width-lipgloss.Width(names[i])+2) + e.Summary
	}
	return lines
}

// Problems renders errors, then warnings, with the hint after a label such
// as "Try:" and the code dimmed at the end of the hint line.
func Problems(errs, warnings []contract.Problem, try string) string {
	var b strings.Builder
	write := func(mark string, p contract.Problem) {
		b.WriteString(mark + " " + p.Message + "\n")
		b.WriteString("  " + dim.Render(try) + " " + p.Hint + "  " + dim.Render(string(p.Code)) + "\n")
	}
	for _, p := range errs {
		write(errorMark.Render("✗"), p)
	}
	for _, p := range warnings {
		write(warningMark.Render("!"), p)
	}
	return b.String()
}
