package extensionrepo

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// A finding is a ### heading under ## Findings of readiness.md. Its bullet
// - **Severity**: grades it. For a warn, its bullet - **Limit**: names the
// spec ID of the limit that the user accepted for it (#409).
type finding struct {
	line     int
	severity string
	limit    string
}

var (
	readinessCommit   = regexp.MustCompile(`(?m)^Commit: ([0-9a-f]{7,40})\s*$`)
	readinessReviewer = regexp.MustCompile(`(?m)^Reviewer:[ \t]*\S`)
	readinessVerdict  = regexp.MustCompile(`(?m)^Verdict: (.+?)\s*$`)
	readinessHeading  = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	readinessPart     = regexp.MustCompile(`^\s*[-*+]\s+\*\*(Severity|Limit)(?:\*\*:|:\*\*)\s*(.*?)\s*$`)
	// readinessSeverity finds a severity in a form that finish cannot grade.
	// Such a form is the word severity, a severity in code, or a severity
	// cell of a table (#413).
	readinessSeverity = regexp.MustCompile(`(?i)\bseverity\b|` + "`(?:block|warn|advise)`" + `|\|\s*(?:block|warn|advise)\s*(?:\||$)`)
)

// readinessFindings reads the findings of readiness.md, outside fenced code.
// It reports each finding that finish cannot grade, so a finding in another
// form never passes as no finding. Under ## Findings, it also reports each
// line that names a severity, except the line - **Severity**: of a finding.
func readinessFindings(readiness string) ([]finding, []string) {
	const fix = "Fix: give each finding its own ### heading under ## Findings, with one line - **Severity**: `block`, `warn`, or `advise`"
	var findings []finding
	var problems []string
	part, current, fence, opened, none, stray, seen := 0, -1, "", 0, false, false, false
	closeFindings := func() {
		if part > 0 && len(findings) == 0 && !none {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: ## Findings holds no ### finding and no None. %s", part, fix))
		}
		part, current, stray = 0, -1, false
	}
	for i, text := range strings.Split(readiness, "\n") {
		trimmed := strings.TrimSpace(text)
		m, heading := readinessPart.FindStringSubmatch(text), readinessHeading.FindStringSubmatch(text)
		// A heading of level 1 or 2 closes ## Findings, so it is no line of a finding.
		section := fence == "" && heading != nil && len(heading[1]) <= 2
		bullet := fence == "" && m != nil && m[1] == "Severity"
		if part > 0 && !section && !bullet && readinessSeverity.MatchString(text) {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: this line names a severity outside the line - **Severity**: of a finding. %s", i+1, fix))
		}
		switch {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			}
			continue
		case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, "~~~"):
			fence, opened = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1]))], i+1
			continue
		}
		if heading != nil {
			switch level := len(heading[1]); {
			case level <= 2:
				closeFindings()
				if level == 2 && heading[2] == "Findings" {
					part, seen = i+1, true
				}
			case level == 3 && part > 0:
				findings = append(findings, finding{line: i + 1})
				current = len(findings) - 1
			}
			continue
		}
		none = none || part > 0 && current < 0 && trimmed == "None."
		if m == nil {
			if part > 0 && current < 0 && trimmed != "" && trimmed != "None." && !stray {
				stray = true
				problems = append(problems, fmt.Sprintf("readiness.md:%d: this text of ## Findings is outside every ### finding. %s", i+1, fix))
			}
			continue
		}
		value := strings.Trim(strings.TrimSuffix(m[2], "."), "`")
		switch {
		case m[1] == "Limit":
			if current >= 0 {
				findings[current].limit = value
			}
		case current < 0:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: this severity has no finding of its own. %s", i+1, fix))
		case findings[current].severity != "":
			problems = append(problems, fmt.Sprintf("readiness.md:%d: this finding already has a severity. %s", i+1, fix))
		default:
			findings[current].severity = value
			if !slices.Contains([]string{"block", "warn", "advise"}, value) {
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the severity %q is not block, warn, or advise. %s", i+1, value, fix))
			}
		}
	}
	if fence != "" {
		problems = append(problems, fmt.Sprintf("readiness.md:%d: this fenced code has no end. Fix: close it with %s", opened, fence))
	}
	closeFindings()
	if !seen {
		problems = append(problems, "readiness.md: it holds no part ## Findings. "+fix+", or write None. under it")
	}
	for _, f := range findings {
		if f.severity == "" {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: this finding has no severity. %s", f.line, fix))
		}
	}
	return findings, problems
}
