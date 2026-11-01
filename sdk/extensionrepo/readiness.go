package extensionrepo

import (
	"cmp"
	"fmt"
	"maps"
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
	// readinessMention finds the ID of a finding anywhere in a line.
	readinessMention = regexp.MustCompile(`\bR[1-9][0-9]*-F[1-9][0-9]*\b`)
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

// readinessFindings reads the findings of readiness.md, outside fenced code.
// It reports each finding that finish cannot grade, so a finding in another
// form never passes as no finding. Under ## Findings, it also reports each
// line that names a severity, except the line - **Severity**: of a finding.
func readinessFindings(readiness string) ([]finding, []string) {
	const fix = "Fix: give each finding its own ### heading under ## Findings, with one line - **Severity**: `block`, `warn`, or `advise`"
	var findings []finding
	var problems []string
	part, current, fence, opened, none, stray, seen := 0, -1, "", 0, false, false, false
	// mentions holds the line of each finding ID outside ## Findings.
	mentions := map[string]int{}
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
		if i >= 3 && !bullet && (readinessSeverity.MatchString(text) || severityCell(text) != "" || readinessGrade.MatchString(text) ||
			part > 0 && !section && readinessWord.MatchString(text)) {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: %s", i+1, severityCause(text, fence)))
		}
		// A heading of level 3 with an ID outside ## Findings has its own
		// message below, so only other lines count as a mention.
		idHeading := fence == "" && heading != nil && len(heading[1]) == 3 && readinessID.MatchString(heading[2])
		if i >= 3 && (part == 0 || section) && !idHeading {
			for _, id := range readinessMention.FindAllString(text, -1) {
				if _, found := mentions[id]; !found {
					mentions[id] = i + 1
				}
			}
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
			case level == 3 && part == 0 && readinessID.MatchString(heading[2]):
				problems = append(problems, fmt.Sprintf("readiness.md:%d: this finding is outside ## Findings, so finish cannot grade it. %s", i+1, fix))
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
		delete(mentions, f.id)
	}
	// An ID outside ## Findings names a finding that finish grades, so no
	// finding hides in another part (#420).
	ids := slices.SortedFunc(maps.Keys(mentions), func(a, b string) int { return mentions[a] - mentions[b] })
	for _, id := range ids {
		problems = append(problems, fmt.Sprintf("readiness.md:%d: this line names the finding %s, which has no ### heading under ## Findings. %s", mentions[id], id, fix))
	}
	return findings, problems
}
