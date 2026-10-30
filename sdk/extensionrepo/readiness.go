package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
)

var (
	readinessCommit  = regexp.MustCompile(`(?m)^Commit: ([0-9a-f]{7,40})\s*$`)
	readinessVerdict = regexp.MustCompile(`(?m)^Verdict: (.+?)\s*$`)
)

// readinessProblems reports a readiness.md without a passing verdict, or
// whose commit is not the code of the working tree. Only readiness.md may
// differ from that commit, so the review can be committed after it.
func readinessProblems() []string {
	const fix = "Fix: run the readiness review of the nicely-extension skill, then record it in readiness.md with the lines Commit: <sha> and Verdict: <verdict>"
	data, err := os.ReadFile("readiness.md")
	if errors.Is(err, os.ErrNotExist) {
		return []string{"readiness.md: no readiness verdict is recorded. " + fix}
	} else if err != nil {
		return []string{"readiness.md: " + err.Error()}
	}
	commit, verdict := readinessCommit.FindSubmatch(data), readinessVerdict.FindSubmatch(data)
	if commit == nil || verdict == nil {
		return []string{"readiness.md: it holds no line Commit: <sha> or Verdict: <verdict>. " + fix}
	}
	var problems []string
	switch v := string(verdict[1]); v {
	case "Ready", "Ready with stated limits":
	case "Needs work":
		problems = append(problems, "readiness.md: the verdict is Needs work. Fix: start the next round from its findings")
	default:
		problems = append(problems, fmt.Sprintf("readiness.md: the verdict %q is not Ready, Ready with stated limits, or Needs work. %s", v, fix))
	}
	sha := string(commit[1])
	if _, err := updateOutput("git", "rev-parse", "--verify", "--quiet", sha+"^{commit}"); err != nil {
		return append(problems, fmt.Sprintf("readiness.md: Git knows no commit %s. %s", sha, fix))
	}
	changed, err := updateOutput("git", "diff", "--name-only", sha, "--")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	added, err := updateOutput("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return append(problems, "readiness.md: "+err.Error())
	}
	var files []string
	for _, file := range strings.Split(string(changed)+string(added), "\n") {
		if file != "" && file != "readiness.md" && !slices.Contains(files, file) {
			files = append(files, file)
		}
	}
	if len(files) != 0 {
		problems = append(problems, fmt.Sprintf("readiness.md: the review is of %s, and these files changed since: %s. Fix: run the readiness review again on this commit", sha, strings.Join(files, ", ")))
	}
	return problems
}
