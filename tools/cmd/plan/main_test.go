package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRepositoryPlan checks the real milestone files, so `just test` fails as
// soon as one drifts from the format.
func TestRepositoryPlan(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"check", "../../.."}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}

// TestReadyRule keeps the docs that state how a milestone becomes ready in
// the words of next's message, so the rule cannot drift apart.
func TestReadyRule(t *testing.T) {
	for _, name := range []string{"docs/north-star/dev-preferences.md", "docs/north-star/decisions/D042-milestones-and-cards.md"} {
		data, err := os.ReadFile(filepath.Join("../../..", name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), readyRule) {
			t.Errorf("%s lacks the rule of next, in its words: %q", name, readyRule)
		}
	}
}

const foundation = `# M00 Foundation

Status: done
Version: v0.0.1

## Demo

It builds.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M00-T1 | Skeleton | agent | — | done |

### M00-T1 Skeleton

- **Read:** [Foundation](#m00-foundation)
- **Proves:** ` + "`proof.txt`" + `
`

const host = `# M01 Host

Status: ready
Version: v0.1.0

## Demo

It dispatches.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M01-T1 | Pick the look | Pascal | — | todo |
| M01-T2 | Manifest | agent | — | todo |
| M01-T3 | Help | agent | M01-T1, M01-T2 | todo |

### M01-T1 Pick the look

Publish the mockups.

### M01-T2 Manifest

- **Read:** [Demo](#demo)
- **Proves:** ` + "`testdata/manifest.txtar`" + `

Read the manifest.

` + "```" + `
# a comment in a block is no heading
` + "```" + `

### M01-T3 Help

- **Read:** [Demo](#demo)
- **Proves:** ` + "`testdata/help.txtar`" + `
`

const parkingLot = `# M99 Parking lot

Status: open

Ideas.
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
			files: map[string]string{"M01-host.md": host},
			want:  []string{"Waiting on Pascal: M01-T1 Pick the look", "Next card: M01-T2 Manifest", "which ships in v0.1.0", "Read the manifest.", "# a comment in a block is no heading"},
		},
		{
			name:  "agent cards blocked by Pascal",
			files: map[string]string{"M01-host.md": edit(host, "| M01-T2 | Manifest | agent | — | todo |", "| M01-T2 | Manifest | agent | M01-T1 | todo |")},
			want:  []string{"No agent card of M01 can start", "Publish the mockups."},
		},
		{
			name: "a planned milestone asks for readiness",
			files: map[string]string{"M01-host.md": edit(host, "Status: ready", "Status: planned") +
				"\n## Open questions\n\n- Which manifest format?\n"},
			want: []string{"Next: make M01 Host ready", "- Which manifest format?", "Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests."},
		},
		{
			name:  "a planned milestone without open questions starts with its first card",
			files: map[string]string{"M01-host.md": edit(host, "Status: ready", "Status: planned")},
			want:  []string{"Next card: M01-T2 Manifest", "M01 has no open questions, so this pull request also sets it to active."},
			lacks: []string{"small stack"},
		},
		{
			name:  "every milestone done",
			files: map[string]string{},
			want:  []string{"Every milestone is done"},
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

func TestStatus(t *testing.T) {
	root := repository(t, map[string]string{"M01-host.md": host})
	var out strings.Builder
	if err := run([]string{"status", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	// M01-T3 waits on Pascal's M01-T1, so the order stops after M01-T2.
	for _, w := range []string{
		"M00  done   1/1  Foundation\n", "M01  ready  0/3  Host  <- current\n", "todo  agent   M01-T3  Help  after M01-T1, M01-T2\n",
		"If each agent card succeeds, just next serves them in this order: M01-T2\n", "These agent cards wait on a card of Pascal: M01-T3\n",
	} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("status lacks %q:\n%s", w, out.String())
		}
	}
	if strings.Contains(out.String(), "M99") {
		t.Errorf("status lists the parking lot:\n%s", out.String())
	}
}

// describe has the dependencies of M03, whose cards just next served in the
// order T1, T2, T4, T5, T3 (#64).
var describe = `# M01 Describe

Status: ready
Version: v0.1.0

## Demo

It describes.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M01-T1 | Listing | agent | — | todo |
| M01-T2 | Detail | agent | M01-T1 | todo |
| M01-T3 | Schemas | agent | M01-T2, M01-T5 | todo |
| M01-T4 | Config | agent | M01-T2 | todo |
| M01-T5 | Help | agent | M01-T2 | todo |
` + describeSections

var describeSections = func() string {
	var b strings.Builder
	for _, c := range []string{"M01-T1 Listing", "M01-T2 Detail", "M01-T3 Schemas", "M01-T4 Config", "M01-T5 Help"} {
		b.WriteString("\n### " + c + "\n\n- **Read:** [Demo](#demo)\n- **Proves:** `proof.txt`\n")
	}
	return b.String()
}()

func TestStatusOrder(t *testing.T) {
	root := repository(t, map[string]string{"M01-describe.md": describe})
	file := filepath.Join(root, "docs", "milestones", "M01-describe.md")
	var next, out, again strings.Builder
	if err := run([]string{"next", root}, &next); err != nil {
		t.Fatalf("%v\n%s", err, next.String())
	}
	if err := run([]string{"status", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if w := "If each agent card succeeds, just next serves them in this order: M01-T1, M01-T2, M01-T4, M01-T5, M01-T3\n"; !strings.Contains(out.String(), w) {
		t.Errorf("status lacks %q:\n%s", w, out.String())
	}
	if strings.Contains(out.String(), "wait on a card of Pascal") {
		t.Errorf("status finds a card of Pascal:\n%s", out.String())
	}
	// The order lives in memory: the plan and the choice of next stay as they were.
	if data, err := os.ReadFile(file); err != nil || string(data) != describe {
		t.Errorf("status changed the milestone file: %v\n%s", err, data)
	}
	if err := run([]string{"next", root}, &again); err != nil || again.String() != next.String() {
		t.Errorf("next changed after status: %v\n%s", err, again.String())
	}
	if !strings.Contains(next.String(), "Next card: M01-T1 Listing") {
		t.Errorf("next lacks M01-T1:\n%s", next.String())
	}
}

func TestCheck(t *testing.T) {
	cases := []struct {
		name, file, text, want string
	}{
		{"checkbox", "M01-host.md", host + "\n- [ ] Done when it works\n", "a checkbox belongs in no milestone file"},
		{"missing section", "M01-host.md", strings.Split(host, "### M01-T3")[0], "M01-T3 needs a section"},
		{"section without row", "M01-host.md", host + "\n### M01-T4 Extra\n", "M01-T4 has a section but no row"},
		{"title mismatch", "M01-host.md", edit(host, "### M01-T2 Manifest", "### M01-T2 Manifests"), "is titled \"Manifests\", and its row \"Manifest\""},
		{"unknown owner", "M01-host.md", edit(host, "| Pascal |", "| Pat |"), "the owner of M01-T1 must be one of"},
		{"unknown dependency", "M01-host.md", edit(host, "M01-T1, M01-T2 |", "M01-T9 |"), "depends on \"M01-T9\""},
		{"cycle", "M01-host.md", edit(host, "| M01-T2 | Manifest | agent | — |", "| M01-T2 | Manifest | agent | M01-T3 |"), "form a cycle"},
		{"agent card without proof", "M01-host.md", edit(host, "- **Proves:** `testdata/help.txtar`", ""), "needs a \"- **Proves:**\" line"},
		{"done card without its proof", "M01-host.md", edit(edit(host, "| M01-T2 | Manifest | agent | — | todo |", "| M01-T2 | Manifest | agent | — | done |"), "Status: ready", "Status: active"), "its proof testdata/manifest.txtar does not exist"},
		{"ready with a done card", "M01-host.md", edit(host, "| M01-T1 | Pick the look | Pascal | — | todo |", "| M01-T1 | Pick the look | Pascal | — | done |"), "a milestone with a done card is active"},
		{"ready with open questions", "M01-host.md", host + "\n## Open questions\n\n- Which?\n", "a ready milestone has no open questions"},
		{"too many agent cards", "M01-host.md", manyCards(6), "6 agent cards is more than 5"},
		{"out of order", "M02-next.md", edit(edit(foundation, "# M00 Foundation", "# M02 Next"), "M00-T1", "M02-T1"), "milestones go in order, so M02 cannot start before M01 is done"},
		{"broken anchor", "M01-host.md", edit(host, "[Demo](#demo)", "[Demo](#nowhere)"), "the link #nowhere points to a missing heading"},
		{"broken file", "M01-host.md", edit(host, "[Demo](#demo)", "[Spec](../spec.md)"), "the link ../spec.md points to a missing file"},
		{"bad version", "M01-host.md", edit(host, "Version: v0.1.0", "Version: v0.2.0"), "Version must be v0.1.0, the release that ships M01"},
		{"empty reading", "M01-host.md", edit(host, "- **Read:** [Demo](#demo)\n- **Proves:** `testdata/manifest.txtar`", "- **Read:**\n- **Proves:** `testdata/manifest.txtar`"), "M01-T2 needs a \"- **Read:**\" line that links"},
		{"reading without a link", "M01-host.md", edit(host, "- **Read:** [Demo](#demo)\n- **Proves:** `testdata/manifest.txtar`", "- **Read:** the spec\n- **Proves:** `testdata/manifest.txtar`"), "M01-T2 needs a \"- **Read:**\" line that links"},
		{"empty proof", "M01-host.md", edit(host, "- **Proves:** `testdata/help.txtar`", "- **Proves:**"), "M01-T3 needs a \"- **Proves:**\" line"},
		{"done card with a missing proof outside a folder", "M00-foundation.md", edit(foundation, "`proof.txt`", "`missing.txt`"), "its proof missing.txt does not exist"},
		{"unknown card in a milestone", "M01-host.md", edit(host, "It dispatches.", "It dispatches, as M01-T9 says."), "docs/milestones/M01-host.md:8: no milestone defines the card M01-T9"},
		{"unknown card in a code block of a milestone", "M01-host.md", edit(host, "# a comment in a block is no heading", "# M00-T2 is gone"), "docs/milestones/M01-host.md:30: no milestone defines the card M00-T2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"M01-host.md": host}
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

// repository writes a repository with the foundation, the parking lot, and
// files, whose names are relative to docs/milestones.
func repository(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "milestones")
	all := map[string]string{"M00-foundation.md": foundation, "M99-parking-lot.md": parkingLot}
	for name, text := range files {
		all[name] = text
	}
	for name, text := range all {
		write(t, filepath.Join(dir, name), text)
	}
	write(t, filepath.Join(root, "proof.txt"), "proof\n")
	return root
}

func manyCards(n int) string {
	text := "# M01 Host\n\nStatus: ready\nVersion: v0.1.0\n\n## Demo\n\nIt grows.\n\n## Cards\n\n| Card | Title | Owner | Depends on | Status |\n|---|---|---|---|---|\n"
	var sections string
	for i := 1; i <= n; i++ {
		id := "M01-T" + string(rune('0'+i))
		text += "| " + id + " | Card | agent | — | todo |\n"
		sections += "\n### " + id + " Card\n\n- **Read:** [Demo](#demo)\n- **Proves:** `x.txtar`\n"
	}
	return text + sections
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
		{"unknown code after code", "docs/milestones/M01-host.md", edit(host, "It dispatches.", "It dispatches, or fails with the code `HELLO_UNKNOWN`."), "docs/milestones/M01-host.md:8: no code table or fixture manifest defines the code HELLO_UNKNOWN"},
		{"unknown code after error code", "README.md", "# Read me\n\nIt answers the error code `DEMO_GONE`.\n", "README.md:3: no code table or fixture manifest defines the code DEMO_GONE"},
		{"unknown code after warning code", "extensions/demo/SKILL.md", "Relay the warning code `DEMO_QUIET`.\n", "extensions/demo/SKILL.md:1: no code table or fixture manifest defines the code DEMO_QUIET"},
		{"unknown code in a Code column", "extensions/demo/notes.md", "| Id | Code | When |\n|---|---|---|\n| `demo.check` | `DEMO_LOST` | It warns |\n", "extensions/demo/notes.md:3: no code table or fixture manifest defines the code DEMO_LOST"},
		{"unknown code in an errors example", "docs/north-star/vision.md", "An answer:\n\n```json\n{\"ok\":false,\"errors\":[{\"code\":\"RUNTIME_LOST\",\"message\":\"m\",\"hint\":\"h\"}]}\n```\n", "docs/north-star/vision.md:4: no code table or fixture manifest defines the code RUNTIME_LOST"},
		{"unknown code after a bracket in a string", "docs/north-star/vision.md", "```json\n{\"ok\":false,\"errors\":[{\"message\":\"Unexpected ]\",\"code\":\"DEMO_GONE\"}]}\n```\n", "docs/north-star/vision.md:2: no code table or fixture manifest defines the code DEMO_GONE"},
		{"prose after a nested fence", "README.md", "````markdown\n```\nIt\u2019s code.\n```\n````\n\nIt\u2019s prose.\n", "README.md:7: the prose holds \u2019 (U+2019)"},
		{"unknown code in a block that the file never closes", "README.md", "# Read me\n\n```json\n{\"errors\":[{\"code\":\"DEMO_OPEN\"}]}\n", "README.md:4: no code table or fixture manifest defines the code DEMO_OPEN"},
		{"code of a table in an example", "extensions/demo/spec.md", "# Demo\n\n## Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `DEMO_BROKEN` | 1 | It broke |\n\nAn extension writes its table so:\n\n```markdown\n## Error codes\n\n| Code | Exit | When |\n|---|---|---|\n| `DEMO_PHANTOM` | 1 | Example |\n```\n\nIt never answers the code `DEMO_PHANTOM`.\n", "extensions/demo/spec.md:19: no code table or fixture manifest defines the code DEMO_PHANTOM"},
		{"unknown code in a warnings example over lines", "docs/north-star/vision.md", "```json\n{\n  \"ok\": true,\n  \"warnings\": [\n    {\"code\": \"DEMO_SLOW\"},\n    {\"code\": \"DEMO_FAST\"}\n  ]\n}\n```\n", "docs/north-star/vision.md:6: no code table or fixture manifest defines the code DEMO_FAST"},
		{"declared code without its manifest", "docs/milestones/M01-host.md", edit(host, "It dispatches.", "It dispatches, with the own code `HELLO_GRUMPY`."), "docs/milestones/M01-host.md:8: no code table or fixture manifest defines the code HELLO_GRUMPY"},
		{"em dash in prose", "docs/milestones/M01-host.md", edit(host, "It dispatches.", "It dispatches \u2014 fast."), "docs/milestones/M01-host.md:8: the prose holds \u2014 (U+2014)"},
		{"curly apostrophe in prose", "AGENTS.md", "# AGENTS.md\n\n## Code\n\nIt\u2019s fine.\n", "AGENTS.md:5: the prose holds \u2019 (U+2019)"},
		{"curly quotes in a table cell", "README.md", "| A | B |\n|---|---|\n| x | \u201cy\u201d |\n", "README.md:3: the prose holds \u201c (U+201C)"},
		{"left single quote in prose", "README.md", "A \u2018word\u2019.\n", "README.md:1: the prose holds \u2018 (U+2018)"},
		{"right double quote in prose", "README.md", "A word\u201d.\n", "README.md:1: the prose holds \u201d (U+201D)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := repository(t, map[string]string{"M01-host.md": host})
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
	root := repository(t, map[string]string{"M01-host.md": edit(host, "It dispatches.", "It dispatches, with the own code `HELLO_GRUMPY`.")})
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
