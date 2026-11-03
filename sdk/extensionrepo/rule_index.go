package extensionrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ruleIndexFile is the rule index of an extension repository, and
// ruleIndexParts holds its other parts. D068 counts its rules whose enforcer
// starts with "Prose only", as measure 1 for owner C.
const (
	ruleIndexFile  = "AGENTS/rule-index.md"
	ruleIndexParts = "AGENTS/rule-index"
)

// printRuleIndex writes the count of the rule index of the current folder to
// stderr. A repository without the file gets a notice, and check goes on.
func printRuleIndex() error {
	data, err := os.ReadFile(ruleIndexFile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s: missing, so no rule is counted\n", ruleIndexFile)
		return nil
	}
	if err != nil {
		return err
	}
	parts, err := filepath.Glob(filepath.Join(ruleIndexParts, "*.md"))
	if err != nil {
		return err
	}
	text := string(data)
	for _, part := range parts {
		data, err := os.ReadFile(part)
		if err != nil {
			return err
		}
		text += "\n" + string(data)
	}
	rules, proseOnly, onPurpose := countRules(text)
	fmt.Fprintf(os.Stderr, "%s: %d of %d rules are Prose only, and %d Prose on purpose\n", ruleIndexFile, proseOnly, rules, onPurpose)
	return nil
}

// countRules counts the rows of each table with an "Enforced by" column, and
// those whose enforcer starts with "Prose only" or "Prose on purpose".
func countRules(text string) (rules, proseOnly, onPurpose int) {
	var header []string
	for line := range strings.Lines(text) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			header = nil
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		column := slices.Index(header, "Enforced by")
		switch {
		case header == nil:
			header = cells
		case column < 0, strings.HasPrefix(line, "|---"), column >= len(cells):
		default:
			rules++
			switch {
			case strings.HasPrefix(cells[column], "Prose only"):
				proseOnly++
			case strings.HasPrefix(cells[column], "Prose on purpose"):
				onPurpose++
			}
		}
	}
	return rules, proseOnly, onPurpose
}
