package extensionrepo

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// maxRounds bounds the rounds of one build of a command, from the intake to
// "Done when" (#417).
const maxRounds = 5

// The ledger of readiness.md records each round of the build in a table
// under ## Rounds. It records the commit of each step of round 1 under
// ## Order of round 1. Each finding under ## Findings keeps its ID from the
// round that first reported it, and gets a state (#417).
var (
	roundColumns  = []string{"Round", "Head", "Change review", "Change commits", "Readiness review", "Finish"}
	orderSteps    = []string{"Intake", "Spec", "Scenario", "Handler"}
	ledgerSHA     = regexp.MustCompile("^`?([0-9a-f]{7,40})`?$")
	ledgerFullSHA = regexp.MustCompile("^`?([0-9a-f]{40})`?$")
	severityRank  = map[string]int{"advise": 1, "warn": 2, "block": 3}
)

type round struct {
	line  int
	cells []string
}

// ledgerProblems reports each rule of the ledger that readiness.md breaks.
// reviewed is the full SHA of its line Commit:.
func ledgerProblems(readiness, reviewed string, findings []finding, spec string, limits []limit) []string {
	const fix = "Fix: give readiness.md a table under ## Rounds with the columns " +
		"Round, Head, Change review, Change commits, Readiness review, and Finish, one row for each round"
	var problems []string
	var rounds []round
	order := map[string]round{}
	for _, l := range SpecLines(readiness) {
		switch {
		case l.Part == "Rounds" && l.Kind == SpecHeader && !slices.Equal(l.Cells, roundColumns):
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the table of ## Rounds has other columns. %s", l.Number, fix))
		case l.Part == "Rounds" && l.Kind == SpecRow:
			rounds = append(rounds, round{l.Number, l.Cells})
		case l.Part == "Order of round 1" && l.Kind == SpecRow && len(l.Cells) == 2:
			order[l.Cells[0]] = round{l.Number, l.Cells}
		}
	}
	if len(rounds) == 0 {
		return append(problems, "readiness.md: it holds no round under ## Rounds. "+fix)
	}
	if len(rounds) > maxRounds {
		problems = append(problems, fmt.Sprintf("readiness.md: the ledger holds %d rounds, and a build has at most %d. "+
			"Fix: stop, and report each open line and each open finding to the user. Do not start another round", len(rounds), maxRounds))
	}
	for i, r := range rounds {
		problems = append(problems, roundProblems(r, i+1, i == len(rounds)-1)...)
	}
	if last := rounds[len(rounds)-1]; len(last.cells) == len(roundColumns) {
		if head, ok := ledgerCommit(last.cells[1]); ok && head != reviewed {
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the last round names the head %s, and the line Commit: names %s. "+
				"Fix: name the reviewed commit in both", last.line, last.cells[1], reviewed))
		}
	}
	problems = append(problems, orderProblems(order)...)
	problems = append(problems, stateProblems(findings, len(rounds), spec, limits)...)
	// The build starts at its intake, so the reviews of an earlier build,
	// such as of another command, are no history of this one.
	since := ""
	if intake, found := order["Intake"]; found {
		since, _ = ledgerCommit(intake.cells[1])
	}
	return append(problems, historyProblems(findings, since)...)
}

// roundProblems checks one row of ## Rounds, the round number of its place.
func roundProblems(r round, number int, last bool) []string {
	at := fmt.Sprintf("readiness.md:%d: ", r.line)
	if len(r.cells) != len(roundColumns) {
		return []string{at + "this round has no cell for each column of ## Rounds. Fix: fill Round, Head, Change review, Change commits, Readiness review, and Finish"}
	}
	var problems []string
	if r.cells[0] != fmt.Sprint(number) {
		problems = append(problems, fmt.Sprintf("%sthis row is round %d, but its Round cell holds %q. Fix: number the rounds 1, 2, 3, in order", at, number, r.cells[0]))
	}
	if !ledgerFullSHA.MatchString(r.cells[1]) {
		problems = append(problems, at+"its Head is no full SHA. Fix: write the 40 characters of the reviewed commit")
	}
	commits := []string{r.cells[1]}
	for _, c := range strings.Split(r.cells[3], "..") {
		commits = append(commits, strings.TrimSpace(c))
	}
	for _, c := range commits {
		if _, ok := ledgerCommit(c); !ok {
			problems = append(problems, fmt.Sprintf("%sthe commit %q is not an ancestor of HEAD. Fix: name the commits of this branch that the round reviewed, as <first>..<last>", at, c))
		}
	}
	if r.cells[2] == "" || r.cells[4] == "" {
		problems = append(problems, at+"it names no reviewer of the change review or of the readiness review. Fix: name the agent that ran each review, which wrote none of its code")
	}
	finish, want := strings.Trim(r.cells[5], "`"), []string{"pass", "fail"}
	if last {
		want = []string{"pending"}
	}
	if !slices.Contains(want, finish) {
		problems = append(problems, fmt.Sprintf("%sits Finish is %q. Fix: write the result of just finish for the round, pass or fail, and pending for the last round, whose result is this run", at, finish))
	}
	return problems
}

// orderProblems checks that round 1 recorded the intake, the spec, the
// first scenario, and the handler, each in a commit after the one before.
// The scenario precedes the handler, so it could fail first.
func orderProblems(order map[string]round) []string {
	const fix = "Fix: give readiness.md a table under ## Order of round 1 with the columns Step and Commit, and the rows Intake, Spec, Scenario, and Handler"
	var problems []string
	previous := ""
	for _, step := range orderSteps {
		r, found := order[step]
		if !found {
			problems = append(problems, fmt.Sprintf("readiness.md: ## Order of round 1 gives no commit for the step %s. %s", step, fix))
			previous = ""
			continue
		}
		commit, ok := ledgerCommit(r.cells[1])
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the commit of the step %s is not an ancestor of HEAD. %s", r.line, step, fix))
		case previous != "" && !ancestor(previous, commit):
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the step %s comes before the step that precedes it in round 1. Fix: do the steps of round 1 in the order of the skill", r.line, step))
		case step == "Handler" && commit == previous:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the handler and the first scenario are in one commit, so the scenario never failed first. Fix: commit the failing scenario before the handler", r.line))
		}
		previous = commit
	}
	return problems
}

// ledgerCommit resolves a commit of the ledger. It returns its full SHA, and
// whether it is HEAD or an ancestor of HEAD.
func ledgerCommit(cell string) (string, bool) {
	m := ledgerSHA.FindStringSubmatch(strings.TrimSpace(cell))
	if m == nil {
		return "", false
	}
	out, err := exec.Command("git", "rev-parse", "--verify", "--quiet", m[1]+"^{commit}").Output()
	sha := strings.TrimSpace(string(out))
	return sha, err == nil && ancestor(sha, "HEAD")
}

func ancestor(commit, of string) bool {
	return exec.Command("git", "merge-base", "--is-ancestor", commit, of).Run() == nil
}

// ledgerRounds counts the rounds of the ledger, for the note of finish.
func ledgerRounds() int {
	data, err := os.ReadFile("readiness.md")
	if err != nil {
		return 0
	}
	n := 0
	for _, l := range SpecLines(string(data)) {
		if l.Part == "Rounds" && l.Kind == SpecRow {
			n++
		}
	}
	return n
}
