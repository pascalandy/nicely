// Package specdoc reads the tables of the north-star docs, so that tests can
// check that the code and the spec agree.
package specdoc

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Tables returns the body rows of every table in the section whose "## "
// heading starts with heading. Cells lose their surrounding spaces.
func Tables(path, heading string) ([][][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tables [][][]string
	inSection, inTable := false, false
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			inSection = strings.HasPrefix(line, "## "+heading)
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
