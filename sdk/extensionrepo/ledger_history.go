package extensionrepo

import (
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// A review is one readiness.md committed in the history of the build.
type review struct {
	revision, commit, text string
}

func committedReviews(revisions string) []review {
	log, err := exec.Command("git", "log", "--format=%H", revisions, "--", "readiness.md").Output()
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
		reviews = append(reviews, r)
	}
	return reviews
}

func readinessHeader(readiness string) string {
	lines := strings.SplitN(readiness, "\n", 4)
	return strings.Join(lines[:min(3, len(lines))], "\n")
}

// buildStart returns the newest readiness.md committed at or before the
// intake that ended the build before this one, or "". A review that needs
// work never ends a build, so moving the intake hides no history (#417).
func buildStart(intake string) string {
	return endOfBuild(intake, 8)
}

// endOfBuild finds the review that ended a build at or before intake. Such a
// review passes as just finish would pass it. Its verdict and its states
// follow from its findings. Its ledger keeps the history of its own build.
// Its limits come from the spec.md of the intake, when only **Limits.**
// differs. So a limit accepted after the review counts. A review without a
// ledger, of an older template, needs a passing verdict only. depth bounds
// the builds that it reads.
func endOfBuild(intake string, depth int) string {
	if depth == 0 {
		return ""
	}
	later, _ := exec.Command("git", "show", intake+":spec.md").Output()
	for _, r := range committedReviews(intake) {
		v := readinessVerdict.FindStringSubmatch(readinessHeader(r.text))
		if v == nil || v[1] != "Ready" && v[1] != "Ready with stated limits" {
			continue
		}
		spec, _ := exec.Command("git", "show", r.revision+":spec.md").Output()
		if sameOutsideLimits(string(spec), string(later)) {
			spec = later
		}
		open, limits := specProblems(string(spec))
		findings, unread := readinessFindings(r.text)
		problems := append(append(open, unread...), verdictProblems(v[1], findings, limits)...)
		if rounds, order, _ := ledgerTables(r.text); len(rounds) != 0 {
			problems = append(problems, stateProblems(findings, len(rounds), string(spec), limits)...)
			since := ""
			if step, found := order["Intake"]; found {
				if commit, ok := ledgerCommit(step.cells[1]); ok {
					since = endOfBuild(commit, depth-1)
				}
			}
			problems = append(problems, historyAt(findings, rounds, since, r.revision+"^")...)
		}
		if len(problems) == 0 {
			return r.revision
		}
	}
	return ""
}

// historyProblems compares the ledger with each review of the build. A
// finding keeps its ID, or its title when it had no ID, in each later review,
// and its severity never goes down. Each round keeps its row, and each
// round matches one committed review, so no loss and no round leaves it.
func historyProblems(findings []finding, rounds []round, since string) []string {
	return historyAt(findings, rounds, since, "HEAD")
}

// historyAt compares the ledger with the reviews of the build up to head.
func historyAt(findings []finding, rounds []round, since, head string) []string {
	current := map[string]finding{}
	for _, f := range findings {
		current[f.id] = f
		current["title "+f.title] = f
	}
	var problems []string
	reported := map[string]bool{}
	revisions := head
	if since != "" {
		revisions = since + ".." + head
	}
	reviews := committedReviews(revisions)
	for _, r := range reviews {
		problems = append(problems, roundHistory(rounds, r.text, r.revision, reported)...)
		earlier, _ := readinessFindings(r.text)
		for _, e := range earlier {
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
	return append(problems, reviewedRounds(rounds, reviews)...)
}

// reviewedRounds matches the rounds with the committed reviews of the build.
// Each review names the head of a round, and each round before the last has
// a committed review of its head.
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
		reviewed = append(reviewed, r.commit)
		if r.commit != "" && !slices.Contains(heads, r.commit) && !slices.Contains(reviewed[:len(reviewed)-1], r.commit) {
			problems = append(problems, fmt.Sprintf("readiness.md: the review of %.12s, committed in %.12s, has no round under ## Rounds. Fix: keep one row for each round, with the head that its review names", r.commit, r.revision))
		}
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
func roundHistory(rounds []round, old, revision string, reported map[string]bool) []string {
	var problems []string
	earlier, _, _ := ledgerTables(old)
	for i, e := range earlier {
		key := "round " + strconv.Itoa(i+1)
		if reported[key] || len(e.cells) != len(roundColumns) {
			continue
		}
		kept := i < len(rounds) && len(rounds[i].cells) == len(roundColumns) &&
			slices.Equal(rounds[i].cells[:5], e.cells[:5]) &&
			(strings.Trim(e.cells[5], "`") == "pending" || rounds[i].cells[5] == e.cells[5])
		if !kept {
			reported[key] = true
			problems = append(problems, fmt.Sprintf("readiness.md: %s of the ledger in %.12s is missing or changed. Fix: keep the row of each earlier round as it was, and add one row for this round", key, revision))
		}
	}
	return problems
}
