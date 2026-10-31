package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// readinessProblems reports a readiness.md without a reviewer or a passing
// verdict, or whose commit is not the code of the working tree. The verdict
// must follow from the findings, their states, and the limits of spec.md
// (#409, #417). Only readiness.md may differ from that commit, so the review
// can be committed after it. Then the ledger of the build must hold.
func readinessProblems(spec string, limits []limit) []string {
	const fix = readinessFix
	data, err := os.ReadFile("readiness.md")
	if errors.Is(err, os.ErrNotExist) {
		return []string{"readiness.md: no readiness verdict is recorded. " + fix}
	} else if err != nil {
		return []string{"readiness.md: " + err.Error()}
	}
	// readiness.md starts with the lines Commit:, Reviewer:, and Verdict:,
	// as the skill says. So finish never reads them in fenced code (#413).
	// An editor can write a byte order mark before them (#416).
	readiness := strings.TrimPrefix(string(data), "\ufeff")
	lines := strings.SplitN(readiness, "\n", 4)
	header := strings.Join(lines[:min(3, len(lines))], "\n")
	commit, verdict := readinessCommit.FindStringSubmatch(header), readinessVerdict.FindStringSubmatch(header)
	if commit == nil || verdict == nil {
		return []string{"readiness.md: its first three lines hold no line Commit: <sha> or Verdict: <verdict>. " + fix}
	}
	findings, problems := readinessFindings(readiness)
	if !readinessReviewer.MatchString(header) {
		problems = append(problems, "readiness.md: it names no reviewer. "+fix)
	}
	problems = append(problems, verdictProblems(verdict[1], findings, limits)...)
	sha, err := updateOutput("git", "rev-parse", "--verify", "--quiet", commit[1]+"^{commit}")
	if err != nil {
		return append(problems, fmt.Sprintf("readiness.md: Git knows no commit %s. %s", commit[1], fix))
	}
	changed, err := updateOutput("git", "diff", "--name-only", commit[1], "--")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	added, err := updateOutput("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	var files []string
	for _, file := range strings.Split(string(changed)+string(added), "\n") {
		if file != "" && file != "readiness.md" && !slices.Contains(files, file) && (file != "spec.md" || !limitsOnly(commit[1])) {
			files = append(files, file)
		}
	}
	if len(files) != 0 {
		problems = append(problems, fmt.Sprintf("readiness.md: the review is of %s, and these files changed since: %s. Fix: run the readiness review again on this commit", commit[1], strings.Join(files, ", ")))
	}
	return append(problems, ledgerProblems(readiness, strings.TrimSpace(string(sha)), findings, spec, limits)...)
}

const (
	readinessFix = "Fix: ask an agent that did not write the code to run the readiness review of the nicely-extension skill, then start readiness.md with the lines Commit: <sha>, Reviewer: <who>, and Verdict: <verdict>"
	verdictFix   = "Fix: write the verdict that the findings and the limits give, as the readiness review of the nicely-extension skill says"
)

// verdictProblems reports a verdict that does not follow from the findings,
// their states, and the limits of spec.md (#409, #417). Finish also runs it
// on an earlier review that would end a build.
func verdictProblems(verdict string, findings []finding, limits []limit) []string {
	var problems []string
	accepted := func(f finding) bool {
		return f.state == "limit" && slices.ContainsFunc(limits, func(l limit) bool { return l.id == f.spec && l.accepted })
	}
	switch v := verdict; v {
	case "Ready":
		for _, f := range findings {
			if (f.severity == "block" || f.severity == "warn") && f.state != "fixed" {
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the verdict is Ready, but this finding has the severity %s and is not fixed. %s", f.line, f.severity, verdictFix))
			}
		}
		for _, l := range limits {
			problems = append(problems, fmt.Sprintf("spec.md:%d: the verdict is Ready, but spec.md states the limit `%s`. %s", l.line, l.id, verdictFix))
		}
	case "Ready with stated limits":
		for _, f := range findings {
			switch {
			case f.state == "fixed":
			case f.severity == "block":
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the verdict is Ready with stated limits, but this finding has the severity block and is not fixed. %s", f.line, verdictFix))
			case f.severity == "warn" && !accepted(f):
				problems = append(problems, fmt.Sprintf("readiness.md:%d: this warn names no limit that the user accepted in spec.md. Fix: give it the state `limit` and the spec ID of its accepted limit in a line - **Spec**: `<id>`, or write the verdict Needs work", f.line))
			}
		}
	case "Needs work":
		problems = append(problems, "readiness.md: the verdict is Needs work. Fix: start the next round from its findings")
	default:
		problems = append(problems, fmt.Sprintf("readiness.md: the verdict %q is not Ready, Ready with stated limits, or Needs work. %s", v, readinessFix))
	}
	return problems
}

// limitsOnly reports whether spec.md differs from its version in commit only
// under **Limits.**. So the user accepts a limit after a review, and no new
// round starts for it (#417).
func limitsOnly(commit string) bool {
	old, err := exec.Command("git", "show", commit+":spec.md").Output()
	if err != nil {
		return false
	}
	current, err := os.ReadFile("spec.md")
	return err == nil && sameOutsideLimits(string(old), string(current))
}

// sameOutsideLimits reports whether two texts of spec.md differ only under
// **Limits.**.
func sameOutsideLimits(a, b string) bool {
	outside := func(spec string) []string {
		lines := strings.Split(spec, "\n")
		var kept []string
		for _, l := range SpecLines(spec) {
			if l.Part != "Limits" {
				kept = append(kept, lines[l.Number-1])
			}
		}
		return kept
	}
	return slices.Equal(outside(a), outside(b))
}
