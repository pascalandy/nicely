package extensionrepo

import (
	"fmt"
	"regexp"
	"strings"
)

// The forms of a severity that finish refuses outside the line
// - **Severity**: of a finding.
var (
	// readinessSeverity finds a severity in code, a form that finish cannot
	// grade, anywhere after the header (#413, #416). The plain words pass, so
	// a title or prose can say that a call blocks or warns (#419).
	readinessSeverity = regexp.MustCompile("(?i)`(block|warn|advise)`")
	// readinessGrade finds the word severity or grade before a severity, in
	// any part. Only marks can stand between them, such as in "Severity:
	// block" or "Grade: warn" (#420, #422).
	readinessGrade = regexp.MustCompile(`(?i)\b(severity|grade[sd]?)\b[^a-z0-9]*\b(block|warn|advise)\b`)
	// readinessLevel finds a severity, a hyphen, and the word level, such as
	// "warn-level", in any part (#422).
	readinessLevel = regexp.MustCompile(`(?i)\b(block|warn|advise)-level\b`)
	// readinessWord finds the word severity under ## Findings (#413).
	readinessWord = regexp.MustCompile(`(?i)\bseverity\b`)
)

// severityCause names why a line that names a severity fails finish, so the
// reviewer sees the word or the table that caused it (#416).
func severityCause(text, fence string) string {
	const fix = "Fix: grade each finding only in its line - **Severity**:, and write "
	code, cell := readinessSeverity.FindStringSubmatch(text), severityCell(text)
	grade, level := readinessGrade.FindStringSubmatch(text), readinessLevel.FindStringSubmatch(text)
	switch {
	case cell != "":
		return fmt.Sprintf("this table names the severity %s. Fix: remove the table, because each finding gives its severity in its line - **Severity**:", cell)
	case code != nil && fence != "":
		return fmt.Sprintf("this fenced code names the severity `%s`. %sthe word %s in other text without backticks", strings.ToLower(code[1]), fix, strings.ToLower(code[1]))
	case code != nil:
		return fmt.Sprintf("this line names the severity `%s` in code, outside the line - **Severity**: of a finding. %sthe word %s in other text without backticks", strings.ToLower(code[1]), fix, strings.ToLower(code[1]))
	case grade != nil:
		return fmt.Sprintf("this line grades with the severity %s outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word %s", strings.ToLower(grade[2]), strings.ToLower(grade[1]))
	case level != nil:
		return fmt.Sprintf("this line grades with the severity %s outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the form %s", strings.ToLower(level[1]), level[0])
	}
	return "this line names the word severity outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word severity"
}

// severityCell returns the severity of a cell of a table row. Any row can
// hold it alone, with only marks around it, such as `warn` or **block**. A
// row that names a finding ID can also end a cell with it, such as
// | R1-F1 | really a warn |. It returns "" for any other line. So the cell
// "the header block" passes (#416, #420, #422).
func severityCell(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "|") {
		return ""
	}
	finding := readinessMention.MatchString(trimmed)
	for _, cell := range strings.Split(trimmed, "|") {
		words := strings.Fields(strings.ToLower(strings.Trim(cell, " \t`*_().:!")))
		if len(words) == 0 || len(words) > 1 && !finding {
			continue
		}
		if word := strings.Trim(words[len(words)-1], "`*_().:!"); severityRank[word] > 0 {
			return word
		}
	}
	return ""
}
