package extensionrepo

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// unfilled finds a line of a form that starts with Fill: after its indent.
var unfilled = regexp.MustCompile(`(?m)^[ \t]*Fill:`)

// A review is one readiness.md committed in the history of HEAD.
type review struct {
	revision, commit, text string
	rounds                 []round
	findings               []finding
	unread                 []string
}

// committedReviews returns each readiness.md committed in HEAD, the newest
// first. The ledger is cumulative, so each one stays its history (#417).
func committedReviews() []review {
	log, err := exec.Command("git", "log", "--format=%H", "HEAD", "--", "readiness.md").Output()
	if err != nil {
		return nil
	}
	var reviews []review
	for _, revision := range strings.Fields(string(log)) {
		data, err := exec.Command("git", "show", revision+":readiness.md").Output()
		if err != nil {
			continue
		}
		r := review{revision: revision, text: strings.TrimPrefix(string(data), "\ufeff")}
		if m := readinessCommit.FindStringSubmatch(readinessHeader(r.text)); m != nil {
			r.commit, _ = ledgerCommit(m[1])
		}
		r.rounds, _, _ = ledgerTables(r.text)
		r.findings, r.unread = readinessFindings(r.text)
		reviews = append(reviews, r)
	}
	return reviews
}

func readinessHeader(readiness string) string {
	lines := strings.SplitN(readiness, "\n", 4)
	return strings.Join(lines[:min(3, len(lines))], "\n")
}

// passProblems checks each round before the last whose Finish is pass. Such a
// round ends a build, so just finish passed at its head. Its newest committed
// review must pass as just finish read it then.
func passProblems(rounds []round, reviews []review) []string {
	var problems []string
	for i, r := range rounds[:max(0, len(rounds)-1)] {
		if len(r.cells) != len(roundColumns) || strings.Trim(r.cells[5], "`") != "pass" {
			continue
		}
		head, _ := ledgerCommit(r.cells[1])
		at := slices.IndexFunc(reviews, func(v review) bool { return v.commit == head })
		next := ""
		if len(rounds[i+1].cells) == len(roundColumns) {
			next, _ = ledgerCommit(rounds[i+1].cells[1])
		}
		if at >= 0 && !passedAt(reviews, at, head, next) {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: round %d says pass, but just finish fails on its review in %.12s. "+
				"Fix: write the result that just finish gave for that round", r.line, i+1, reviews[at].revision))
		}
	}
	return problems
}

// passedAt reports whether just finish passed on the review reviews[at] of
// head, as the repository stood when the review was committed. The forms
// hold no Fill: line and no open line, and only readiness.md and **Limits.**
// of spec.md differ from head. A later spec.md before the next round can
// stand in when only **Limits.** differs, so a limit committed after the
// review counts. The verdict and the states follow from the findings, which
// keep each finding of the reviews before. A review without a ledger, of a
// template before it, needs no state and no history.
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
	changed, err := exec.Command("git", "diff", "--name-only", head, r.revision, "--").Output()
	if err != nil || slices.ContainsFunc(strings.Fields(string(changed)), func(f string) bool { return f != "readiness.md" && f != "spec.md" }) {
		return false
	}
	v := readinessVerdict.FindStringSubmatch(readinessHeader(r.text))
	if v == nil || len(r.unread) != 0 {
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
		if len(open) != 0 || len(verdictProblems(v[1], r.findings, limits)) != 0 {
			continue
		}
		if len(r.rounds) == 0 || len(stateProblems(r.findings, len(r.rounds), spec, limits)) == 0 &&
			len(findingHistory(r.findings, reviews[at+1:], map[string]bool{})) == 0 {
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
// keeps its ID, or its title when it had no ID, and its severity never goes
// down.
func findingHistory(findings []finding, reviews []review, reported map[string]bool) []string {
	current := map[string]finding{}
	for _, f := range findings {
		current[f.id] = f
		current["title "+f.title] = f
	}
	var problems []string
	for _, r := range reviews {
		for _, e := range r.findings {
			key, name := e.id, e.id
			if key == "" {
				key, name = "title "+e.title, "titled "+strconv.Quote(e.title)
			}
			now, kept := current[key]
			switch {
			case reported[key]:
			case !kept:
				reported[key] = true
				problems = append(problems, fmt.Sprintf("readiness.md: the finding %s of the review in %.12s is missing. Fix: keep each earlier finding under ## Findings, with its ID and its state", name, r.revision))
			case severityRank[now.severity] < severityRank[e.severity]:
				reported[key] = true
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the finding %s had the severity %s in %.12s, and now has %s. Fix: keep its highest severity. A warn or block closes only as fixed or limit", now.line, name, e.severity, r.revision, now.severity))
			}
		}
	}
	return problems
}

// reviewedRounds matches the rounds with the committed reviews. Each review
// with a ledger names the head of a round, and each round before the last
// has a committed review of its head.
func reviewedRounds(rounds []round, reviews []review) []string {
	var heads, reviewed []string
	for _, r := range rounds {
		head := ""
		if len(r.cells) == len(roundColumns) {
			head, _ = ledgerCommit(r.cells[1])
		}
		heads = append(heads, head)
	}
	var problems []string
	for _, r := range reviews {
		if r.commit != "" && len(r.rounds) != 0 && !slices.Contains(heads, r.commit) && !slices.Contains(reviewed, r.commit) {
			problems = append(problems, fmt.Sprintf("readiness.md: the review of %.12s, committed in %.12s, has no round under ## Rounds. Fix: keep one row for each round, with the head that its review names", r.commit, r.revision))
		}
		reviewed = append(reviewed, r.commit)
	}
	for i, head := range heads[:max(0, len(heads)-1)] {
		if head != "" && !slices.Contains(reviewed, head) {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: round %d has no committed review of its head. Fix: commit readiness.md alone at the end of each round", rounds[i].line, i+1))
		}
	}
	return problems
}

// roundHistory requires each round of an earlier committed ledger in the
// current one, with the same cells. Only a Finish that was pending changes.
func roundHistory(rounds []round, r review, reported map[string]bool) []string {
	var problems []string
	for i, e := range r.rounds {
		key := "round " + strconv.Itoa(i+1)
		if reported[key] || len(e.cells) != len(roundColumns) {
			continue
		}
		kept := i < len(rounds) && len(rounds[i].cells) == len(roundColumns) &&
			slices.Equal(rounds[i].cells[:5], e.cells[:5]) &&
			(strings.Trim(e.cells[5], "`") == "pending" || rounds[i].cells[5] == e.cells[5])
		if !kept {
			reported[key] = true
			problems = append(problems, fmt.Sprintf("readiness.md: %s of the ledger in %.12s is missing or changed. Fix: keep the row of each earlier round as it was, and add one row for this round", key, r.revision))
		}
	}
	return problems
}
