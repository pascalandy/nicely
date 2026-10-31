package extensionrepo

import (
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// stateProblems checks the state of each finding in the build of n rounds.
// A new finding of round n can have no state yet.
func stateProblems(findings []finding, n int, spec string, limits []limit) []string {
	var proven []string
	for _, l := range SpecLines(spec) {
		if l.ID != "" && (l.Part == "Failures" || l.Part == "Acceptance") {
			proven = append(proven, l.ID)
		}
	}
	var problems []string
	seen := map[string]bool{}
	for _, f := range findings {
		at := fmt.Sprintf("readiness.md:%d: ", f.line)
		m := readinessID.FindStringSubmatch(f.id)
		if m == nil {
			problems = append(problems, at+"this finding has no ID. Fix: start its heading with R, the round that first reported it, then -F and its number, such as ### R2-F1")
			continue
		}
		first, _ := strconv.Atoi(m[2])
		switch {
		case seen[f.id]:
			problems = append(problems, at+"another finding has the ID "+f.id+". Fix: give each finding its own ID, and keep it in later rounds")
		case first > n:
			problems = append(problems, fmt.Sprintf("%sthe finding %s is of round %d, after the last round %d of ## Rounds. Fix: correct its ID or the rounds", at, f.id, first, n))
		}
		seen[f.id] = true
		switch f.state {
		case "":
			if first < n {
				problems = append(problems, fmt.Sprintf("%sthe finding %s of round %d has no state. Fix: the reviewer of this round gives it the state open, fixed, limit, or declined", at, f.id, first))
			}
		case "open":
			problems = append(problems, fmt.Sprintf("%sthe finding %s is open, even if ## Open lines holds None. Fix: prove its spec ID in a round, or ask the user to accept it as a limit", at, f.id))
		case "fixed":
			if commit, ok := ledgerCommit(f.commit); !ok || commit == "" {
				problems = append(problems, at+"this fixed finding names no commit of HEAD. Fix: name the commit that fixed it in a line - **Commit**: <sha>")
			}
			if !slices.Contains(proven, f.spec) {
				problems = append(problems, at+"this fixed finding names no line of Failures or Acceptance in spec.md. Fix: name the spec ID that its test proves in a line - **Spec**: `<id>`")
			}
		case "limit":
			if !slices.ContainsFunc(limits, func(l limit) bool { return l.id == f.spec && l.accepted }) {
				problems = append(problems, at+"this limit names no limit that the user accepted in spec.md. Fix: name its row under **Limits.** in a line - **Spec**: `<id>`, after the user accepts it")
			}
		case "declined":
			if f.severity != "advise" {
				problems = append(problems, fmt.Sprintf("%sthe finding %s has the severity %s, and only an advise can be declined. Fix: fix it, or ask the user to accept it as a limit", at, f.id, f.severity))
			}
			if f.why == "" {
				problems = append(problems, at+"this declined finding gives no reason. Fix: write the reason of the author in a line - **Reason**:")
			}
		default:
			problems = append(problems, fmt.Sprintf("%sthe state %q is not open, fixed, limit, or declined", at, f.state))
		}
	}
	return problems
}

// historyProblems compares the findings with each committed readiness.md of
// this branch. A finding keeps its ID in each later review, and its severity
// never goes down, so no loss leaves the ledger.
func historyProblems(findings []finding) []string {
	revisions, err := exec.Command("git", "log", "--format=%H", "HEAD", "--", "readiness.md").Output()
	if err != nil {
		return []string{"readiness.md: git log: " + err.Error()}
	}
	current := map[string]finding{}
	for _, f := range findings {
		current[f.id] = f
	}
	var problems []string
	reported := map[string]bool{}
	for _, revision := range strings.Fields(string(revisions)) {
		old, err := exec.Command("git", "show", revision+":readiness.md").Output()
		if err != nil {
			continue
		}
		earlier, _ := readinessFindings(strings.TrimPrefix(string(old), "\ufeff"))
		for _, e := range earlier {
			now, kept := current[e.id]
			switch {
			case e.id == "" || reported[e.id]:
			case !kept:
				reported[e.id] = true
				problems = append(problems, fmt.Sprintf("readiness.md: the finding %s of the review in %.12s is missing. Fix: keep each earlier finding under ## Findings, with its ID and its state", e.id, revision))
			case severityRank[now.severity] < severityRank[e.severity]:
				reported[e.id] = true
				problems = append(problems, fmt.Sprintf("readiness.md:%d: the finding %s had the severity %s in %.12s, and now has %s. Fix: keep its highest severity. A warn or block closes only as fixed or limit", now.line, e.id, e.severity, revision, now.severity))
			}
		}
	}
	return problems
}
