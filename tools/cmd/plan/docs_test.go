package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDocs(t *testing.T) {
	cases := []struct {
		name, file, text, want string
	}{
		{"prose in milestone README", "docs/milestones/README.md", "It\u2019s current.\n", "the prose holds \u2019"},
		{"missing file", "docs/north-star/vision.md", "See [the spec](spec.md).\n", "docs/north-star/vision.md:1: the link spec.md points to a missing file"},
		{"missing file in the plan's readme", "docs/backlog/README.md", "See [the spec](spec.md).\n", "docs/backlog/README.md:1: the link spec.md points to a missing file"},
		{"missing heading", "AGENTS.md", "# AGENTS.md\n\n## Code\n\nSee [Rules](#rules).\n", "AGENTS.md:5: the link #rules points to a missing heading"},
		{"missing heading in another file", "extensions/skill/spec.md", "[Code](../../AGENTS.md#rules)\n", "extensions/skill/spec.md:1: the link ../../AGENTS.md#rules points to a missing heading"},
		{"link text in a code span", "README.md", "See [`ncly`](gone.md).\n", "README.md:1: the link gone.md points to a missing file"},
		{"unknown card", "docs/north-star/vision.md", "COR-100-T1 is done.\nCOR-100-T2 is not.\n", "docs/north-star/vision.md:2: no item defines the card COR-100-T2"},
		{"unknown card in a code span", "extensions/skill/spec.md", "Run `XT-107-T2`.\n", "extensions/skill/spec.md:1: no item defines the card XT-107-T2"},
		{"unknown card in a code block", "README.md", "# Title\n\n```\nXT-101-T1\n```\n", "README.md:4: no item defines the card XT-101-T1"},
		{"missing file in a skill of the agents", ".agents/skills/release-nicely/SKILL.md", "See [D043](gone.md).\n", ".agents/skills/release-nicely/SKILL.md:1: the link gone.md points to a missing file"},
		{"missing file in the rules of the scenarios", "testdata/script/README.md", "See [the spec](spec.md).\n", "testdata/script/README.md:1: the link spec.md points to a missing file"},
		{"missing file in the rules of an extension's scenarios", "extensions/agent/testdata/script/README.md", "See [the spec](spec.md).\n", "extensions/agent/testdata/script/README.md:1: the link spec.md points to a missing file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := repository(t, nil)
			write(t, filepath.Join(root, "AGENTS.md"), "# AGENTS.md\n\n## Code\n")
			write(t, filepath.Join(root, filepath.FromSlash(tc.file)), tc.text)
			var out strings.Builder
			if err := run([]string{"check", root}, &out); err == nil {
				t.Fatalf("check passed:\n%s", tc.text)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Errorf("check lacks %q:\n%s", tc.want, out.String())
			}
		})
	}
}

func TestCheckDocsReportsAnUnreadableFile(t *testing.T) {
	root := repository(t, nil)
	if err := os.Symlink("gone.md", filepath.Join(root, "dangling.md")); err != nil {
		t.Skip("no symlinks:", err)
	}
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err == nil {
		t.Fatal("check passed a dangling link")
	}
	if !strings.Contains(out.String(), "dangling.md: open") {
		t.Errorf("check lacks the unreadable file:\n%s", out.String())
	}
}

func TestCheckDocsPasses(t *testing.T) {
	// A card of another item is known, in prose and in code, and the
	// archives keep the card IDs of their time.
	root := repository(t, map[string]string{"XT-101-host.md": edit(host, "It dispatches.", "It dispatches, after COR-100-T1.")})
	write(t, filepath.Join(root, "AGENTS.md"), "# AGENTS.md\n\n## Code\n")
	write(t, filepath.Join(root, "docs", "north-star", "vision.md"), "[Code](../../AGENTS.md#code), [site](https://example.com), [folder](../backlog/)\n\nCOR-100-T1 and `XT-101-T3` hold, as do XT-101-T2/XT-101-T1 and XT-101-T10x.\n\n```\n[example](gone.md)\n```\n")
	write(t, filepath.Join(root, "docs", "archived", "chat.md"), "[old](gone.md), from XT-104-T9 before the split\n")
	write(t, filepath.Join(root, "extensions", "markdown", "testdata", "broken.md"), "[old](gone.md)\n")
	write(t, filepath.Join(root, "docs", "north-star", "code.md"), "Write `[a](gone.md)` for a link.\n\n- A list\n\n  ~~~\n  [example](gone.md)\n  ~~~\n")
	write(t, filepath.Join(root, ".github", "notes.md"), "[old](gone.md)\n")
	// A fixture beside the rules of the scenarios keeps its broken link too.
	// The check never walks a testdata folder, so a fixture folder that it cannot
	// read fails nothing.
	write(t, filepath.Join(root, "testdata", "script", "broken.md"), "[old](gone.md)\n")
	opaque := filepath.Join(root, "testdata", "opaque")
	if err := os.Mkdir(opaque, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(opaque, 0o700) })
	// The plan's README remains a general document.
	write(t, filepath.Join(root, "docs", "backlog", "README.md"), "# Backlog\n\n## Queue\n\n1. XT-101\n\n## References\n\n[Cards](XT-101-host.md#cards)\n")
	// The link and its heading read the same fences: a short fence inside a
	// longer block closes nothing.
	write(t, filepath.Join(root, "docs", "north-star", "nested.md"), "# Demo\n\n````markdown\n```\n````\n\n## Target\n\n[Target](#target)\n")
	// A marker indented by four spaces is code, so it opens no fence that
	// would hide the headings below it.
	write(t, filepath.Join(root, "docs", "north-star", "indented.md"), "# Indented\n\n    ~~~\n\n## Tilde\n\nText.\n\n    ```\n\n## Backtick\n")
	write(t, filepath.Join(root, "docs", "north-star", "links.md"), "[Tilde](indented.md#tilde), [Backtick](indented.md#backtick)\n")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}
