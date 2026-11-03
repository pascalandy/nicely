package main

import (
	"fmt"
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

// TestLedgerCorrections checks the Corrections table of the QA ledger, which
// measure 2 of D068 reads: each correction names its level and its owner.
func TestLedgerCorrections(t *testing.T) {
	data, err := os.ReadFile("../../../docs/qa/ledger.md")
	if err != nil {
		t.Fatal(err)
	}
	ledger := string(data)
	if problems := correctionProblems(ledger); len(problems) > 0 {
		t.Fatalf("docs/qa/ledger.md: %v", problems)
	}
	row := "| 2 | A |\n"
	for _, cut := range []struct{ old, new, want string }{
		{row, "| 5 | A |\n", "level 5"},
		{row, "| 2 | D |\n", "owner D"},
		{row, "|  | A |\n", "level "},
		{"| Correction | Failure | Level | Owner |\n|---|---|---|---|\n", "| Correction | Failure | Owner |\n|---|---|---|\n", "no Level and Owner"},
		{"## Corrections\n", "## Fixes\n", "no Corrections"},
	} {
		changed := strings.Replace(ledger, cut.old, cut.new, 1)
		problems := correctionProblems(changed)
		if changed == ledger || !slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, cut.want) }) {
			t.Errorf("a ledger with %q passed, or gave %q", cut.new, problems)
		}
	}
}

// correctionProblems returns what is wrong with the Corrections table of the
// ledger: a missing section or column, or a level or an owner outside D068.
func correctionProblems(ledger string) []string {
	var problems []string
	var section string
	var header []string
	found, columns := false, false
	for line := range strings.Lines(ledger) {
		line = strings.TrimRight(line, "\n")
		if title, isHeading := strings.CutPrefix(line, "## "); isHeading {
			section = title
			continue
		}
		if section != "Corrections" {
			continue
		}
		found = true
		cells, isHeader := tableRow(line, &header)
		level, owner := slices.Index(header, "Level"), slices.Index(header, "Owner")
		if isHeader {
			columns = level >= 0 && owner >= 0
			continue
		}
		if cells == nil || !columns || len(cells) != len(header) {
			continue
		}
		if !slices.Contains([]string{"1", "2", "3", "4"}, cells[level]) {
			problems = append(problems, fmt.Sprintf("%s: level %s is not 1, 2, 3, or 4", cells[0], cells[level]))
		}
		if !slices.Contains([]string{"A", "B", "C"}, cells[owner]) {
			problems = append(problems, fmt.Sprintf("%s: owner %s is not A, B, or C", cells[0], cells[owner]))
		}
	}
	switch {
	case !found:
		problems = append(problems, "no Corrections section")
	case !columns:
		problems = append(problems, "the Corrections table has no Level and Owner columns")
	}
	return problems
}
