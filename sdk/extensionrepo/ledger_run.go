package extensionrepo

import (
	"fmt"
	"os/exec"
	"strings"
)

// firstRunProblems checks the record of the first failing run of the current
// build (#420). Its commit is the Scenario commit of [order], so a record of
// an earlier build never stands in. Its scenario is a scenario of that
// commit, its exit status is not 0, and its assertion holds a line of that
// scenario. just signoff runs the scenario again, and the reviewer judges
// whether the assertion proves the missing behavior (#422).
func firstRunProblems(run *firstRun, o order) []string {
	const fix = "Fix: run the first scenario at its commit with just scenario <name>, before the handler, and record that run in the table [first_run] " +
		"with the keys commit, scenario, exit, and assertion"
	if run == nil {
		return []string{"review.toml: it records no first failing run, so no record shows that the first scenario of round 1 failed before the handler. " + fix}
	}
	var problems []string
	recorded, _ := ledgerCommit(o.Scenario)
	commit, _ := ledgerCommit(run.Commit)
	if commit == "" || commit != recorded {
		problems = append(problems, fmt.Sprintf("review.toml: the first run names the commit %q, which is not the Scenario commit of [order]. %s", run.Commit, fix))
	}
	switch {
	case run.Scenario == "":
		problems = append(problems, "review.toml: the first run names no scenario. "+fix)
	case !scenarioName.MatchString(run.Scenario):
		problems = append(problems, fmt.Sprintf("review.toml: the first run names the scenario %q, which is no name of a scenario. "+
			"just signoff runs only just scenario <name> again, never the text of the record. %s", run.Scenario, fix))
	case commit != "":
		problems = append(problems, scenarioProblems(commit, run.Scenario, run.Assertion, fix)...)
	}
	if run.Exit == 0 {
		problems = append(problems, "review.toml: the first run of the scenario exited 0, so it did not fail. "+fix)
	}
	if strings.TrimSpace(strings.TrimSuffix(run.Assertion, ".")) == "" {
		problems = append(problems, "review.toml: the first run names no assertion that failed. "+
			"A compile error is no failed assertion, because it proves no missing behavior. "+fix)
	}
	return problems
}

// scenarioProblems checks that the scenario name exists at commit, and that
// the assertion holds one of its lines.
func scenarioProblems(commit, name, assertion, fix string) []string {
	file := "testdata/script/" + name + ".txtar"
	script, err := exec.Command("git", "show", commit+":"+file).Output()
	if err != nil {
		return []string{fmt.Sprintf("review.toml: the first run runs the scenario %s, which the Scenario commit %s does not hold. %s", file, commit, fix)}
	}
	if strings.TrimSpace(assertion) == "" || scenarioLine(string(script), assertion) == 0 {
		return []string{fmt.Sprintf("review.toml: the first run names the assertion %q, which holds no line of %s at %s. "+
			"Fix: copy the line of the scenario that failed, as the output of the run names it", assertion, file, commit)}
	}
	return nil
}

// scenarioLine returns the number of the first line of the script of a
// scenario, before its files, that text holds. It skips blank lines and
// comments. It returns 0 when text holds no such line.
func scenarioLine(script, text string) int {
	for i, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- ") && strings.HasSuffix(line, " --") {
			break
		}
		if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(text, line) {
			return i + 1
		}
	}
	return 0
}
