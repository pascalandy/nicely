package extensionrepo

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// A finding is a ### heading under ## Findings of readiness.md, which starts
// with its ID, such as R2-F3. Its bullet - **Severity**: grades it. Its
// bullets - **State**:, - **Spec**:, - **Commit**:, and - **Reason**: give
// its state in the ledger (#417).
type finding struct {
	line                                          int
	id, title, severity, state, spec, commit, why string
}

var (
	readinessCommit   = regexp.MustCompile(`(?m)^Commit: ([0-9a-f]{7,40})\s*$`)
	readinessReviewer = regexp.MustCompile(`(?m)^Reviewer:[ \t]*\S`)
	readinessVerdict  = regexp.MustCompile(`(?m)^Verdict: (.+?)\s*$`)
	readinessHeading  = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	readinessPart     = regexp.MustCompile(`^\s*[-*+]\s+\*\*(Severity|State|Spec|Commit|Reason|Limit)(?:\*\*:|:\*\*)\s*(.*?)\s*$`)
	readinessID       = regexp.MustCompile(`^(R([1-9][0-9]*)-F[1-9][0-9]*)\b`)
	// readinessSeverity finds a severity in a form that finish cannot grade,
	// such as prose, code, or a table, anywhere after the header (#413, #416).
	readinessSeverity = regexp.MustCompile(`(?i)\b(block|warn|advise)\b`)
	// readinessWord finds the word severity under ## Findings (#413).
	readinessWord = regexp.MustCompile(`(?i)\bseverity\b`)
)

// severityCause names why a line that names a severity fails finish, so the
// reviewer sees the word or the table that caused it (#416).
func severityCause(text, fence string) string {
	const fix = "Fix: grade each finding only in its line - **Severity**:, and rephrase this line without "
	switch m := readinessSeverity.FindString(text); {
	case fence != "":
		return fmt.Sprintf("this fenced code names the severity %s. %sthe word %s", strings.ToLower(m), fix, strings.ToLower(m))
	case strings.HasPrefix(strings.TrimSpace(text), "|") && m != "":
		return fmt.Sprintf("this table names the severity %s. Fix: remove the table, because each finding gives its severity in its line - **Severity**:", strings.ToLower(m))
	case m != "":
		return fmt.Sprintf("this line names the severity %s outside the line - **Severity**: of a finding. %sthe word %s", strings.ToLower(m), fix, strings.ToLower(m))
	}
	return "this line names the word severity outside the line - **Severity**: of a finding. " + fix + "the word severity"
}

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
		bullet := part > 0 && fence == "" && m != nil && m[1] == "Severity"
		// The first three lines are the header. A severity elsewhere fails,
		// also under another part after ## Findings (#416).
		if i >= 3 && !bullet && (readinessSeverity.MatchString(text) || part > 0 && !section && readinessWord.MatchString(text)) {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: %s", i+1, severityCause(text, fence)))
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
				f := finding{line: i + 1, title: heading[2]}
				if id := readinessID.FindStringSubmatch(heading[2]); id != nil {
					f.id, f.title = id[1], strings.TrimSpace(heading[2][len(id[0]):])
				}
				findings = append(findings, f)
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
		case m[1] != "Severity":
			if current >= 0 {
				switch f := &findings[current]; m[1] {
				case "State":
					f.state = value
				case "Spec":
					f.spec = value
				case "Commit":
					f.commit = value
				case "Reason":
					f.why = value
				case "Limit":
					// The form of #409 named an accepted limit in its own bullet.
					f.state, f.spec = cmp.Or(f.state, "limit"), cmp.Or(f.spec, value)
				}
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
