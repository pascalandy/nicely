package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBacklogDiagnosticOwnership(t *testing.T) {
	for _, tc := range []struct{ name, old, replacement, want string }{
		{"invalid frontmatter", "core: release checks", "core: ''", "frontmatter keys must"},
		{"broken link", "[Demo](#demo)", "[Missing](missing.md)", "the link missing.md points to a missing file"},
		{"unknown card", "It checks the release.", "It checks COR-101-T9.", "no item defines the card COR-101-T9"},
		{"typography", "It checks the release.", "It checks\u2014the release.", "the prose holds \u2014"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := backlogRepository(t, edit(readyItem, tc.old, tc.replacement), "1. COR-101\n")
			var out strings.Builder
			if err := run([]string{"check", root}, &out); err == nil {
				t.Fatal("check accepted the failing fixture")
			}
			if strings.Count(out.String(), "\n") != 1 || !strings.Contains(out.String(), tc.want) {
				t.Fatalf("want one diagnostic containing %q, got:\n%s", tc.want, out.String())
			}
		})
	}
}

func TestBacklogCardStates(t *testing.T) {
	done := edit(readyItem, "| - | todo |", "| - | done |")
	for _, tc := range []struct{ name, text, queue, want string }{
		{"Parked done", edit(done, "Ready.", "Parked."), "", "a Parked item has no started card"},
		{"Planned done", edit(done, "Ready.", "Planned."), "1. COR-101\n", "a Planned item needs open questions or todo cards, and no done card"},
		{"Planned empty", strings.Split(edit(readyItem, "Ready.", "Planned."), "## Cards")[0], "1. COR-101\n", "a Planned item needs open questions or todo cards"},
		{"Ready empty", strings.Split(readyItem, "## Cards")[0], "1. COR-101\n", "a Ready item needs todo cards"},
		{"Active all done", edit(done, "Ready.", "Active."), "1. COR-101\n", "an Active item needs a done card and a todo card"},
	} {
		t.Run(tc.name, func(t *testing.T) { assertRejected(t, backlogRepository(t, tc.text, tc.queue), tc.want) })
	}
	active := edit(done, "Ready.", "Active.")
	active = edit(active, "### COR-101-T1 Check", "| COR-101-T2 | Follow up | agent | COR-101-T1 | todo |\n\n### COR-101-T1 Check")
	active += "\n### COR-101-T2 Follow up\n\n- **Read:** [Demo](#demo)\n- **Proves:** `proof.txt`\n"
	root := backlogRepository(t, active, "1. COR-101\n")
	write(t, filepath.Join(root, backlogDir, ".DS_Store"), "Finder data")
	write(t, filepath.Join(root, milestoneDir, ".DS_Store"), "Finder data")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("valid Active item: %v\n%s", err, out.String())
	}
}
