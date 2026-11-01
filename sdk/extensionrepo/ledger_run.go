package extensionrepo

import (
	"fmt"
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
// an earlier build never stands in. It also needs the command, an exit status
// that is not 0, and the assertion that failed. The reviewer judges whether
// that assertion proves the missing behavior, so a compile error is no proof.
func firstRunProblems(readiness string, order map[string]round) []string {
	const fix = "Fix: run the first scenario at its commit, before the handler, and record that run under ## First failing run " +
		"in the lines - **Commit**:, - **Command**:, - **Exit**:, and - **Assertion**:"
	parts, at, found, fence := map[string]string{}, 0, false, ""
	for i, text := range strings.Split(readiness, "\n") {
		if trimmed := strings.TrimSpace(text); fence != "" || strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			switch {
			case fence == "":
				fence = trimmed[:3]
			case strings.HasPrefix(trimmed, fence):
				fence = ""
			}
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
			parts[m[1]] = strings.Trim(m[2], "`")
		}
	}
	if !found {
		return []string{"readiness.md: it holds no part ## First failing run, so no record shows that the first scenario of round 1 failed before the handler. " + fix}
	}
	var problems []string
	scenario, recorded := order["Scenario"], ""
	if len(scenario.cells) == 2 {
		recorded, _ = ledgerCommit(scenario.cells[1])
	}
	if commit, _ := ledgerCommit(parts["Commit"]); commit == "" || commit != recorded {
		problems = append(problems, fmt.Sprintf("readiness.md: ## First failing run names the commit %q, which is not the Scenario commit of ## Order of round 1. %s", parts["Commit"], fix))
	}
	if parts["Command"] == "" {
		problems = append(problems, "readiness.md: ## First failing run names no command. "+fix)
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
