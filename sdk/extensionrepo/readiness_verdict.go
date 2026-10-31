package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

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
