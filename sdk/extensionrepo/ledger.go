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
	rounds, order, problems := ledgerTables(readiness)
	if len(rounds) == 0 {
		return append(problems, "readiness.md: it holds no round under ## Rounds. "+fix)
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
	sizes, ended := builds(rounds)
	for b, size := range sizes {
		if size > maxRounds {
			problems = append(problems, fmt.Sprintf("readiness.md: build %d of the ledger holds %d rounds, and a build has at most %d. "+
				"Fix: stop, and report each open line and each open finding to the user. Do not start another round", b+1, size, maxRounds))
		}
	}
	reviews := committedReviews()
	problems = append(problems, orderProblems(order, ended)...)
	problems = append(problems, stateProblems(findings, len(rounds), spec, limits)...)
	problems = append(problems, passProblems(rounds, reviews, spec, limits)...)
	return append(problems, historyProblems(findings, rounds, reviews)...)
}

// builds splits the rounds into builds. A round before the last whose Finish
// is pass ends a build, and the next round starts the next one. It returns
// the number of rounds of each build, and the head of the last round of the
// build before the current one, or "".
func builds(rounds []round) ([]int, string) {
	sizes, ended := []int{0}, ""
	for i, r := range rounds {
		sizes[len(sizes)-1]++
		if i < len(rounds)-1 && len(r.cells) == len(roundColumns) && strings.Trim(r.cells[5], "`") == "pass" {
			sizes, ended = append(sizes, 0), r.cells[1]
		}
	}
	return sizes, ended
}

// ledgerTables reads the tables of ## Rounds and ## Order of round 1. Each
// part ends at the next heading of level 1 or 2. Any other line in it fails,
// so no row hides behind a label or a heading.
func ledgerTables(readiness string) ([]round, map[string]round, []string) {
	const fix = "Fix: give readiness.md a table under ## Rounds with the columns " +
		"Round, Head, Change review, Change commits, Readiness review, and Finish, one row for each round"
	var problems []string
	var rounds []round
	order := map[string]round{}
	lines, part := strings.Split(readiness, "\n"), ""
	for _, l := range SpecLines(readiness) {
		if h := readinessHeading.FindStringSubmatch(lines[l.Number-1]); h != nil && len(h[1]) <= 2 {
			part = h[2]
			continue
		}
		if part != "Rounds" && part != "Order of round 1" {
			continue
		}
		switch {
		case l.Kind == SpecBlank || l.Kind == SpecSeparator:
		case l.Kind == SpecHeader && part == "Rounds" && !slices.Equal(l.Cells, roundColumns):
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the table of ## Rounds has other columns. %s", l.Number, fix))
		case l.Kind == SpecHeader:
		case l.Kind == SpecRow && part == "Rounds":
			rounds = append(rounds, round{l.Number, l.Cells})
		case l.Kind == SpecRow && len(l.Cells) == 2:
			order[l.Cells[0]] = round{l.Number, l.Cells}
		default:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: this line of ## %s is no row of its table. Fix: keep only the table in this part", l.Number, part))
		}
	}
	return rounds, order, problems
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
func orderProblems(order map[string]round, ended string) []string {
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
		case step == "Intake" && ended != "" && ancestor(commit, strings.Trim(ended, "`")):
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the intake of this build comes before the end of the build before it, at %s. Fix: record the intake of this build in a new commit", r.line, ended))
		case previous != "" && !ancestor(previous, commit):
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the step %s comes before the step that precedes it in round 1. Fix: do the steps of round 1 in the order of the skill", r.line, step))
		case step == "Handler" && commit == previous:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the handler and the first scenario are in one commit, so the scenario never failed first. Fix: commit the failing scenario before the handler", r.line))
		case commit == previous:
			problems = append(problems, fmt.Sprintf("readiness.md:%d: the step %s is in the commit of the step before it. Fix: commit each of the first four steps of round 1 on its own", r.line, step))
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

// ledgerRounds counts the rounds of the current build, for the note of finish.
func ledgerRounds() int {
	data, err := os.ReadFile("readiness.md")
	if err != nil {
		return 0
	}
	rounds, _, _ := ledgerTables(string(data))
	sizes, _ := builds(rounds)
	return sizes[len(sizes)-1]
}
