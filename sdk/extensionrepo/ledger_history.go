package extensionrepo

import (
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// A review is one readiness.md committed in the history of HEAD.
type review struct {
	revision, commit, text string
	rounds                 []round
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
		reviews = append(reviews, r)
	}
	return reviews
}

func readinessHeader(readiness string) string {
	lines := strings.SplitN(readiness, "\n", 4)
	return strings.Join(lines[:min(3, len(lines))], "\n")
}

// passProblems checks each round before the last whose Finish is pass. Such a
// round ends a build, so its newest committed review must pass. Its verdict
// and its states follow from its findings and from the limits of spec.md. It
// keeps each finding of the reviews before it. A review without a ledger, of
// a template before it, needs a passing verdict only.
func passProblems(rounds []round, reviews []review, spec string, limits []limit) []string {
	var problems []string
	for i, r := range rounds[:max(0, len(rounds)-1)] {
		if len(r.cells) != len(roundColumns) || strings.Trim(r.cells[5], "`") != "pass" {
			continue
		}
		head, _ := ledgerCommit(r.cells[1])
		at := slices.IndexFunc(reviews, func(v review) bool { return v.commit == head })
		if at < 0 {
			continue
		}
		v := readinessVerdict.FindStringSubmatch(readinessHeader(reviews[at].text))
		findings, unread := readinessFindings(reviews[at].text)
		failed := v == nil || len(unread) != 0
		if !failed {
			failed = len(verdictProblems(v[1], findings, limits)) != 0
		}
		if !failed && len(reviews[at].rounds) != 0 {
			failed = len(stateProblems(findings, len(reviews[at].rounds), spec, limits)) != 0 ||
				len(findingHistory(findings, reviews[at+1:], map[string]bool{})) != 0
		}
		if failed {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: round %d says pass, but its review in %.12s does not pass. Fix: write the result that just finish gave for that round", r.line, i+1, reviews[at].revision))
		}
	}
	return problems
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
