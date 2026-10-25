package main

import (
	"os"
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
	write(t, name, "# History\n\nStatus: nonsense\nVersion: no version\nCOR-101-T9 and it’s history.\n")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("history was parsed as a plan: %v\n%s", err, out.String())
	}
	write(t, name, "# History\n\n[Missing](gone.md)\n")
	assertRejected(t, root, "the link gone.md points to a missing file")
}

func TestReadyRule(t *testing.T) {
	data, err := os.ReadFile("../../../docs/north-star/dev-preferences.md")
	if err != nil {
		t.Fatal(err)
	}
	root := repository(t, map[string]string{"XT-101-host.md": edit(host, "state: Ready", "state: Planned") +
		"\n## Open questions\n\nWhich manifest format?\n"})
	var out strings.Builder
	if err := run([]string{"next", root}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), readyRule) || !strings.Contains(string(data), readyRule) {
		t.Fatalf("next and dev-preferences must share the readiness rule: %q", readyRule)
	}
}
