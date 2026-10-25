package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNextUsesQueueOrder(t *testing.T) {
	root := backlogRepository(t, readyItem, "1. XT-102\n2. COR-101\n")
	transcript := strings.NewReplacer("COR-101", "XT-102", "core:", "extension:", "release checks", "transcript").Replace(readyItem)
	write(t, filepath.Join(root, backlogDir, "XT-102-transcript.md"), transcript)
	var out strings.Builder
	if err := run([]string{"next", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Next card: XT-102-T1 Check") || strings.Contains(out.String(), "Next card: COR-101") {
		t.Fatalf("next did not follow the queue:\n%s", out.String())
	}
}

func TestStatusFromFrontmatter(t *testing.T) {
	text := strings.ReplaceAll(readyItem, "# COR-101 Release checks", "# A heading unrelated to the display name")
	root := backlogRepository(t, text, "1. COR-101\n")
	write(t, filepath.Join(root, backlogDir, "XT-601-ignite.md"), ideaItem)
	var out strings.Builder
	if err := run([]string{"status", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	want := "XT\nID      Extension  State\nXT-601  ignite     Idea. Revisit when a project needs it\n\nCOR\nID       Core            State\nCOR-101  release checks  Ready. Next comes the check\n"
	if out.String() != want {
		t.Fatalf("status =\n%s\nwant =\n%s", out.String(), want)
	}
}

func TestFrozenMilestonesAreOnlyLinkChecked(t *testing.T) {
	root := backlogRepository(t, readyItem, "1. COR-101\n")
	name := filepath.Join(root, milestoneDir, "M00-foundation.md")
	write(t, name, "# History\n\nStatus: nonsense\nVersion: no version\nM00-T999 is history.\n")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("history was parsed as a plan: %v\n%s", err, out.String())
	}
	write(t, name, "# History\n\n[Missing](gone.md)\n")
	assertRejected(t, root, "the link gone.md points to a missing file")
}

func TestStateEscapesRespectProse(t *testing.T) {
	for _, tc := range []struct{ name, state, want string }{
		{"long dash", `state: "Ready. Next\u2014comes"`, "the prose holds \u2014"},
		{"curly quote", `state: "Ready. Next\u2019comes"`, "the prose holds \u2019"},
		{"newline", `state: "Ready. Next\ncomes"`, "nonempty strings on one line"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := edit(readyItem, "state: Ready. Next comes the check", tc.state)
			assertRejected(t, backlogRepository(t, text, "1. COR-101\n"), tc.want)
		})
	}
}

func TestItemBodyMetadataExample(t *testing.T) {
	text := readyItem + "\n## Later\n\n```yaml\nStatus: old example\nVersion: old example\n```\n"
	root := backlogRepository(t, text, "1. COR-101\n")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("metadata example: %v\n%s", err, out.String())
	}
}
