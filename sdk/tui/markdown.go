package tui

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
)

// Markdown renders a skill's Markdown in the look of ncly skill view: the
// style of the help, with headings in bold uppercase, code in cyan, and
// dashes for bullets, wrapped at width. Without color, code keeps its
// backticks, so it stays apart from the prose. Glamour writes the styles,
// and Output turns them into plain text when color is off.
func Markdown(markdown string, width int, noColor bool) (string, error) {
	r, err := glamour.NewTermRenderer(glamour.WithStyles(markdownStyle(noColor)), glamour.WithWordWrap(width))
	if err != nil {
		return "", err
	}
	text, err := r.Render(markdown)
	if err != nil {
		return "", err
	}
	// Glamour pads each line to the width, which a copy would keep.
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n"), nil
}

// markdownStyle is the Glamour style of Markdown, with the 16 ANSI colors
// only.
func markdownStyle(noColor bool) ansi.StyleConfig {
	on, off, indent, quote, cyan := true, false, uint(2), uint(1), "6"
	code := ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &cyan, Upper: &off}}
	if noColor {
		code = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "`", Suffix: "`", Upper: &off}}
	}
	heading := ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{BlockSuffix: "\n", Bold: &on, Upper: &on}}
	return ansi.StyleConfig{
		Document:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{BlockSuffix: "\n"}},
		Paragraph:      ansi.StyleBlock{Margin: &indent},
		List:           ansi.StyleList{StyleBlock: ansi.StyleBlock{Margin: &indent}, LevelIndent: 2},
		Heading:        heading,
		Item:           ansi.StylePrimitive{BlockPrefix: "- "},
		Enumeration:    ansi.StylePrimitive{BlockPrefix: ". "},
		Code:           code,
		CodeBlock:      ansi.StyleCodeBlock{StyleBlock: ansi.StyleBlock{Margin: &indent}},
		BlockQuote:     ansi.StyleBlock{Indent: &quote, IndentToken: &[]string{"| "}[0]},
		HorizontalRule: ansi.StylePrimitive{Format: "\n--------\n"},
		Strong:         ansi.StylePrimitive{Bold: &on},
		Emph:           ansi.StylePrimitive{Italic: &on},
		Task:           ansi.StyleTask{Ticked: "[x] ", Unticked: "[ ] "},
		Table:          ansi.StyleTable{CenterSeparator: &[]string{"|"}[0], ColumnSeparator: &[]string{"|"}[0], RowSeparator: &[]string{"-"}[0]},
	}
}
