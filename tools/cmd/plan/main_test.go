package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestRepositoryPlan checks the real backlog files, so `just test` fails as
// soon as one drifts from the format.
func TestRepositoryPlan(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"check", "../../.."}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}

const foundation = `---
id: COR-100
core: foundation
state: Shipped. It builds
---

# COR-100 Foundation

## Demo

It builds.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-100-T1 | Skeleton | agent | — | done |

### COR-100-T1 Skeleton

- **Read:** [Foundation](#cor-100-foundation)
- **Proves:** ` + "`proof.txt`" + `
`

const host = `---
id: XT-101
extension: Host
state: Ready. It dispatches
---

# XT-101 Host

## Demo

It dispatches.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| XT-101-T1 | Pick the look | Pascal | — | todo |
| XT-101-T2 | Manifest | agent | — | todo |
| XT-101-T3 | Help | agent | XT-101-T1, XT-101-T2 | todo |

### XT-101-T1 Pick the look

Publish the mockups.

### XT-101-T2 Manifest

- **Read:** [Demo](#demo)
- **Proves:** ` + "`testdata/manifest.txtar`" + `

Read the manifest.

` + "```" + `
# a comment in a block is no heading
` + "```" + `

### XT-101-T3 Help

- **Read:** [Demo](#demo)
- **Proves:** ` + "`testdata/help.txtar`" + `
`

func TestNext(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
		lacks []string
	}{
		{
			name:  "first ready agent card, with the cards waiting on Pascal",
			files: map[string]string{"XT-101-host.md": host},
			want:  []string{"Waiting on Pascal: XT-101-T1 Pick the look", "Next card: XT-101-T2 Manifest", "Read the manifest.", "# a comment in a block is no heading"},
		},
		{
			name:  "agent cards blocked by Pascal",
			files: map[string]string{"XT-101-host.md": edit(host, "| XT-101-T2 | Manifest | agent | — | todo |", "| XT-101-T2 | Manifest | agent | XT-101-T1 | todo |")},
			want:  []string{"No agent card of XT-101 can start", "Publish the mockups."},
		},
		{
			name: "a Planned item asks for readiness",
			files: map[string]string{"XT-101-host.md": edit(host, "state: Ready", "state: Planned") +
				"\n## Open questions\n\n- Which manifest format?\n"},
			want: []string{"Next: make XT-101 Host ready", "- Which manifest format?", "Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests."},
		},
		{
			name:  "a Planned item without open questions starts with its first card",
			files: map[string]string{"XT-101-host.md": edit(host, "state: Ready", "state: Planned")},
			want:  []string{"Next card: XT-101-T2 Manifest", "The pull request sets this card to done."},
			lacks: []string{"small stack"},
		},
		{
			name:  "empty queue",
			files: map[string]string{},
			want:  []string{"The queue is empty"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := repository(t, tc.files)
			var out strings.Builder
			if err := run([]string{"next", root}, &out); err != nil {
				t.Fatalf("%v\n%s", err, out.String())
			}
			for _, w := range tc.want {
				if !strings.Contains(out.String(), w) {
					t.Errorf("next lacks %q:\n%s", w, out.String())
				}
			}
			for _, l := range tc.lacks {
				if strings.Contains(out.String(), l) {
					t.Errorf("next has %q:\n%s", l, out.String())
				}
			}
		})
	}
}

func TestCheck(t *testing.T) {
	cases := []struct {
		name, file, text, want string
	}{
		{"checkbox", "XT-101-host.md", host + "\n- [ ] Done when it works\n", "a checkbox belongs in no item file"},
		{"missing section", "XT-101-host.md", strings.Split(host, "### XT-101-T3")[0], "XT-101-T3 needs a section"},
		{"section without row", "XT-101-host.md", host + "\n### XT-101-T4 Extra\n", "XT-101-T4 has a section but no row"},
		{"title mismatch", "XT-101-host.md", edit(host, "### XT-101-T2 Manifest", "### XT-101-T2 Manifests"), "is titled \"Manifests\", and its row \"Manifest\""},
		{"unknown owner", "XT-101-host.md", edit(host, "| Pascal |", "| Pat |"), "the owner of XT-101-T1 must be one of"},
		{"unknown dependency", "XT-101-host.md", edit(host, "XT-101-T1, XT-101-T2 |", "XT-101-T9 |"), "depends on \"XT-101-T9\""},
		{"cycle", "XT-101-host.md", edit(host, "| XT-101-T2 | Manifest | agent | — |", "| XT-101-T2 | Manifest | agent | XT-101-T3 |"), "form a cycle"},
		{"agent card without proof", "XT-101-host.md", edit(host, "- **Proves:** `testdata/help.txtar`", ""), "needs a \"- **Proves:**\" line"},
		{"done card without its proof", "XT-101-host.md", edit(edit(host, "| XT-101-T2 | Manifest | agent | — | todo |", "| XT-101-T2 | Manifest | agent | — | done |"), "state: Ready", "state: Active"), "its proof testdata/manifest.txtar does not exist"},
		{"ready with a done card", "XT-101-host.md", edit(host, "| XT-101-T1 | Pick the look | Pascal | — | todo |", "| XT-101-T1 | Pick the look | Pascal | — | done |"), "a Ready item needs todo cards, no done card, and no open questions"},
		{"ready with open questions", "XT-101-host.md", host + "\n## Open questions\n\n- Which?\n", "a Ready item needs todo cards, no done card, and no open questions"},
		{"broken anchor", "XT-101-host.md", edit(host, "[Demo](#demo)", "[Demo](#nowhere)"), "the link #nowhere points to a missing heading"},
		{"broken file", "XT-101-host.md", edit(host, "[Demo](#demo)", "[Spec](../spec.md)"), "the link ../spec.md points to a missing file"},
		{"empty reading", "XT-101-host.md", edit(host, "- **Read:** [Demo](#demo)\n- **Proves:** `testdata/manifest.txtar`", "- **Read:**\n- **Proves:** `testdata/manifest.txtar`"), "XT-101-T2 needs a \"- **Read:**\" line that links"},
		{"reading without a link", "XT-101-host.md", edit(host, "- **Read:** [Demo](#demo)\n- **Proves:** `testdata/manifest.txtar`", "- **Read:** the spec\n- **Proves:** `testdata/manifest.txtar`"), "XT-101-T2 needs a \"- **Read:**\" line that links"},
		{"empty proof", "XT-101-host.md", edit(host, "- **Proves:** `testdata/help.txtar`", "- **Proves:**"), "XT-101-T3 needs a \"- **Proves:**\" line"},
		{"done card with a missing proof outside a folder", "COR-100-foundation.md", edit(foundation, "`proof.txt`", "`missing.txt`"), "its proof missing.txt does not exist"},
		{"unknown card in an item", "XT-101-host.md", edit(host, "It dispatches.", "It dispatches, as XT-101-T9 says."), "docs/backlog/XT-101-host.md:11: no item defines the card XT-101-T9"},
		{"unknown card in a code block of an item", "XT-101-host.md", edit(host, "# a comment in a block is no heading", "# COR-100-T2 is gone"), "docs/backlog/XT-101-host.md:33: no item defines the card COR-100-T2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"XT-101-host.md": host}
			files[tc.file] = tc.text
			root := repository(t, files)
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

func TestCheckDecisions(t *testing.T) {
	index := "# Decisions\n\n| Decision | Title |\n|---|---|\n| [D001](D001-first.md) | First |\n"
	cases := []struct {
		name, file, text, want string
	}{
		{"no row in the index", "D002-second.md", "# D002 Second\n", "D002-second.md: the decision needs a row"},
		{"title of another number", "D002-second.md", "# D009 Second\n", "D002-second.md:1: the first line must be \"# D002 <title>\""},
		{"number taken twice", "D001-again.md", "# D001 Again\n", "uses the same number"},
		{"bad file name", "D2-second.md", "# D2 Second\n", "D2-second.md: the file name must look like D001-short-title.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := repository(t, nil)
			dir := filepath.Join(root, "docs", "north-star", "decisions")
			write(t, filepath.Join(dir, "README.md"), index)
			write(t, filepath.Join(dir, "D001-first.md"), "# D001 First\n")
			write(t, filepath.Join(dir, tc.file), tc.text)
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

// repository writes shipped history and the supplied backlog fixtures.
func repository(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, backlogDir)
	all := map[string]string{"COR-100-foundation.md": foundation}
	for name, text := range files {
		all[name] = text
	}
	var names []string
	for name := range all {
		names = append(names, name)
	}
	slices.Sort(names)
	queue := "# Backlog\n\n## Queue\n\n"
	n := 0
	for _, name := range names {
		text := all[name]
		write(t, filepath.Join(dir, name), text)
		if strings.Contains(text, "state: Ready") || strings.Contains(text, "state: Planned") || strings.Contains(text, "state: Active") {
			n++
			id := strings.Split(strings.TrimPrefix(strings.Split(text, "\n")[1], "id: "), " ")[0]
			queue += fmt.Sprintf("%d. %s\n", n, id)
		}
	}
	write(t, filepath.Join(dir, "README.md"), queue)
	write(t, filepath.Join(root, "proof.txt"), "proof\n")
	return root
}

func edit(text, old, replacement string) string {
	if !strings.Contains(text, old) {
		panic("fixture lacks " + old)
	}
	return strings.Replace(text, old, replacement, 1)
}

func write(t *testing.T, name, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sources writes the files that define the codes a doc may name. They are the
// parts of contract.md, an extension's spec and a part of it, and the manifest
// of the hello fixture.
func sources(t *testing.T, root string) {
	t.Helper()
	write(t, filepath.Join(root, "docs", "north-star", "contract", "exit-codes.md"), "# Exit codes\n\n| Code | Meaning |\n|---|---|\n| `78` | A human must act |\n")
	write(t, filepath.Join(root, "docs", "north-star", "contract", "error-codes.md"), "# Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `USAGE_INVALID` | 2 | A bad flag |\n\n| Code | When |\n|---|---|\n| `CONFIG_UNKNOWN_KEY` | An unknown key |\n")
	write(t, filepath.Join(root, "extensions", "demo", "spec.md"), "# Demo\n\n## Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `DEMO_BROKEN` | 1 | It broke |\n")
	write(t, filepath.Join(root, "extensions", "demo", "spec", "codes.md"), "# Codes\n\n| Warning | When |\n|---|---|\n| `DEMO_SLOW` | It was slow |\n")
	write(t, filepath.Join(root, "testdata", "extensions", "hello", "manifest.toml"), "domain = \"hello\"\n\n[[commands]]\npath = \"hello run\"\ncode = \"HELLO_NOT_AN_ERROR\"\n\n[[errors]]\ncode = \"HELLO_GRUMPY\"\nexit = 1\n")
}

func TestCheckCodesAndTypography(t *testing.T) {
	cases := []struct {
		name, file, text, want string
	}{
		{"unknown code after code", "docs/backlog/XT-101-host.md", edit(host, "It dispatches.", "It dispatches, or fails with the code `HELLO_UNKNOWN`."), "docs/backlog/XT-101-host.md:11: no code table or fixture manifest defines the code HELLO_UNKNOWN"},
		{"unknown code after error code", "README.md", "# Read me\n\nIt answers the error code `DEMO_GONE`.\n", "README.md:3: no code table or fixture manifest defines the code DEMO_GONE"},
		{"unknown code after warning code", "extensions/demo/SKILL.md", "Relay the warning code `DEMO_QUIET`.\n", "extensions/demo/SKILL.md:1: no code table or fixture manifest defines the code DEMO_QUIET"},
		{"unknown code in a Code column", "extensions/demo/notes.md", "| Id | Code | When |\n|---|---|---|\n| `demo.check` | `DEMO_LOST` | It warns |\n", "extensions/demo/notes.md:3: no code table or fixture manifest defines the code DEMO_LOST"},
		{"unknown code in an errors example", "docs/north-star/vision.md", "An answer:\n\n```json\n{\"ok\":false,\"errors\":[{\"code\":\"RUNTIME_LOST\",\"message\":\"m\",\"hint\":\"h\"}]}\n```\n", "docs/north-star/vision.md:4: no code table or fixture manifest defines the code RUNTIME_LOST"},
		{"unknown code after a bracket in a string", "docs/north-star/vision.md", "```json\n{\"ok\":false,\"errors\":[{\"message\":\"Unexpected ]\",\"code\":\"DEMO_GONE\"}]}\n```\n", "docs/north-star/vision.md:2: no code table or fixture manifest defines the code DEMO_GONE"},
		{"prose after a nested fence", "README.md", "````markdown\n```\nIt\u2019s code.\n```\n````\n\nIt\u2019s prose.\n", "README.md:7: the prose holds \u2019 (U+2019)"},
		{"unknown code in a block that the file never closes", "README.md", "# Read me\n\n```json\n{\"errors\":[{\"code\":\"DEMO_OPEN\"}]}\n", "README.md:4: no code table or fixture manifest defines the code DEMO_OPEN"},
		{"code of a table in an example", "extensions/demo/spec.md", "# Demo\n\n## Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `DEMO_BROKEN` | 1 | It broke |\n\nAn extension writes its table so:\n\n```markdown\n## Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `DEMO_PHANTOM` | 1 | Example |\n```\n\nIt never answers the code `DEMO_PHANTOM`.\n", "extensions/demo/spec.md:19: no code table or fixture manifest defines the code DEMO_PHANTOM"},
		{"unknown code in a warnings example over lines", "docs/north-star/vision.md", "```json\n{\n  \"ok\": true,\n  \"warnings\": [\n    {\"code\": \"DEMO_SLOW\"},\n    {\"code\": \"DEMO_FAST\"}\n  ]\n}\n```\n", "docs/north-star/vision.md:6: no code table or fixture manifest defines the code DEMO_FAST"},
		{"declared code without its manifest", "docs/backlog/XT-101-host.md", edit(host, "It dispatches.", "It dispatches, with the own code `HELLO_GRUMPY`."), "docs/backlog/XT-101-host.md:11: no code table or fixture manifest defines the code HELLO_GRUMPY"},
		{"em dash in prose", "docs/backlog/XT-101-host.md", edit(host, "It dispatches.", "It dispatches \u2014 fast."), "docs/backlog/XT-101-host.md:11: the prose holds \u2014 (U+2014)"},
		{"curly apostrophe in prose", "AGENTS.md", "# AGENTS.md\n\n## Code\n\nIt\u2019s fine.\n", "AGENTS.md:5: the prose holds \u2019 (U+2019)"},
		{"curly quotes in a table cell", "README.md", "| A | B |\n|---|---|\n| x | \u201cy\u201d |\n", "README.md:3: the prose holds \u201c (U+201C)"},
		{"left single quote in prose", "README.md", "A \u2018word\u2019.\n", "README.md:1: the prose holds \u2018 (U+2018)"},
		{"right double quote in prose", "README.md", "A word\u201d.\n", "README.md:1: the prose holds \u201d (U+201D)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := repository(t, map[string]string{"XT-101-host.md": host})
			sources(t, root)
			if tc.name == "declared code without its manifest" {
				if err := os.Remove(filepath.Join(root, "testdata", "extensions", "hello", "manifest.toml")); err != nil {
					t.Fatal(err)
				}
			}
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

func TestCheckCodesAndTypographyPass(t *testing.T) {
	// Every code below has its source, or stands outside the three contexts
	// whose role is explicit: a cell under Code, a span after "code", and the
	// "code" of an errors or warnings example.
	known := "It answers the code `USAGE_INVALID`, the error code `DEMO_BROKEN`, the warning code `DEMO_SLOW`, the warning code `CONFIG_UNKNOWN_KEY`, and, as an example, the code `FLEET_HOST_DOWN`.\n"
	outside := "Set `NCLY_JSON` and `NCLY_LANG`, exit code `78` or `2`, and expect `SOME_THING`, OTHER_THING, the codes `NONE_SUCH`, and code: `NOPE_NOPE`.\n\n" +
		"| Code | Meaning |\n|---|---|\n| `0` | Success |\n| `130` | Ctrl-C |\n\n" +
		"```json\n{\"ok\":false,\"errors\":[{\"code\":\"USAGE_INVALID\",\"message\":\"m\"}],\"warnings\":[{\"code\":\"DEMO_SLOW\"}]}\n{\"checks\":[{\"id\":\"x\",\"code\":\"NOT_IN_AN_ENVELOPE\"}]}\n```\n\n" +
		"```\nstderr '\"code\":\"NOT_IN_AN_ENVELOPE\"'\n```\n"
	typography := "Write `a \u2014 b` and `\u2018x\u2019`.\n\n```\nIt\u2019s \u201cquoted\u201d \u2014 here.\n```\n\n- A list\n\n  ~~~\n  \u2014 in an indented block\n  ~~~\n"
	nested := "````markdown\n```\nIt\u2019s \u2014 code.\n```\n````\n\n" +
		"```json\n{\"errors\":[{\"message\":\"[\",\"code\":\"USAGE_INVALID\"}],\"checks\":[{\"code\":\"OUT_OF_SCOPE\"}]}\n```\n"
	root := repository(t, map[string]string{"XT-101-host.md": edit(host, "It dispatches.", "It dispatches, with the own code `HELLO_GRUMPY`.")})
	sources(t, root)
	// The manifest is TOML: a comment and a literal string change nothing.
	write(t, filepath.Join(root, "testdata", "extensions", "hello", "manifest.toml"), "domain = \"hello\"\n\n[[errors]] # the own codes\ncode = 'HELLO_GRUMPY'\nexit = 1\n")
	write(t, filepath.Join(root, "AGENTS.md"), "# AGENTS.md\n\n## Code\n")
	write(t, filepath.Join(root, "docs", "north-star", "vision.md"), known+"\n"+outside+"\n"+typography+"\n"+nested)
	write(t, filepath.Join(root, "docs", "archived", "chat.md"), "The code `GONE_CODE` \u2014 it\u2019s old.\n")
	write(t, filepath.Join(root, "extensions", "demo", "testdata", "notes.md"), "The code `FIXTURE_ONLY` \u2014 here.\n")
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}
