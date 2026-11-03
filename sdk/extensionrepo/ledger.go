package extensionrepo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// maxRounds bounds the rounds of one build of a command, from the intake to
// "Done when" (#417).
const maxRounds = 5

// ledgerSHA is the form of a commit that the ledger names.
var ledgerSHA = regexp.MustCompile("^([0-9a-f]{7,40})$")

// ledgerProblems reports each rule of the ledger that the review record
// breaks. Its commit is the full SHA of the reviewed commit.
func ledgerProblems(review reviewRecord, spec string, limits []limit) []string {
	rounds := review.Rounds
	if len(rounds) == 0 {
		return []string{"review.toml: it holds no round. Fix: give it one table [[rounds]] for each round, with head, change_review, change_commits, readiness_review, and finish"}
	}
	var problems []string
	for i, r := range rounds {
		problems = append(problems, roundProblems(r, i+1, i == len(rounds)-1)...)
	}
	if head, ok := ledgerCommit(rounds[len(rounds)-1].Head); ok && head != review.Commit {
		problems = append(problems, fmt.Sprintf("review.toml: the last round names the head %s, and the key commit names %s. Fix: name the reviewed commit in both", rounds[len(rounds)-1].Head, review.Commit))
	}
	sizes, ended := builds(rounds)
	for b, size := range sizes {
		if size > maxRounds {
			problems = append(problems, fmt.Sprintf("review.toml: build %d of the ledger holds %d rounds, and a build has at most %d. "+
				"Fix: stop, and report each open line and each open finding to the user. Do not start another round", b+1, size, maxRounds))
		}
	}
	markdown := markdownEnd()
	reviews := committedReviews(markdown)
	problems = append(problems, orderProblems(review.Order, ended, markdown)...)
	problems = append(problems, firstRunProblems(review.FirstRun, review.Order)...)
	problems = append(problems, stateProblems(review.Findings, len(rounds), spec, limits)...)
	problems = append(problems, passProblems(rounds, reviews)...)
	return append(problems, historyProblems(review.Findings, rounds, reviews)...)
}

// builds splits the rounds into builds. A round before the last whose finish
// is pass ends a build, and the next round starts the next one. It returns
// the number of rounds of each build, and the head of the last round of the
// build before the current one, or "".
func builds(rounds []round) ([]int, string) {
	sizes, ended := []int{0}, ""
	for i, r := range rounds {
		sizes[len(sizes)-1]++
		if i < len(rounds)-1 && r.Finish == "pass" {
			sizes, ended = append(sizes, 0), r.Head
		}
	}
	return sizes, ended
}

// roundProblems checks one round, the round number of its place.
func roundProblems(r round, number int, last bool) []string {
	at := fmt.Sprintf("review.toml: round %d: ", number)
	var problems []string
	if !fullSHA.MatchString(r.Head) {
		problems = append(problems, at+"its head is no full SHA. Fix: write the 40 characters of the reviewed commit")
	}
	commits := []string{r.Head}
	for _, c := range strings.Split(r.ChangeCommits, "..") {
		commits = append(commits, strings.TrimSpace(c))
	}
	for _, c := range commits {
		if _, ok := ledgerCommit(c); !ok {
			problems = append(problems, fmt.Sprintf("%sthe commit %q is not an ancestor of HEAD. Fix: name the commits of this branch that the round reviewed, as <first>..<last>", at, c))
		}
	}
	if r.ChangeReview == "" || r.ReadinessReview == "" {
		problems = append(problems, at+"it names no reviewer of the change review or of the readiness review. Fix: name the agent that ran each review, which wrote none of its code")
	}
	want := "pass or fail"
	if last {
		want = "pending"
	}
	if last && r.Finish != "pending" || !last && r.Finish != "pass" && r.Finish != "fail" {
		problems = append(problems, fmt.Sprintf("%sits finish is %q. Fix: write %s. An earlier round holds the result of just finish, and the last round is pending, since this run gives its result", at, r.Finish, want))
	}
	return problems
}

// orderProblems checks that round 1 recorded the intake, the spec, the
// first scenario, and the handler, each in a commit after the one before.
// The scenario precedes the handler, so it could fail first. The intake
// comes after the end of the build before. It also comes after the last
// Markdown readiness.md, so the records start a new build after an update.
func orderProblems(o order, ended, markdown string) []string {
	const fix = "Fix: give review.toml a table [order] with the commits intake, spec, scenario, and handler"
	var problems []string
	previous := ""
	for _, step := range [][2]string{{"intake", o.Intake}, {"spec", o.Spec}, {"scenario", o.Scenario}, {"handler", o.Handler}} {
		if step[1] == "" {
			problems = append(problems, fmt.Sprintf("review.toml: [order] gives no commit for the step %s. %s", step[0], fix))
			previous = ""
			continue
		}
		commit, ok := ledgerCommit(step[1])
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("review.toml: the commit of the step %s is not an ancestor of HEAD. %s", step[0], fix))
		case step[0] == "intake" && ended != "" && ancestor(commit, ended):
			problems = append(problems, fmt.Sprintf("review.toml: the intake of this build comes before the end of the build before it, at %s. Fix: record the intake of this build in a new commit", ended))
		case step[0] == "intake" && markdown != "" && ancestor(commit, markdown):
			problems = append(problems, fmt.Sprintf("review.toml: the intake of this build comes before the last Markdown readiness.md, in %.12s, of a template before typed records. "+
				"Fix: start a new build with a new intake commit, and keep the Markdown reviews as history", markdown))
		case previous != "" && !ancestor(previous, commit):
			problems = append(problems, fmt.Sprintf("review.toml: the step %s comes before the step that precedes it in round 1. Fix: do the steps of round 1 in the order of the skill", step[0]))
		case step[0] == "handler" && commit == previous:
			problems = append(problems, "review.toml: the handler and the first scenario are in one commit, so the scenario never failed first. Fix: commit the failing scenario before the handler")
		case commit == previous:
			problems = append(problems, fmt.Sprintf("review.toml: the step %s is in the commit of the step before it. Fix: commit each of the first four steps of round 1 on its own", step[0]))
		}
		if step[0] == "spec" && ok && o.Intake != "" {
			if intake, found := ledgerCommit(o.Intake); found {
				problems = append(problems, intakeOrderProblems(intake, commit)...)
			}
		}
		previous = commit
	}
	return problems
}

// markdownEnd returns the newest commit of HEAD that changed a readiness.md
// beside no review.toml, or "". Such a readiness.md is a Markdown review of
// a template before typed records. Finish reads none of them.
func markdownEnd() string {
	log, err := exec.Command("git", "log", "--format=%H", "HEAD", "--", "readiness.md").Output()
	if err != nil {
		return ""
	}
	for _, revision := range strings.Fields(string(log)) {
		if exec.Command("git", "cat-file", "-e", revision+":"+reviewFile).Run() != nil {
			return revision
		}
	}
	return ""
}

// A resolved commit of the ledger. One run of finish resolves each value once,
// because the history repeats the same commits.
type resolved struct {
	sha string
	ok  bool
}

var resolvedCommits = map[string]resolved{}

// ledgerCommit resolves a commit of the ledger. It returns its full SHA, and
// whether it is HEAD or an ancestor of HEAD.
func ledgerCommit(value string) (string, bool) {
	m := ledgerSHA.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return "", false
	}
	if r, found := resolvedCommits[m[1]]; found {
		return r.sha, r.ok
	}
	out, err := exec.Command("git", "rev-parse", "--verify", "--quiet", m[1]+"^{commit}").Output()
	sha := strings.TrimSpace(string(out))
	r := resolved{sha, err == nil && ancestor(sha, "HEAD")}
	resolvedCommits[m[1]] = r
	return r.sha, r.ok
}

func ancestor(commit, of string) bool {
	return exec.Command("git", "merge-base", "--is-ancestor", commit, of).Run() == nil
}

// ledgerRounds counts the rounds of the current build, for the note of finish.
func ledgerRounds() int {
	review, found, problems := readReview()
	if !found || len(problems) != 0 {
		return 0
	}
	sizes, _ := builds(review.Rounds)
	return sizes[len(sizes)-1]
}
