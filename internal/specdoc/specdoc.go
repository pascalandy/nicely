// Package specdoc reads the tables of the north-star docs, so that tests can
// check that the code and the spec agree.
package specdoc

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// title returns the text of a "## " heading, or of the "# " heading on the
// first line. Either starts a section, because a part of a split doc opens
// with its section as "# ". A "# " line further down, such as a shell comment
// in a code block, starts nothing.
func title(line string, first bool) (string, bool) {
	if text, ok := strings.CutPrefix(line, "## "); ok {
		return text, true
	}
	if !first {
		return "", false
	}
	return strings.CutPrefix(line, "# ")
}

// Tables returns the body rows of every table in the section whose title
// starts with heading. Cells lose their surrounding spaces.
func Tables(path, heading string) ([][][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tables [][][]string
	inSection, inTable, first := false, false, true
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		text, ok := title(line, first)
		first = false
		if ok {
			inSection = strings.HasPrefix(text, heading)
		}
		isRow := inSection && strings.HasPrefix(line, "|")
		switch {
		case isRow && !inTable:
			tables = append(tables, nil)
		case isRow && !strings.HasPrefix(line, "|---"):
			var cells []string
			for cell := range strings.SplitSeq(strings.Trim(line, "|"), "|") {
				cells = append(cells, strings.TrimSpace(cell))
			}
			tables[len(tables)-1] = append(tables[len(tables)-1], cells)
		}
		inTable = isRow
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("%s: no table under %q", path, heading)
	}
	return tables, nil
}

var code = regexp.MustCompile("`([^`]+)`")

// Backticked returns the code spans of a cell, such as -h and --help in
// "`-h`, `--help`".
func Backticked(cell string) []string {
	var spans []string
	for _, m := range code.FindAllStringSubmatch(cell, -1) {
		spans = append(spans, m[1])
	}
	return spans
}

// Block returns the lines of the first fenced code block in the section
// whose title starts with heading.
func Block(path, heading string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	inSection, inBlock, first := false, false, true
	for line := range strings.Lines(string(data)) {
		line = strings.TrimRight(line, "\n")
		text, isTitle := title(line, first)
		first = false
		switch {
		case isTitle:
			inSection = strings.HasPrefix(text, heading)
		case inSection && strings.HasPrefix(line, "```"):
			if inBlock {
				return lines, nil
			}
			inBlock = true
		case inBlock:
			lines = append(lines, line)
		}
	}
	return nil, fmt.Errorf("%s: no code block under %q", path, heading)
}
