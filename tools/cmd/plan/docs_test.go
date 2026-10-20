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
		{"missing file", "docs/north-star/vision.md", "See [the spec](spec.md).\n", "docs/north-star/vision.md:1: the link spec.md points to a missing file"},
		{"missing file in the plan's readme", "docs/milestones/README.md", "See [the spec](spec.md).\n", "docs/milestones/README.md:1: the link spec.md points to a missing file"},
		{"missing heading", "AGENTS.md", "# AGENTS.md\n\n## Code\n\nSee [Rules](#rules).\n", "AGENTS.md:5: the link #rules points to a missing heading"},
		{"missing heading in another file", "extensions/skill/spec.md", "[Code](../../AGENTS.md#rules)\n", "extensions/skill/spec.md:1: the link ../../AGENTS.md#rules points to a missing heading"},
		{"link text in a code span", "README.md", "See [`ncly`](gone.md).\n", "README.md:1: the link gone.md points to a missing file"},
		{"unknown card", "docs/north-star/vision.md", "M00-T1 is done.\nM00-T2 is not.\n", "docs/north-star/vision.md:2: no milestone defines the card M00-T2"},
		{"unknown card in a code span", "extensions/skill/spec.md", "Run `M07-T2`.\n", "extensions/skill/spec.md:1: no milestone defines the card M07-T2"},
		{"unknown card in a code block", "README.md", "# Title\n\n```\nM01-T1\n```\n", "README.md:4: no milestone defines the card M01-T1"},
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
	// A card of another milestone is known, in prose and in code, and the
	// archives keep the card IDs of their time.
	root := repository(t, map[string]string{"M01-host.md": edit(host, "It dispatches.", "It dispatches, after M00-T1.")})
	write(t, filepath.Join(root, "AGENTS.md"), "# AGENTS.md\n\n## Code\n")
	write(t, filepath.Join(root, "docs", "north-star", "vision.md"), "[Code](../../AGENTS.md#code), [site](https://example.com), [folder](../milestones/)\n\nM00-T1 and `M01-T3` hold, as do M01-T2/M01-T1 and M01-T10x.\n\n```\n[example](gone.md)\n```\n")
	write(t, filepath.Join(root, "docs", "archived", "chat.md"), "[old](gone.md), from M04-T9 before the split\n")
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
	// The plan's readme is no milestone.
	write(t, filepath.Join(root, "docs", "milestones", "README.md"), "# Milestones\n\n[Cards](M01-host.md#cards)\n")
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
