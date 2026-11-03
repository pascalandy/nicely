package extensionrepo

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var (
	readinessHeading = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	findingID        = regexp.MustCompile(`^R([1-9][0-9]*)-F[1-9][0-9]*$`)
	fullSHA          = regexp.MustCompile(`^[0-9a-f]{40}$`)
	severityRank     = map[string]int{"advise": 1, "warn": 2, "block": 3}
)

const readinessFix = "Fix: ask an agent that did not write the code to run the readiness review of the nicely-extension skill, which writes review.toml, then run just records"

// reviewProblems reports a review record without a reviewer or a passing
// verdict, or whose commit is not the code of the working tree. The verdict
// must follow from the findings, their states, and the limits of spec.md
// (#409, #417). Only the record and its view may differ from that commit, so
// the review can be committed after it. Then the ledger of the build must
// hold. review is nil when review.toml is missing.
func reviewProblems(review *reviewRecord, spec string, limits []limit) []string {
	if review == nil {
		return []string{"review.toml: no readiness review is recorded. " + readinessFix}
	}
	var problems []string
	if review.Reviewer == "" {
		problems = append(problems, "review.toml: it names no reviewer. "+readinessFix)
	}
	problems = append(problems, findingProblems(review.Findings)...)
	problems = append(problems, verdictProblems(review.Verdict, review.Findings, limits)...)
	if !fullSHA.MatchString(review.Commit) {
		return append(problems, fmt.Sprintf("review.toml: its commit %q is no full SHA. Fix: write the 40 characters of the reviewed commit", review.Commit))
	}
	if _, err := updateOutput("git", "rev-parse", "--verify", "--quiet", review.Commit+"^{commit}"); err != nil {
		return append(problems, fmt.Sprintf("review.toml: Git knows no commit %s. %s", review.Commit, readinessFix))
	}
	changed, err := updateOutput("git", "diff", "--name-only", review.Commit, "--")
	if err != nil {
		return append(problems, "review.toml: "+err.Error())
	}
	added, err := updateOutput("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return append(problems, "review.toml: "+err.Error())
	}
	var files []string
	for _, file := range strings.Split(string(changed)+string(added), "\n") {
		if file != "" && !slices.Contains([]string{"readiness.md", reviewFile}, file) && !slices.Contains(files, file) && (file != "spec.md" || !limitsOnly(review.Commit)) {
			files = append(files, file)
		}
	}
	if len(files) != 0 {
		problems = append(problems, fmt.Sprintf("review.toml: the review is of %s, and these files changed since: %s. Fix: run the readiness review again on this commit", review.Commit, strings.Join(files, ", ")))
	}
	return append(problems, ledgerProblems(*review, spec, limits)...)
}

// findingProblems reports a finding without its ID, title, severity, state,
// spec ID, or family, and two findings with one ID. A finding without a state
// passed finish at the parent of 59d33f38, so each finding gives its state
// (#423). Finish reads a severity only in its key (#422).
func findingProblems(findings []finding) []string {
	var problems []string
	seen := map[string]bool{}
	for _, f := range findings {
		name := f.ID
		switch {
		case f.ID == "":
			name = "without an ID"
			problems = append(problems, "review.toml: a finding has no id. Fix: give it the ID R<round>-F<number>, such as R2-F1, of the round that first reported it")
		case !findingID.MatchString(f.ID):
			problems = append(problems, fmt.Sprintf("review.toml: the finding %q has no ID of the form R<round>-F<number>. Fix: give it an ID such as R2-F1, of the round that first reported it", f.ID))
		case seen[f.ID]:
			problems = append(problems, fmt.Sprintf("review.toml: two findings have the ID %s. Fix: give each finding its own ID, and keep it in later rounds", f.ID))
		}
		seen[f.ID] = true
		for _, part := range [][2]string{{"title", f.Title}, {"spec", f.Spec}, {"family", f.Family}} {
			if part[1] == "" {
				problems = append(problems, fmt.Sprintf("review.toml: the finding %s has no %s. Fix: give each finding its title, the spec ID of its line, and the family of ## Case matrix of that line", name, part[0]))
			}
		}
		switch {
		case f.Severity == "":
			problems = append(problems, fmt.Sprintf("review.toml: the finding %s has no severity. Fix: give it the severity block, warn, or advise, in its key severity", name))
		case severityRank[f.Severity] == 0:
			problems = append(problems, fmt.Sprintf("review.toml: the finding %s has the severity %q, which is not block, warn, or advise", name, f.Severity))
		}
		switch f.State {
		case "":
			problems = append(problems, fmt.Sprintf("review.toml: the finding %s has no state. Fix: give it the state open, fixed, limit, or declined", name))
		case "open", "fixed", "limit", "declined":
		default:
			problems = append(problems, fmt.Sprintf("review.toml: the finding %s has the state %q, which is not open, fixed, limit, or declined", name, f.State))
		}
	}
	return problems
}
