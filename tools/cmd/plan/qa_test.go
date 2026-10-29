package main

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// qaAxes are the axes of docs/qa/north-star.md, which each UAT judges as a
// criteria family and reports in a section of its own. See #377.
var qaAxes = []string{"Automation", "Economy"}

var reportEntry = regexp.MustCompile(`^\d+\. (\w+)`)

func TestProtocolKeepsQAAxes(t *testing.T) {
	data, err := os.ReadFile("../../../docs/qa/protocol.md")
	if err != nil {
		t.Fatal(err)
	}
	protocol := string(data)
	if missing := missingAxes(protocol); len(missing) > 0 {
		t.Fatalf("docs/qa/protocol.md lost the axes of docs/qa/north-star.md: %v", missing)
	}
	for _, cut := range []struct{ old, new, missing string }{
		{"| Economy |", "| Thrift |", "Criteria families: Economy"},
		{"5. Automation:", "5. Checks:", "Report: Automation"},
	} {
		changed := strings.Replace(protocol, cut.old, cut.new, 1)
		if changed == protocol || !slices.Contains(missingAxes(changed), cut.missing) {
			t.Errorf("a protocol without %q passed", cut.missing)
		}
	}
}

// missingAxes names each axis that the Criteria families table or the
// Report list of the protocol does not hold.
func missingAxes(protocol string) []string {
	found := map[string]bool{}
	var section string
	var header []string
	for line := range strings.Lines(protocol) {
		line = strings.TrimRight(line, "\n")
		if title, isHeading := strings.CutPrefix(line, "## "); isHeading {
			section = title
			continue
		}
		switch section {
		case "Criteria families":
			if cells, isHeader := tableRow(line, &header); len(cells) > 0 && !isHeader {
				found[section+": "+cells[0]] = true
			}
		case "Report":
			if m := reportEntry.FindStringSubmatch(line); m != nil {
				found[section+": "+m[1]] = true
			}
		}
	}
	var missing []string
	for _, section := range []string{"Criteria families", "Report"} {
		for _, axis := range qaAxes {
			if !found[section+": "+axis] {
				missing = append(missing, section+": "+axis)
			}
		}
	}
	return missing
}
