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
			want: []string{"Next: make M01 Host ready", "- Which manifest format?"},
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
		})
	}
}

func TestStatus(t *testing.T) {
	root := repository(t, map[string]string{"M01-host.md": host})
	var out strings.Builder
	if err := run([]string{"status", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, w := range []string{"M00  done   1/1  Foundation\n", "M01  ready  0/3  Host  <- current\n", "todo  agent   M01-T3  Help  after M01-T1, M01-T2\n"} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("status lacks %q:\n%s", w, out.String())
		}
	}
	if strings.Contains(out.String(), "M99") {
		t.Errorf("status lists the parking lot:\n%s", out.String())
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
		{"bad version", "M01-host.md", edit(host, "Version: v0.1.0", "Version: 0.1"), "Version must name the release"},
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
