package main

import "strings"

// tableRow returns the cells of a table row, or nil for another line, which
// ends the table. The first row of a table is its header, which it stores in
// header; the delimiter row returns no cells.
func tableRow(line string, header *[]string) (cells []string, isHeader bool) {
	if !strings.HasPrefix(strings.TrimSpace(line), "|") {
		*header = nil
		return nil, false
	}
	cells = tableCells(line)
	switch {
	case *header == nil:
		*header = cells
		return cells, true
	case strings.HasPrefix(strings.TrimSpace(line), "|---"):
		return nil, false
	}
	return cells, false
}

// tableCells splits a table row into its cells, without their spaces.
func tableCells(line string) []string {
	var cells []string
	for cell := range strings.SplitSeq(strings.Trim(strings.TrimSpace(line), "|"), "|") {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

// cellCodes returns the code spans of a cell, or the cell when it has none.
func cellCodes(cell string) []string {
	spans := codeSpan.FindAllStringSubmatch(cell, -1)
	if len(spans) == 0 {
		return []string{cell}
	}
	var values []string
	for _, s := range spans {
		values = append(values, s[1])
	}
	return values
}

// fence follows the fenced code blocks of a Markdown file, indented or not. A
// block opens with three or more backticks or tildes, and only a line of at
// least as many of the same character, alone, closes it.
type fence struct {
	open string // the marker of the current block, or "" outside one
}

// in reports whether line opens, closes, or lies inside a fenced block.
func (f *fence) in(line string) bool {
	text := strings.TrimLeft(line, " ")
	marker := ""
	if strings.HasPrefix(text, "```") || strings.HasPrefix(text, "~~~") {
		marker = text[:len(text)-len(strings.TrimLeft(text, text[:1]))]
	}
	switch {
	case f.open == "":
		f.open = marker
		return marker != ""
	case marker != "" && marker[0] == f.open[0] && len(marker) >= len(f.open) && strings.TrimSpace(text[len(marker):]) == "":
		f.open = ""
	}
	return true
}
