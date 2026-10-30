package extensiontest_test

import (
	"cmp"
	"os"
	"strings"
	"testing"
)

// An edit changes the first old of a file of the fixture into new. An edit
// without old removes its file, and one with new alone writes it.
type edit struct{ file, old, new string }

// TestEachCheckFailsAndNamesItsFix breaks a copy of a fixture repository
// in one way per case, then runs its go test. That test must fail with the
// fix of the check that the break trips. The fixture's own go test, which
// just test runs, proves that the unbroken repository passes. The cases run
// one at a time, because each go test builds and links. In parallel, they
// slowed the pty scenarios of cmd/ncly past their limit on a 12-core Mac.
func TestEachCheckFailsAndNamesItsFix(t *testing.T) {
	breakSchema := edit{"manifest.toml", `count = { type = "integer" }`, `count = { type = "string" }`}
	cases := []struct {
		name string
		// fixture defaults to standalone.
		fixture string
		edits   []edit
		want    []string
		// absent holds what the output must not hold.
		absent []string
	}{
		{
			name:   "a manifest that names a missing catalog ID",
			edits:  []edit{{"manifest.toml", "standalone.tally.summary", "standalone.missing.summary"}},
			want:   []string{`locales/en.toml lacks the ID "standalone.missing.summary" that the manifest names. Fix the extension until extension.Check accepts it`},
			absent: []string{"map[", "<no value>", "names.."},
		},
		{
			name:  "a command without a scenario",
			edits: []edit{{file: "testdata/script/keys.txtar"}},
			want:  []string{"testdata/script: no scenario runs `ncly standalone fetch`. Fix: add a scenario that runs it"},
		},
		{
			name:  "a skill that names an undeclared command",
			edits: []edit{{"skill/SKILL.md", "Add `--json`", "Sum with `ncly standalone total`.\nAdd `--json`"}},
			want:  []string{"skill/SKILL.md: names `ncly standalone total`, which the manifest lacks. Fix: declare that command in manifest.toml"},
		},
		{
			name:  "an answer that breaks its schema",
			edits: []edit{breakSchema},
			want:  []string{"tally.txtar:4: `ncly standalone tally --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly standalone tally`"},
		},
		{
			name:  "an answer that breaks its schema, after a flag before the domain",
			edits: []edit{breakSchema, {"testdata/script/tally.txtar", "ncly standalone tally --json\nstdout '\"count\":3'\n", ""}},
			want:  []string{"tally.txtar:6: `ncly --json standalone tally`: the answer on stdout breaks the schema"},
		},
		{
			name:  "a Nicely version in .copier-answers.yml that differs from go.mod",
			edits: []edit{{".copier-answers.yml", "_commit: v0.18.0", "_commit: v0.17.0"}},
			want:  []string{`.copier-answers.yml: the template comes from Nicely "v0.17.0", and go.mod pins Nicely v0.18.0. Fix: run just update v0.18.0`},
		},
		{
			name: "a call that skips the answer check",
			edits: []edit{
				{"testdata/script/tally.txtar", "\nncly standalone tally --json\n", "\nexec ncly --json standalone tally\n"},
				{"testdata/script/lipgloss.txtar", "\nncly-standalone --version\n", "\nexec $XDG_DATA_HOME/nicely/extensions/ncly-standalone --version\n"},
				{"testdata/script/keys.txtar", "\ngrant example\n", "\ngrant example\nexec ncly standalone fetch --json # not in the background &\nexec ncly standalone fetch --json & # in the background\nwait\n"},
			},
			want: []string{
				"testdata/script/tally.txtar:4: exec runs ncly, so the kit cannot check its answer. Fix: drop exec, and run the kit's command ncly",
				"testdata/script/lipgloss.txtar:12: exec runs ncly-standalone, so the kit cannot check its answer. Fix: drop exec",
				"testdata/script/keys.txtar:11: exec runs ncly, so the kit cannot check its answer",
			},
			absent: []string{"keys.txtar:12:"},
		},
		{
			name: "a line of spec.md without an ID, with the ID of another, or without a claim",
			edits: []edit{
				{"spec.md", "| `tally-fails` |", "| tally-fails |"},
				{"spec.md", "| `key-missing` |", "| `not-granted` |"},
				{"testdata/script/tally.txtar", "# spec: tally-missing\n", ""},
			},
			want: []string{
				"spec.md:14: this line of Failures has no ID. Fix: give it an ID in backticks",
				"spec.md:25: the ID `not-granted` repeats the one of line 24. Fix: give each line its own ID",
				"spec.md:13: no test claims `tally-missing`. Fix: write the line `# spec: tally-missing` above the commands that prove it",
				"testdata/script/keys.txtar:16: `# spec: key-missing` names no line of Failures or Acceptance in spec.md. Fix: correct the ID",
			},
			absent: []string{"no test claims `not-granted`", "no test claims `no-keychain`"},
		},
		{
			name: "a claim of acceptance.sh covers a line, and one names no line",
			edits: []edit{
				{"spec.md", "## ncly standalone fetch", "## Acceptance\n\n- `total`: The count of a real input\n\n## ncly standalone fetch"},
				{file: "acceptance.sh", new: "#!/usr/bin/env bash\n# spec: total\n# spec: ghost\n"},
			},
			want:   []string{"acceptance.sh:3: `# spec: ghost` names no line of Failures or Acceptance in spec.md"},
			absent: []string{"no test claims `total`"},
		},
		{
			name:  "a repository without spec.md",
			edits: []edit{{file: "spec.md"}},
			want:  []string{"spec.md: the repository has no spec. Fix: add spec.md"},
		},
		{
			name:  "an extension without a skill",
			edits: []edit{{"extension.go", "\tSkill:    skill,\n", ""}},
			want:  []string{"skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named standalone"},
		},
		{
			name:    "issue 335 stdout mentions an unexecuted command",
			fixture: "standalone-named",
			edits:   []edit{{file: "testdata/script/convert.txtar", new: "ncly skill view video-convert\nstdout 'ncly video convert clip.mov'\n"}},
			want:    []string{"testdata/script: no scenario runs `ncly video convert`. Fix: add a scenario that runs it"},
		},
		{
			name:    "issue 335 path-qualified exits bypasses a broken schema",
			fixture: "standalone-named",
			edits: []edit{
				{"manifest.toml", `file = { type = "string" }`, `file = { type = "integer" }`},
				{file: "testdata/script/convert.txtar", new: "ncly skill view video-convert\nstdout 'ncly video convert clip.mov'\ngrant example\nexits 0 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert clip.mov --json\n"},
			},
			want: []string{"use the registered command ncly-video-convert instead of its path"},
		},
		{
			name:    "a sibling cannot cover the root",
			fixture: "standalone-named",
			edits:   []edit{{file: "testdata/script/convert.txtar", new: "ncly video convert-hdr clip.mov --json\n"}},
			want:    []string{"testdata/script: no scenario runs `ncly video convert`. Fix: add a scenario that runs it"},
		},
		{
			name:    "a direct answer must match the root schema",
			fixture: "standalone-named",
			edits: []edit{
				{"manifest.toml", `file = { type = "string" }`, `file = { type = "integer" }`},
				{file: "testdata/script/convert.txtar", new: "ncly-video-convert clip.mov --json\n"},
			},
			want: []string{"`ncly-video-convert clip.mov --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly video convert`"},
		},
		{
			name:    "a host answer must match the root schema",
			fixture: "standalone-named",
			edits:   []edit{{"manifest.toml", `file = { type = "string" }`, `file = { type = "integer" }`}},
			want: []string{
				"`ncly video convert clip.mov --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly video convert`",
				"manifest.toml: example `ncly video convert clip.mov --json`: the answer on stdout breaks the schema",
			},
		},
		{
			name:    "a direct foreground exec cannot skip the answer check",
			fixture: "standalone-named",
			edits:   []edit{{file: "testdata/script/convert.txtar", new: "exec $XDG_DATA_HOME/nicely/extensions/ncly-video-convert clip.mov --json\n"}},
			want:    []string{"convert.txtar:1: exec runs ncly-video-convert, so the kit cannot check its answer. Fix: drop exec"},
		},
		{
			name:    "a missing skill names the slug",
			fixture: "standalone-named",
			edits:   []edit{{"extension.go", "\tSkill:    skill,\n", ""}},
			want:    []string{"skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named video-convert"},
		},
		{
			name:    "a named skill that names an undeclared command",
			fixture: "standalone-named",
			edits: []edit{
				{"manifest.toml", "path = \"video convert\"", "path = \"video convert\"\nsummary = \"video-convert.summary\"\n\n[[commands]]\npath = \"video convert run\""},
				{"extension.go", `"video convert": convert`, `"video convert run": convert`},
				{file: "testdata/script/convert.txtar", new: "grant example\nncly video convert run clip.mov --json\nstdout '\"file\":\"clip.mov\"'\n"},
				{"skill/SKILL.md", "ncly video convert clip.mov", "ncly video convert typo clip.mov"},
			},
			want: []string{"skill/SKILL.md: names `ncly video convert typo`, which the manifest lacks. Fix: declare that command in manifest.toml"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := goTest(t, cmp.Or(c.fixture, "standalone"), os.Getenv("PATH"), c.edits...)
			t.Logf("fixture go test: %v\n%s", err, out)
			if err == nil {
				t.Fatalf("go test passed on the broken copy:\n%s", out)
			}
			for _, want := range c.want {
				if !strings.Contains(out, want) {
					t.Errorf("go test failed without %q:\n%s", want, out)
				}
			}
			for _, absent := range c.absent {
				if strings.Contains(out, absent) {
					t.Errorf("go test failed with %q:\n%s", absent, out)
				}
			}
		})
	}
}
