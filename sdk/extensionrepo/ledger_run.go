package extensionrepo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// firstRun finds a line of ## First failing run, in the bold forms that
// readinessPart reads. That part records the run of the first scenario of
// round 1 at its commit, before the handler (#420).
var firstRun = regexp.MustCompile(`^\s*[-*+]\s+\*\*(Commit|Command|Exit|Assertion)(?:\*\*:|:\*\*)\s*(.*?)\s*$`)

// firstRunProblems checks the record of the first failing run of the current
// build. It needs the Scenario commit of ## Order of round 1, so a record of
// an earlier build never stands in. It also needs the command just scenario
// <name> of a scenario of that commit, an exit status that is not 0, and the
// assertion that failed, which holds a line of that scenario. just signoff
// runs the scenario again, and the reviewer judges whether the assertion
// proves the missing behavior (#422).
func firstRunProblems(readiness string, order map[string]round) []string {
	const fix = "Fix: run the first scenario at its commit with just scenario <name>, before the handler, and record that run under ## First failing run " +
		"in the lines - **Commit**:, - **Command**:, - **Exit**:, and - **Assertion**:"
	parts, found := firstRunRecord(readiness)
	if !found {
		return []string{"readiness.md: it holds no part ## First failing run, so no record shows that the first scenario of round 1 failed before the handler. " + fix}
	}
	var problems []string
	scenario, recorded := order["Scenario"], ""
	if len(scenario.cells) == 2 {
		recorded, _ = ledgerCommit(scenario.cells[1])
	}
	commit, _ := ledgerCommit(parts["Commit"])
	if commit == "" || commit != recorded {
		problems = append(problems, fmt.Sprintf("readiness.md: ## First failing run names the commit %q, which is not the Scenario commit of ## Order of round 1. %s", parts["Commit"], fix))
	}
	if parts["Command"] == "" {
		problems = append(problems, "readiness.md: ## First failing run names no command. "+fix)
	} else if name := scenarioRecipe.FindStringSubmatch(parts["Command"]); name == nil {
		problems = append(problems, fmt.Sprintf("readiness.md: ## First failing run names the command %q, which is not just scenario <name>. "+
			"just signoff runs only that recipe again, never the text of the record. %s", parts["Command"], fix))
	} else if commit != "" {
		problems = append(problems, scenarioProblems(commit, name[1], parts["Assertion"], fix)...)
	}
	switch exit, err := strconv.Atoi(parts["Exit"]); {
	case err != nil:
		problems = append(problems, fmt.Sprintf("readiness.md: the exit status %q of the first run is no number. %s", parts["Exit"], fix))
	case exit == 0:
		problems = append(problems, "readiness.md: the first run of the scenario exited 0, so it did not fail. "+fix)
	}
	if strings.TrimSpace(strings.TrimSuffix(parts["Assertion"], ".")) == "" {
		problems = append(problems, "readiness.md: ## First failing run names no assertion that failed. "+
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
		return []string{fmt.Sprintf("readiness.md: ## First failing run runs the scenario %s, which the Scenario commit %s does not hold. %s", file, commit, fix)}
	}
	if strings.TrimSpace(assertion) == "" || scenarioLine(string(script), assertion) == 0 {
		return []string{fmt.Sprintf("readiness.md: ## First failing run names the assertion %q, which holds no line of %s at %s. "+
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

// firstRunRecord reads the lines of ## First failing run, outside fenced
// code. It also says whether readiness.md holds that part.
func firstRunRecord(readiness string) (map[string]string, bool) {
	parts, at, found, fence := map[string]string{}, 0, false, ""
	for i, text := range strings.Split(readiness, "\n") {
		// A fence closes with its own run of marks, alone on its line, as in readinessFindings.
		switch trimmed := strings.TrimSpace(text); {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			}
			continue
		case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, "~~~"):
			fence = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1]))]
			continue
		}
		if h := readinessHeading.FindStringSubmatch(text); h != nil && len(h[1]) <= 2 {
			found = found || h[2] == "First failing run"
			at = 0
			if h[2] == "First failing run" {
				at = i + 1
			}
			continue
		}
		if m := firstRun.FindStringSubmatch(text); at > 0 && m != nil {
			parts[m[1]] = strings.Trim(strings.TrimSuffix(m[2], "."), "`")
		}
	}
	return parts, found
}
