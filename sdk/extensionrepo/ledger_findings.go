package extensionrepo

import (
	"fmt"
	"slices"
	"strconv"
)

// stateProblems checks the state of each finding in the build of n rounds
// (#417). An open finding names its row under ## Open lines (#420, #423).
// A fixed finding names a line of a family of ## Case matrix. That family
// has a test that passes and a test that refuses (T43 of N2). Review still
// judges that the family is the whole family of the problem.
func stateProblems(findings []finding, n int, spec string, limits []limit) []string {
	proven, open := map[string]bool{}, map[string]bool{}
	for _, l := range SpecLines(spec) {
		switch {
		case l.ID == "":
		case l.Part == "Failures" || l.Part == "Acceptance":
			proven[l.ID] = true
		case l.Part == "Open lines" && l.Kind == SpecRow:
			open[l.ID] = true
		}
	}
	families := matrixFamilies(spec)
	var problems []string
	for _, f := range findings {
		at := "review.toml: the finding " + f.ID
		if m := findingID.FindStringSubmatch(f.ID); m != nil {
			if first, _ := strconv.Atoi(m[1]); first > n {
				problems = append(problems, fmt.Sprintf("%s is of round %d, after the last round %d of the ledger. Fix: correct its ID or the rounds", at, first, n))
			}
		}
		switch f.State {
		case "open":
			problems = append(problems, at+" is open, even if ## Open lines holds None. Fix: prove its spec ID in a round, or ask the user to accept it as a limit")
			if !open[f.Spec] {
				problems = append(problems, fmt.Sprintf("spec.md: ## Open lines holds no row `%s`, which the open finding %s names. "+
					"Fix: give the open finding the spec ID of its row under ## Open lines, so spec.md and the ledger agree on the open work", f.Spec, f.ID))
			}
		case "fixed":
			if commit, ok := ledgerCommit(f.Commit); !ok || commit == "" {
				problems = append(problems, "review.toml: the fixed finding "+f.ID+" names no commit of HEAD. Fix: name the commit that fixed it in its key commit")
			}
			if !proven[f.Spec] {
				problems = append(problems, "review.toml: the fixed finding "+f.ID+" names no line of Failures or Acceptance in spec.md. Fix: name the spec ID that its test proves in its key spec")
			}
			problems = append(problems, pairProblems(f, families, proven)...)
		case "limit":
			if !acceptedLimit(f, limits) {
				problems = append(problems, at+" names no limit that the user accepted in spec.md. Fix: name its row under **Limits.** in its key spec, after the user accepts it")
			}
		case "declined":
			if f.Severity != "advise" {
				problems = append(problems, fmt.Sprintf("%s has the severity %s, and only an advise can be declined. Fix: fix it, or ask the user to accept it as a limit", at, f.Severity))
			}
			if f.Reason == "" {
				problems = append(problems, "review.toml: the declined finding "+f.ID+" gives no reason. Fix: write the reason of the author in its key reason")
			}
		}
	}
	return problems
}

// pairProblems checks the family of a fixed finding. The family is a row of
// ## Case matrix, and it holds the spec ID of the finding. It also names a
// proven line that passes and one that refuses. So a fix adds a pair of
// tests of the family, not only a test of its example (#423).
func pairProblems(f finding, families map[string]family, proven map[string]bool) []string {
	at := "review.toml: the fixed finding " + f.ID
	fam, found := families[f.Family]
	if !found {
		return []string{fmt.Sprintf("%s names the family %q, which is no row of ## Case matrix. Fix: name the family of its pair of tests, as the column Family of spec.md names it", at, f.Family)}
	}
	var problems []string
	if !slices.Contains(fam.accepted, f.Spec) && !slices.Contains(fam.refused, f.Spec) {
		problems = append(problems, fmt.Sprintf("%s names the line `%s`, which is not in its family %q. Fix: name the family whose row names that line", at, f.Spec, f.Family))
	}
	if !slices.ContainsFunc(fam.accepted, func(id string) bool { return proven[id] }) {
		problems = append(problems, fmt.Sprintf("%s has no passing test in its family %q. Fix: add a test of the family that gives an input that passes, and name its line in Accepted", at, f.Family))
	}
	if !slices.ContainsFunc(fam.refused, func(id string) bool { return proven[id] }) {
		problems = append(problems, fmt.Sprintf("%s has no refused test in its family %q. Fix: add a test of the family that refuses an input, and name its line in Refused", at, f.Family))
	}
	return problems
}
