package extensionrepo

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// SpecCheckOutputLimit requires one accepted Limits row with a description and
// the exact Output case. The matrix must bind that ID too.
func SpecCheckOutputLimit(spec, id, name string) error {
	notAccepted := fmt.Errorf("limit %q is not accepted", id)
	_, limits := specProblems(spec)
	count, accepted := 0, false
	for _, limit := range limits {
		if limit.id == id {
			count++
			accepted = limit.accepted
		}
	}
	if count != 1 || !accepted {
		return notAccepted
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
				return notAccepted
			}
			if !specLimitDescription(line.Cells[2], id) {
				return fmt.Errorf("limit %q is missing Limit text describing what is retained", id)
			}
			if specPlaceholder(line.Cells[3]) || !SpecNamesOutput(line.Cells[2], name) {
				return notAccepted
			}
			return nil
		}
	}
	return notAccepted
}

// SpecNamesOutput reports whether a semicolon-separated cell declares the exact
// Output case. The Limit and matrix Source cells share this grammar.
func SpecNamesOutput(cell, name string) bool {
	for _, part := range specParts(cell) {
		if part == "Output: "+strings.ReplaceAll(name, "`", "") {
			return true
		}
	}
	return false
}

func specParts(cell string) []string {
	parts := strings.Split(strings.ReplaceAll(cell, "`", ""), ";")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func specLimitDescription(cell, id string) bool {
	described := false
	for _, part := range specParts(cell) {
		text := strings.Trim(part, "* .-")
		if text == "" || text == id || strings.HasPrefix(strings.ToLower(text), "output:") {
			continue
		}
		if specPlaceholder(text) {
			return false
		}
		described = true
	}
	return described
}

func specPlaceholder(text string) bool {
	text = strings.Trim(text, "`* .-")
	return text == "" || strings.EqualFold(text, "None") || strings.HasPrefix(strings.ToLower(text), "fill:")
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
