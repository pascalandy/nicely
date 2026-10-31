package extensionrepo

import (
	"errors"
	"fmt"
	"os"
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
// severity outside the line - **Severity**: of a finding, in fenced code too.
func readinessFindings(readiness string) ([]finding, []string) {
	const fix = "Fix: give each finding its own ### heading under ## Findings, with one line - **Severity**: `block`, `warn`, or `advise`"
	var findings []finding
	var problems []string
	hidden := func(line int) {
		problems = append(problems, fmt.Sprintf("readiness.md:%d: this line names a severity outside the line - **Severity**: of a finding. %s", line, fix))
	}
	part, current, fence, opened, none, stray, seen := 0, -1, "", 0, false, false, false
	closeFindings := func() {
		if part > 0 && len(findings) == 0 && !none {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: ## Findings holds no ### finding and no None. %s", part, fix))
		}
		part, current, stray = 0, -1, false
	}
	for i, text := range strings.Split(readiness, "\n") {
		trimmed := strings.TrimSpace(text)
		switch {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			} else if part > 0 && readinessSeverity.MatchString(text) {
				hidden(i + 1)
			}
			continue
		case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, "~~~"):
			fence, opened = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1]))], i+1
			continue
		}
		if m := readinessHeading.FindStringSubmatch(text); m != nil {
			switch level := len(m[1]); {
			case level <= 2:
				closeFindings()
				if level == 2 && m[2] == "Findings" {
					part, seen = i+1, true
				}
			case level == 3 && part > 0:
				findings = append(findings, finding{line: i + 1})
				current = len(findings) - 1
			}
			continue
		}
		none = none || part > 0 && current < 0 && trimmed == "None."
		m := readinessPart.FindStringSubmatch(text)
		if m == nil {
			if part > 0 && current < 0 && trimmed != "" && trimmed != "None." && !stray {
				stray = true
				problems = append(problems, fmt.Sprintf("readiness.md:%d: this text of ## Findings is outside every ### finding. %s", i+1, fix))
			} else if current >= 0 && readinessSeverity.MatchString(text) {
				hidden(i + 1)
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

// readinessProblems reports a readiness.md without a reviewer or a passing
// verdict, or whose commit is not the code of the working tree. The verdict
// must follow from the findings and from the limits of spec.md (#409). Only
// readiness.md may differ from that commit, so the review can be committed
// after it.
func readinessProblems(limits []limit) []string {
	const fix = "Fix: ask an agent that did not write the code to run the readiness review of the nicely-extension skill, then start readiness.md with the lines Commit: <sha>, Reviewer: <who>, and Verdict: <verdict>"
	const verdictFix = "Fix: write the verdict that the findings and the limits give, as the readiness review of the nicely-extension skill says"
	data, err := os.ReadFile("readiness.md")
	if errors.Is(err, os.ErrNotExist) {
		return []string{"readiness.md: no readiness verdict is recorded. " + fix}
	} else if err != nil {
		return []string{"readiness.md: " + err.Error()}
	}
	// readiness.md starts with the lines Commit:, Reviewer:, and Verdict:,
	// as the skill says. So finish never reads them in fenced code (#413).
	lines := strings.SplitN(string(data), "\n", 4)
	header := strings.Join(lines[:min(3, len(lines))], "\n")
	commit, verdict := readinessCommit.FindStringSubmatch(header), readinessVerdict.FindStringSubmatch(header)
	if commit == nil || verdict == nil {
		return []string{"readiness.md: it holds no line Commit: <sha> or Verdict: <verdict>. " + fix}
	}
	findings, problems := readinessFindings(string(data))
	if !readinessReviewer.MatchString(header) {
		problems = append(problems, "readiness.md: it names no reviewer. "+fix)
	}
	switch v := verdict[1]; v {
	case "Ready":
		for _, f := range findings {
			if f.severity == "block" || f.severity == "warn" {
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the verdict is Ready, but this finding has the severity %s. %s", f.line, f.severity, verdictFix))
			}
		}
		for _, l := range limits {
			problems = append(problems, fmt.Sprintf("spec.md:%d: the verdict is Ready, but spec.md states the limit `%s`. %s", l.line, l.id, verdictFix))
		}
	case "Ready with stated limits":
		for _, f := range findings {
			switch {
			case f.severity == "block":
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the verdict is Ready with stated limits, but this finding has the severity block. %s", f.line, verdictFix))
			case f.severity == "warn" && !slices.ContainsFunc(limits, func(l limit) bool { return l.id == f.limit && l.accepted }):
				problems = append(problems, fmt.Sprintf("readiness.md:%d: this warn names no limit that the user accepted in spec.md. Fix: name the spec ID of its accepted limit in a line - **Limit**: `<id>`, or write the verdict Needs work", f.line))
			}
		}
	case "Needs work":
		problems = append(problems, "readiness.md: the verdict is Needs work. Fix: start the next round from its findings")
	default:
		problems = append(problems, fmt.Sprintf("readiness.md: the verdict %q is not Ready, Ready with stated limits, or Needs work. %s", v, fix))
	}
	sha := commit[1]
	if _, err := updateOutput("git", "rev-parse", "--verify", "--quiet", sha+"^{commit}"); err != nil {
		return append(problems, fmt.Sprintf("readiness.md: Git knows no commit %s. %s", sha, fix))
	}
	changed, err := updateOutput("git", "diff", "--name-only", sha, "--")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	added, err := updateOutput("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	var files []string
	for _, file := range strings.Split(string(changed)+string(added), "\n") {
		if file != "" && file != "readiness.md" && !slices.Contains(files, file) {
			files = append(files, file)
		}
	}
	if len(files) != 0 {
		problems = append(problems, fmt.Sprintf("readiness.md: the review is of %s, and these files changed since: %s. Fix: run the readiness review again on this commit", sha, strings.Join(files, ", ")))
	}
	return problems
}
