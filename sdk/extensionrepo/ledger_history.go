package extensionrepo

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// unfilled finds a line of a form that starts with Fill: after its indent,
// and intakeUnfilled a line of intake.toml that starts with # Fill:.
var (
	unfilled       = regexp.MustCompile(`(?m)^[ \t]*Fill:`)
	intakeUnfilled = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*Fill:`)
)

// A review is one review.toml committed in the history of HEAD.
type review struct {
	revision, commit string
	record           reviewRecord
	readable         bool
}

// committedReviews returns each review.toml committed in HEAD after the
// commit markdown, the newest first. The ledger is cumulative, so each one
// stays its history (#417). The Markdown reviews before markdown are no
// history of the records, as build-records.md in the contract says.
func committedReviews(markdown string) []review {
	args := []string{"log", "--format=%H", "HEAD", "--", reviewFile}
	if markdown != "" {
		args = []string{"log", "--format=%H", markdown + "..HEAD", "--", reviewFile}
	}
	log, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil
	}
	var reviews []review
	for _, revision := range strings.Fields(string(log)) {
		data, err := exec.Command("git", "show", revision+":"+reviewFile).Output()
		if err != nil {
			continue
		}
		r := review{revision: revision}
		r.readable = len(decodeRecord(reviewFile, string(data), &r.record)) == 0 && len(findingProblems(r.record.Findings)) == 0
		r.commit, _ = ledgerCommit(r.record.Commit)
		reviews = append(reviews, r)
	}
	return reviews
}

// passProblems checks each round before the last whose finish is pass. Such a
// round ends a build, so just finish passed at its head. Its newest committed
// review must pass as just finish read it then.
func passProblems(rounds []round, reviews []review) []string {
	var problems []string
	for i, r := range rounds[:max(0, len(rounds)-1)] {
		if r.Finish != "pass" {
			continue
		}
		head, _ := ledgerCommit(r.Head)
		at := slices.IndexFunc(reviews, func(v review) bool { return v.commit == head })
		next, _ := ledgerCommit(rounds[i+1].Head)
		if at >= 0 && !passedAt(reviews, at, head, next) {
			problems = append(problems, fmt.Sprintf("review.toml: round %d says pass, but just finish fails on its review in %.12s. "+
				"Fix: write the result that just finish gave for that round", i+1, reviews[at].revision))
		}
	}
	return problems
}

// passedAt reports whether just finish passed on the review reviews[at] of
// head, as the repository stood when the review was committed. The forms
// hold no Fill: line and no open line, and only the records, their views,
// and **Limits.** of spec.md differ from head. A later spec.md before the
// next round can stand in when only **Limits.** differs, so a limit
// committed after the review counts. The rows of ## Open lines of those
// limits may leave too. The records read, and their views follow them. The
// verdict and the states follow from the findings, which keep each finding
// of the reviews before.
func passedAt(reviews []review, at int, head, next string) bool {
	r := reviews[at]
	show := func(commit, file string) (string, bool) {
		data, err := exec.Command("git", "show", commit+":"+file).Output()
		return string(data), err == nil && !unfilled.MatchString(string(data))
	}
	for _, form := range []string{"vision.md", "qa.md"} {
		if _, ok := show(r.revision, form); !ok {
			return false
		}
	}
	if intake, ok := show(r.revision, intakeFile); !ok || intakeUnfilled.MatchString(intake) {
		return false
	}
	changed, err := exec.Command("git", "diff", "--name-only", head, r.revision, "--").Output()
	if err != nil || slices.ContainsFunc(strings.Fields(string(changed)), func(f string) bool {
		return !slices.Contains([]string{"readiness.md", reviewFile, "spec.md"}, f)
	}) {
		return false
	}
	// A stale view failed just finish then, so its review ended no build.
	intake, shown := intakeAt(r.revision)
	if view, _ := show(r.revision, "readiness.md"); !r.readable || !shown || view != readinessView(r.record) {
		return false
	}
	reviewed, _ := show(head, "spec.md")
	candidates := []string{r.revision}
	if next != "" {
		later, _ := exec.Command("git", "log", "--format=%H", r.revision+".."+next, "--", "spec.md").Output()
		candidates = append(candidates, strings.Fields(string(later))...)
	}
	for _, commit := range candidates {
		spec, ok := show(commit, "spec.md")
		if !ok || !sameOutsideLimits(reviewed, spec) {
			continue
		}
		open, limits := specProblems(spec)
		if len(open) != 0 || len(verdictProblems(r.record.Verdict, r.record.Findings, limits)) != 0 {
			continue
		}
		if len(stateProblems(r.record.Findings, len(r.record.Rounds), spec, limits)) == 0 && len(intakeProblems(intake, spec, limits)) == 0 &&
			len(findingHistory(r.record.Findings, reviews[at+1:], map[string]bool{})) == 0 {
			return true
		}
	}
	return false
}

// historyProblems compares the ledger with each committed review. Each round
// keeps its row, and each round matches one committed review, so no loss and
// no round leaves the ledger.
func historyProblems(findings []finding, rounds []round, reviews []review) []string {
	var problems []string
	reported := map[string]bool{}
	for _, r := range reviews {
		problems = append(problems, roundHistory(rounds, r, reported)...)
	}
	problems = append(problems, findingHistory(findings, reviews, reported)...)
	return append(problems, reviewedRounds(rounds, reviews)...)
}

// findingHistory requires each finding of the reviews in findings. A finding
// keeps its ID, and its severity never goes down.
func findingHistory(findings []finding, reviews []review, reported map[string]bool) []string {
	current := map[string]finding{}
	for _, f := range findings {
		current[f.ID] = f
	}
	var problems []string
	for _, r := range reviews {
		for _, e := range r.record.Findings {
			now, kept := current[e.ID]
			switch {
			case e.ID == "" || reported[e.ID]:
			case !kept:
				reported[e.ID] = true
				problems = append(problems, fmt.Sprintf("review.toml: the finding %s of the review in %.12s is missing. Fix: keep each earlier finding, with its ID and its state", e.ID, r.revision))
			case severityRank[now.Severity] < severityRank[e.Severity]:
				reported[e.ID] = true
				problems = append(problems, fmt.Sprintf("review.toml: the finding %s had the severity %s in %.12s, and now has %s. Fix: keep its highest severity. A warn or block closes only as fixed or limit", e.ID, e.Severity, r.revision, now.Severity))
			}
		}
	}
	return problems
}

// reviewedRounds matches the rounds with the committed reviews. Each review
// names the head of a round, and each round before the last has a committed
// review of its head.
func reviewedRounds(rounds []round, reviews []review) []string {
	var heads, reviewed []string
	for _, r := range rounds {
		head, _ := ledgerCommit(r.Head)
		heads = append(heads, head)
	}
	var problems []string
	for _, r := range reviews {
		if r.commit != "" && !slices.Contains(heads, r.commit) && !slices.Contains(reviewed, r.commit) {
			problems = append(problems, fmt.Sprintf("review.toml: the review of %.12s, committed in %.12s, has no round. Fix: keep one table [[rounds]] for each round, with the head that its review names", r.commit, r.revision))
		}
		reviewed = append(reviewed, r.commit)
	}
	for i, head := range heads[:max(0, len(heads)-1)] {
		if head != "" && !slices.Contains(reviewed, head) {
			problems = append(problems, fmt.Sprintf("review.toml: round %d has no committed review of its head. Fix: commit review.toml and readiness.md alone at the end of each round", i+1))
		}
	}
	return problems
}

// roundHistory requires each round of an earlier committed ledger in the
// current one, with the same values. Only a finish that was pending changes.
func roundHistory(rounds []round, r review, reported map[string]bool) []string {
	var problems []string
	for i, e := range r.record.Rounds {
		key := fmt.Sprintf("round %d", i+1)
		if reported[key] {
			continue
		}
		kept := i < len(rounds) && rounds[i].Head == e.Head && rounds[i].ChangeReview == e.ChangeReview &&
			rounds[i].ChangeCommits == e.ChangeCommits && rounds[i].ReadinessReview == e.ReadinessReview &&
			(e.Finish == "pending" || rounds[i].Finish == e.Finish)
		if !kept {
			reported[key] = true
			problems = append(problems, fmt.Sprintf("review.toml: %s of the ledger in %.12s is missing or changed. Fix: keep each earlier round as it was, and add one table [[rounds]] for this round", key, r.revision))
		}
	}
	return problems
}

// intakeAt returns the intake record at revision. It also reports whether
// the record reads, and ## Intake of vision.md is its view, as just finish
// checked them then.
func intakeAt(revision string) (intakeRecord, bool) {
	var intake intakeRecord
	data, err := exec.Command("git", "show", revision+":"+intakeFile).Output()
	if err != nil || len(decodeRecord(intakeFile, string(data), &intake)) != 0 {
		return intake, false
	}
	vision, err := exec.Command("git", "show", revision+":vision.md").Output()
	view, ok := withIntake(string(vision), intakeView(intake))
	return intake, err == nil && ok && view == string(vision)
}
