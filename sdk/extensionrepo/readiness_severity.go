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
	// readinessGrade finds the word severity before a severity, a grade in
	// prose such as "Severity: block", in any part (#420).
	readinessGrade = regexp.MustCompile(`(?i)\bseverity\b[^a-z0-9]*\b(block|warn|advise)\b`)
	// readinessWord finds the word severity under ## Findings (#413).
	readinessWord = regexp.MustCompile(`(?i)\bseverity\b`)
)

// severityCause names why a line that names a severity fails finish, so the
// reviewer sees the word or the table that caused it (#416).
func severityCause(text, fence string) string {
	const fix = "Fix: grade each finding only in its line - **Severity**:, and write "
	code, cell, grade := readinessSeverity.FindStringSubmatch(text), severityCell(text), readinessGrade.FindStringSubmatch(text)
	switch {
	case cell != "":
		return fmt.Sprintf("this table names the severity %s. Fix: remove the table, because each finding gives its severity in its line - **Severity**:", cell)
	case code != nil && fence != "":
		return fmt.Sprintf("this fenced code names the severity `%s`. %sthe word %s in other text without backticks", strings.ToLower(code[1]), fix, strings.ToLower(code[1]))
	case code != nil:
		return fmt.Sprintf("this line names the severity `%s` in code, outside the line - **Severity**: of a finding. %sthe word %s in other text without backticks", strings.ToLower(code[1]), fix, strings.ToLower(code[1]))
	case grade != nil:
		return fmt.Sprintf("this line grades with the severity %s outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word severity", strings.ToLower(grade[1]))
	}
	return "this line names the word severity outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word severity"
}

// severityCell returns the severity that a cell of a table row holds alone,
// with only marks around it, such as `warn`, **block**, or (advise). It
// returns "" for any other line (#416, #420).
func severityCell(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "|") {
		return ""
	}
	for _, cell := range strings.Split(trimmed, "|") {
		if word := strings.ToLower(strings.Trim(cell, " \t`*_().:!")); severityRank[word] > 0 {
			return word
		}
	}
	return ""
}
