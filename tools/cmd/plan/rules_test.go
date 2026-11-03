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

// TestRepositoryRules requires both indexes of D068 in this repository.
// TestRepositoryPlan checks their problems.
func TestRepositoryRules(t *testing.T) {
	indexes, _, err := ruleIndexes("../../..")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, index := range indexes {
		if index.rules == 0 {
			t.Errorf("the %s index holds no rule", index.name)
		}
		names = append(names, index.name)
	}
	if !slices.Equal(names, []string{"core", "template"}) {
		t.Fatalf("indexes = %v, want core and template", names)
	}
}

const coreFixture = `# Rules

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| G01 | Read the files | [AGENTS.md][agents] | Prose only |
| G02 | Run the check | [AGENTS.md][agents] | [Check][check], partial: one path |
| G03 | Judge the job | [AGENTS.md][agents] | Prose on purpose: only a reader judges it |

[agents]: ../../../AGENTS.md
[check]: ../../../check.sh
`

// TestRuleIndexRejects mutates a valid core index, and requires the problem
// of each mutation.
func TestRuleIndexRejects(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "AGENTS.md"), "# Agents\n")
	write(t, filepath.Join(root, "check.sh"), "exit 0\n")
	write(t, filepath.Join(root, "sdk", "extensionrepo", "gate.go"), "package extensionrepo\n")
	write(t, filepath.Join(root, "sdk", "extensionrepo", "help.go"), "package extensionrepo\n")
	file := filepath.Join(root, filepath.FromSlash(coreRules), "session.md")
	readme := filepath.Join(root, filepath.FromSlash(coreRules), "README.md")
	valid := edit(coreFixture, "[Check][check], partial", "[Check][check], [gate][gate], partial") + "[gate]: ../../../sdk/extensionrepo/gate.go\n"
	write(t, readme, "# Core rule index\n\n## Support code\n\n| File | What it does |\n|---|---|\n| [help.go](../../../sdk/extensionrepo/help.go) | Helps |\n")
	write(t, file, valid)
	if problems, err := checkRuleIndexes(root); err != nil || len(problems) > 0 {
		t.Fatalf("the valid index failed: %v %q", err, problems)
	}
	for _, c := range []struct{ name, old, new, want string }{
		{"no link", "[Check][check], [gate][gate], partial: one path", "the check", `docs/north-star/rules/session.md:6: the enforcer "the check" links no file`},
		{"empty", "[Check][check], [gate][gate], partial: one path", "", `docs/north-star/rules/session.md:6: the enforcer "" links no file`},
		{"missing file", "[check]: ../../../check.sh", "[check]: ../../../gone.sh", "session.md:6: the link ../../../gone.sh reaches no file"},
		{"folder", "[check]: ../../../check.sh", "[check]: ../../../sdk", "session.md:6: the link ../../../sdk reaches no file"},
		{"outside", "[check]: ../../../check.sh", "[check]: https://example.com/check.sh", "session.md:6: the link https://example.com/check.sh leaves this repository"},
		{"no definition", "[Check][check]", "[Check][nothing]", "session.md:6: the reference [nothing] has no definition"},
		{"source", "| G01 | Read the files | [AGENTS.md][agents]", "| G01 | Read the files | [AGENTS.md][gone]", "session.md:5: the reference [gone] has no definition"},
		{"no judgment", "Prose on purpose: only a reader judges it", "Prose on purpose", "session.md:7: write the judgment that keeps the rule in prose"},
		{"empty judgment", "Prose on purpose: only a reader judges it", "Prose on purpose: ", "session.md:7: write the judgment"},
		{"duplicate", "| G03 |", "| G01 |", "session.md:7: docs/north-star/rules/session.md:5 also names the rule G01"},
		{"cells", "| G03 | Judge the job |", "| G03 |", "session.md:7: the row has 3 cells, and its header 4"},
		{"unnamed gate", "[Check][check], [gate][gate], partial", "[Check][check], partial", "sdk/extensionrepo/gate.go: no row of a rule index links this file"},
		{"no rules", "| Enforced by |", "| Checks |", "session.md: the core index holds no row under an Enforced by column"},
	} {
		t.Run(c.name, func(t *testing.T) {
			write(t, file, edit(valid, c.old, c.new))
			problems, err := checkRuleIndexes(root)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, c.want) }) {
				t.Fatalf("want %q, got %q", c.want, problems)
			}
		})
	}
	write(t, readme, "# Core rule index\n")
	write(t, file, valid)
	problems, _ := checkRuleIndexes(root)
	if !slices.Equal(problems, []string{"sdk/extensionrepo/help.go: no row of a rule index links this file. Link it from the rules that it enforces, or list it under Support code in docs/north-star/rules/README.md"}) {
		t.Fatalf("a gate file outside Support code gave %q", problems)
	}
}

// TestStatusCountsRules requires the count of each index by enforcer.
func TestStatusCountsRules(t *testing.T) {
	root := backlogRepository(t, readyItem, "1. COR-101\n")
	write(t, filepath.Join(root, "AGENTS.md"), "# Agents\n")
	write(t, filepath.Join(root, "check.sh"), "exit 0\n")
	write(t, filepath.Join(root, filepath.FromSlash(coreRules), "session.md"), coreFixture)
	write(t, filepath.Join(root, filepath.FromSlash(templateIndex)), "| Rule | Enforced by |\n|---|---|\n| One | Prose only; [check](../../check.sh) |\n| Two | Prose only |\n")
	write(t, filepath.Join(root, filepath.FromSlash(templateParts), "loop.md.jinja"), "| Rule | Enforced by |\n|---|---|\n| Three | Prose on purpose: a reader judges it |\n")
	var out strings.Builder
	if err := run([]string{"status", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	want := "\nRule indexes\nIndex     Rules  Prose only  Prose on purpose\ncore      3      1           1\ntemplate  3      2           1\n"
	if !strings.HasSuffix(out.String(), want) {
		t.Fatalf("status =\n%s\nwant the suffix\n%s", out.String(), want)
	}
}
