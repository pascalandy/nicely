package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// indexAt7f3ae60e holds the ledger rows of the template rule index at
// 7f3ae60e. That commit added ledger_history.go, and no row named it until
// b3dc8b48 added the row of indexAtb3dc8b48.
const indexAt7f3ae60e = `# Rule index

| Rule | Enforced by |
|---|---|
| The ledger names the head, the reviewers, the commits that each review covered, and the result of ` + "`just finish`" + ` of each round | [Ledger][ledger] |
| Each finding keeps its ID and its highest severity, and gets a state each round; a ` + "`warn` or `block` closes only as `fixed` or `limit`" + ` | [Ledger findings][ledger-findings] |

[ledger]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger.go
[ledger-findings]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_findings.go
`

var indexAtb3dc8b48 = edit(edit(indexAt7f3ae60e,
	"| [Ledger findings][ledger-findings] |\n",
	"| [Ledger findings][ledger-findings], [ledger history][ledger-history] |\n| Each round keeps its row and has a committed review; a build ends only at a review that passes | [Ledger history][ledger-history] |\n"),
	"[ledger-findings]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_findings.go\n",
	"[ledger-findings]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_findings.go\n[ledger-history]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_history.go\n")

// TestRuleIndexNamesEachGate reproduces 7f3ae60e: a gate file enforced rules
// that no row of the template index named.
func TestRuleIndexNamesEachGate(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"ledger.go", "ledger_findings.go", "ledger_history.go", "ledger_test.go"} {
		write(t, filepath.Join(root, "sdk", "extensionrepo", name), "package extensionrepo\n")
	}
	index := filepath.Join(root, filepath.FromSlash(templateIndex))
	write(t, index, indexAt7f3ae60e)
	problems, err := checkRuleIndexes(root)
	if err != nil {
		t.Fatal(err)
	}
	want := "sdk/extensionrepo/ledger_history.go: no row of a rule index links this file"
	if !slices.ContainsFunc(problems, func(p string) bool { return strings.HasPrefix(p, want) }) {
		t.Fatalf("the index of 7f3ae60e passed, want %q, got %q", want, problems)
	}
	write(t, index, indexAtb3dc8b48)
	if problems, err := checkRuleIndexes(root); err != nil || len(problems) > 0 {
		t.Fatalf("the index of b3dc8b48 failed: %v %q", err, problems)
	}
}
