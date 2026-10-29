package skill

import (
	"io"
	"strings"

	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
	"golang.org/x/term"
)

// write prints one line per skill, as the help prints its commands: its
// name, in bold, and a colon, then its description on one line. The SDK
// reports a failed write.
func write(w io.Writer, skills []Skill) {
	entries := make([]tui.Entry, len(skills))
	for i, s := range skills {
		entries[i] = tui.Entry{Name: s.Name, Summary: strings.Join(strings.Fields(s.Description), " ")}
	}
	_, _ = io.WriteString(tui.Output(w, noColor()), tui.Names(entries))
}

// render prints a skill in interactive mode: its folder as a first line,
// then its Markdown without the frontmatter, which Glamour would draw as a
// rule and a heading, in the look of sdk/tui, wrapped at the width of the
// terminal. The SDK reports a failed write.
func render(w io.Writer, p *i18n.Printer, s Skill) {
	width := 80
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		if columns, _, err := term.GetSize(int(f.Fd())); err == nil && columns > 0 {
			width = columns
		}
	}
	text, err := tui.Markdown(body(s.content), width, noColor())
	if err != nil {
		text = s.content
	}
	out := tui.Output(w, noColor())
	if s.Path != nil {
		_, _ = io.WriteString(out, tui.Muted(p.T("skill.folder"))+"  "+*s.Path+"\n\n")
	}
	_, _ = io.WriteString(out, text)
}

// body returns a SKILL.md without its frontmatter, which gather has checked.
func body(content string) string {
	lines := strings.SplitAfter(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r\n") == "---" {
			return strings.Join(lines[i+1:], "")
		}
	}
	return content
}
