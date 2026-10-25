package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// checkDecisions reports a decision file whose name, title, or number breaks
// the rules of decisions/README.md, or that has no row in that index.
func checkDecisions(root string) []string {
	entries, err := os.ReadDir(filepath.Join(root, decisionDir))
	if err != nil {
		return nil
	}
	index, _ := os.ReadFile(filepath.Join(root, decisionDir, "README.md"))
	var problems []string
	seen := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "README.md" {
			continue
		}
		where := decisionDir + "/" + e.Name()
		match := decisionName.FindStringSubmatch(e.Name())
		if match == nil {
			problems = append(problems, where+": the file name must look like D001-short-title.md")
			continue
		}
		id := match[1]
		if other, taken := seen[id]; taken {
			problems = append(problems, fmt.Sprintf("%s: %s uses the same number", where, other))
		}
		seen[id] = e.Name()
		data, err := os.ReadFile(filepath.Join(root, decisionDir, e.Name()))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", where, err))
			continue
		}
		if !strings.HasPrefix(string(data), "# "+id+" ") {
			problems = append(problems, fmt.Sprintf("%s:1: the first line must be \"# %s <title>\"", where, id))
		}
		if !strings.Contains(string(index), "]("+e.Name()+")") {
			problems = append(problems, fmt.Sprintf("%s: the decision needs a row in %s/README.md", where, decisionDir))
		}
	}
	return problems
}
