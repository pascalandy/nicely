package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const readyItem = `---
id: COR-101
core: release checks
state: Ready. Next comes the check
---

# COR-101 Release checks

## Demo

It checks the release.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-101-T1 | Check | agent | - | todo |

### COR-101-T1 Check

- **Read:** [Demo](#demo)
- **Proves:** ` + "`proof.txt`" + `
`

func backlogRepository(t *testing.T, text, queue string) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, backlogDir, "COR-101-release-checks.md"), text)
	write(t, filepath.Join(root, backlogDir, "README.md"), "# Backlog\n\n## Queue\n\n"+queue)
	write(t, filepath.Join(root, "proof.txt"), "proof\n")
	return root
}

func TestBacklogRejects(t *testing.T) {
	cases := []struct{ name, old, replacement, want string }{
		{"missing frontmatter", "---\nid: COR-101\ncore: release checks\nstate: Ready. Next comes the check\n---", "", "frontmatter needs exactly three keys"},
		{"missing key", "core: release checks\n", "", "frontmatter needs exactly three keys"},
		{"extra key", "state: Ready. Next comes the check", "state: Ready\nextra: value", "frontmatter needs exactly three keys"},
		{"misordered keys", "id: COR-101\ncore: release checks", "core: release checks\nid: COR-101", "frontmatter keys must be id, core, state, in that order"},
		{"wrong ownership key", "core: release checks", "extension: release checks", "frontmatter keys must be id, core, state"},
		{"duplicate key", "core: release checks", "id: COR-101", "frontmatter keys must be id, core, state"},
		{"empty value", "core: release checks", "core: ''", "nonempty strings on one line"},
		{"collection value", "core: release checks", "core: [release, checks]", "nonempty strings on one line"},
		{"nonstring value", "core: release checks", "core: 42", "nonempty strings on one line"},
		{"invalid YAML", "core: release checks", "core: [", "invalid YAML frontmatter"},
		{"ID differs", "id: COR-101", "id: COR-102", "frontmatter id must match the file ID COR-101"},
		{"unknown state", "Ready. Next", "Waiting. Next", "state must start with Idea, Parked, Planned, Ready, Active, or Shipped"},
		{"status later in state", "Ready. Next comes the check", "Next is Ready", "state must start with"},
		{"two-sentence demo", "It checks the release.", "It checks. It releases.", "a one-sentence Demo"},
		{"ready done", "| - | todo |", "| - | done |", "a Ready item needs todo cards, no done card, and no open questions"},
		{"ready questions", "## Cards", "## Open questions\n\nWhich check?\n\n## Cards", "a Ready item needs todo cards, no done card, and no open questions"},
		{"parked done", "state: Ready", "state: Parked", "queue item COR-101 must be Planned, Ready, or Active"},
		{"active no done", "state: Ready", "state: Active", "an Active item needs a done card and a todo card"},
		{"shipped todo", "state: Ready", "state: Shipped", "a Shipped item has no todo card"},
		{"idea chosen band", "state: Ready", "state: Idea", "Idea needs a 6xx ID"},
		{"wrong card ID", "COR-101-T1", "COR-102-T1", "the card \"COR-102-T1\" must be named COR-101-T<number>"},
		{"malformed card ID", "COR-101-T1", "COR-101-1", "must be named COR-101-T<number>"},
		{"unknown card reference", "It checks the release.", "It checks COR-101-T9.", "no item defines the card COR-101-T9"},
		{"state long dash", "Next comes", "Next\u2014comes", "the prose holds \u2014"},
		{"state curly quote", "Next comes", "Next\u2019comes", "the prose holds \u2019"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := backlogRepository(t, strings.ReplaceAll(readyItem, tc.old, tc.replacement), "1. COR-101\n")
			assertRejected(t, root, tc.want)
		})
	}
}

func TestBacklogRejectsFilesAndQueue(t *testing.T) {
	cases := []struct{ name, file, text, queue, want string }{
		{"new milestone", "docs/milestones/M26-new.md", "# M26 New\n", "1. COR-101\n", "only the 16 frozen milestone files may remain"},
		{"no keywords", "docs/backlog/XT-601.md", ideaItem, "1. COR-101\n", "1 to 3 lowercase keywords"},
		{"four keywords", "docs/backlog/XT-601-one-two-three-four.md", ideaItem, "1. COR-101\n", "1 to 3 lowercase keywords"},
		{"uppercase keyword", "docs/backlog/XT-601-Other.md", ideaItem, "1. COR-101\n", "1 to 3 lowercase keywords"},
		{"invalid band", "docs/backlog/XT-201-idea.md", strings.ReplaceAll(ideaItem, "XT-601", "XT-201"), "1. COR-101\n", "the ID must be in the 1xx or 6xx band"},
		{"idea band not Idea", "docs/backlog/XT-601-idea.md", edit(ideaItem, "Idea.", "Parked."), "1. COR-101\n", "a 6xx item must have status Idea"},
		{"idea cards", "docs/backlog/XT-601-idea.md", strings.NewReplacer("COR-101", "XT-601", "core:", "extension:", "Ready.", "Idea.").Replace(readyItem), "1. COR-101\n", "an Idea item has no cards"},
		{"missing queue item", "", "", "1. COR-199\n", "queue names no item COR-199"},
		{"ineligible queue item", "docs/backlog/XT-601-idea.md", ideaItem, "1. COR-101\n2. XT-601\n", "queue item XT-601 must be Planned, Ready, or Active"},
		{"omitted queue item", "", "", "", "a Planned, Ready, or Active item must be in the queue"},
		{"duplicate queue item", "", "", "1. COR-101\n2. COR-101\n", "duplicate queue entry COR-101"},
		{"malformed queue entry", "", "", "- COR-101\n", "a queue entry must look like 1. COR-101"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := backlogRepository(t, readyItem, tc.queue)
			if tc.file != "" {
				write(t, filepath.Join(root, tc.file), tc.text)
			}
			assertRejected(t, root, tc.want)
		})
	}
}

const ideaItem = `---
id: XT-601
extension: ignite
state: Idea. Revisit when a project needs it
---

# XT-601 Ignite

## Demo

It starts a project.
`

func assertRejected(t *testing.T, root, want string) {
	t.Helper()
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err == nil {
		t.Fatalf("check accepted the failing fixture:\n%s", out.String())
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("check lacks %q:\n%s", want, out.String())
	}
	t.Log(strings.TrimSpace(out.String()))
}

func TestBacklogStates(t *testing.T) {
	cases := []struct{ name, text, queue string }{
		{"Ready", readyItem, "1. COR-101\n"},
		{"Planned cards", edit(readyItem, "Ready.", "Planned."), "1. COR-101\n"},
		{"Planned questions without cards", strings.Split(edit(readyItem, "Ready.", "Planned."), "## Cards")[0] + "## Open questions\n\nWhich check?\n", "1. COR-101\n"},
		{"Parked", edit(readyItem, "Ready.", "Parked.") + "\n## Open questions\n\nWhich check?\n", ""},
		{"Shipped", strings.NewReplacer("Ready.", "Shipped.", "| - | todo |", "| - | done |").Replace(readyItem), ""},
		{"Shipped external without cards", strings.Split(edit(readyItem, "Ready.", "Shipped."), "## Cards")[0], ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := backlogRepository(t, tc.text, tc.queue)
			write(t, filepath.Join(root, backlogDir, "XT-601-ignite.md"), ideaItem)
			var out strings.Builder
			if err := run([]string{"check", root}, &out); err != nil {
				t.Fatalf("%v\n%s", err, out.String())
			}
		})
	}
}
