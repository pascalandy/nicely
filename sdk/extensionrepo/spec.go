package extensionrepo

import (
	"regexp"
	"slices"
	"strings"
)

// SpecAcceptedLimit reports whether id names one complete, accepted Limits row.
// It uses the acceptance state that finish reads and requires the canonical form.
func SpecAcceptedLimit(spec, id string) bool {
	_, limits := specProblems(spec)
	count, accepted := 0, false
	for _, limit := range limits {
		if limit.id == id {
			count++
			accepted = limit.accepted
		}
	}
	if count != 1 || !accepted {
		return false
	}
	header := false
	for _, line := range SpecLines(spec) {
		if line.Part != "Limits" {
			continue
		}
		switch line.Kind {
		case SpecLabel, SpecHeading:
			header = false
		case SpecHeader:
			header = slices.Equal(line.Cells, []string{"ID", "State", "Limit", "Accepted"})
		case SpecRow:
			if line.ID != id {
				continue
			}
			if !header || len(line.Cells) != 4 {
				return false
			}
			for _, cell := range line.Cells[2:] {
				text := strings.Trim(cell, "`* .-")
				if text == "" || strings.EqualFold(text, "None") || strings.HasPrefix(strings.ToLower(text), "fill:") {
					return false
				}
			}
			return true
		}
	}
	return false
}

// SpecKind says what a line of spec.md is.
type SpecKind int

// The kinds of a line of spec.md. SpecText is prose.
const (
	SpecText SpecKind = iota
	SpecBlank
	SpecSeparator
	SpecHeading
	SpecLabel
	SpecHeader
	SpecRow
	SpecBullet
)

// A SpecLine is one line of spec.md, read as "Spec IDs" in the contract
// says. A part starts at a heading, or at a line that starts with a bold
// label, such as **Failures.**, and ends at the next one. The test kit reads
// the parts Failures and Acceptance, and finish reads Open lines and Limits.
type SpecLine struct {
	Part   string
	Number int
	Kind   SpecKind
	// Cells holds the cells of a header or a row, without spaces around them.
	Cells []string
	// ID holds the spec ID that starts a row or a bullet, or "".
	ID string
	// Rest holds the text after the bold label of a SpecLabel.
	Rest string
}

var (
	specHeading   = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*$`)
	specLabel     = regexp.MustCompile(`^\*\*([^*]+)\*\*`)
	specBullet    = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+(.*)$`)
	specSeparator = regexp.MustCompile(`^\s*\|?(?:\s*:?-+:?\s*\|)*\s*:?-+:?\s*\|?\s*$`)
	specCellID    = regexp.MustCompile("^`([a-z][a-z0-9]*(?:-[a-z0-9]+)*)`$")
	specBulletID  = regexp.MustCompile("^`([a-z][a-z0-9]*(?:-[a-z0-9]+)*)`:")
)

// SpecLines reads each line of a spec.md. A row starts with its ID as its
// first cell. A bullet starts with its ID and a colon, such as
// - `duration`: the duration. So a bullet that only starts with code holds
// no ID. The header row of a table is the row above its separator.
func SpecLines(spec string) []SpecLine {
	lines := strings.Split(spec, "\n")
	found := make([]SpecLine, 0, len(lines))
	part := ""
	for i, text := range lines {
		l := SpecLine{Number: i + 1}
		row := strings.TrimSpace(text)
		switch {
		case row == "":
			l.Kind = SpecBlank
		case specHeading.MatchString(text):
			part = specHeading.FindStringSubmatch(text)[1]
			l.Kind = SpecHeading
		case specLabel.MatchString(text):
			m := specLabel.FindStringSubmatch(text)
			part = strings.TrimSuffix(strings.TrimSpace(m[1]), ".")
			l.Kind, l.Rest = SpecLabel, strings.TrimSpace(text[len(m[0]):])
		case strings.HasPrefix(row, "|") && specSeparator.MatchString(row):
			l.Kind = SpecSeparator
		case strings.HasPrefix(row, "|"):
			l.Kind, l.Cells = SpecRow, cells(row)
			if i+1 < len(lines) && specSeparator.MatchString(lines[i+1]) {
				l.Kind = SpecHeader
			} else if m := specCellID.FindStringSubmatch(l.Cells[0]); m != nil {
				l.ID = m[1]
			}
		case specBullet.MatchString(text):
			l.Kind = SpecBullet
			if m := specBulletID.FindStringSubmatch(specBullet.FindStringSubmatch(text)[1]); m != nil {
				l.ID = m[1]
			}
		}
		l.Part = part
		found = append(found, l)
	}
	return found
}

// cells splits a row of a table, which starts with a pipe.
func cells(row string) []string {
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|"), "|")
	for i, cell := range parts {
		parts[i] = strings.TrimSpace(cell)
	}
	return parts
}
