package extensiontest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestEachCheckFailsAndNamesItsFix breaks a copy of the standalone fixture
// in one way per case, then runs its go test. That test must fail with the
// fix of the check that the break trips. The fixture's own go test, which
// just test runs, proves that the unbroken repository passes.
func TestEachCheckFailsAndNamesItsFix(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// An edit without old removes its file.
	type edit struct{ file, old, new string }
	breakSchema := edit{"manifest.toml", `count = { type = "integer" }`, `count = { type = "string" }`}
	cases := []struct {
		name  string
		edits []edit
		want  string
	}{
		{
			name:  "a command without a scenario",
			edits: []edit{{file: "testdata/script/keys.txtar"}},
			want:  "testdata/script: no scenario runs `ncly standalone fetch`. Fix: add a scenario that runs it",
		},
		{
			name:  "a skill that names an undeclared command",
			edits: []edit{{"skill/SKILL.md", "Add `--json`", "Sum with `ncly standalone total`.\nAdd `--json`"}},
			want:  "skill/SKILL.md: names `ncly standalone total`, which the manifest lacks. Fix: declare that command in manifest.toml",
		},
		{
			name:  "an answer that breaks its schema",
			edits: []edit{breakSchema},
			want:  "tally.txtar:4: `ncly standalone tally --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly standalone tally`",
		},
		{
			name:  "an answer that breaks its schema, after a flag before the domain",
			edits: []edit{breakSchema, {"testdata/script/tally.txtar", "ncly standalone tally --json\nstdout '\"count\":3'\n", ""}},
			want:  "tally.txtar:6: `ncly --json standalone tally`: the answer on stdout breaks the schema",
		},
		{
			name:  "a Nicely version in .copier-answers.yml that differs from go.mod",
			edits: []edit{{".copier-answers.yml", "_commit: v0.18.0", "_commit: v0.17.0"}},
			want:  `.copier-answers.yml: the template comes from Nicely "v0.17.0", and go.mod pins Nicely v0.18.0. Fix: run just update v0.18.0`,
		},
		{
			name:  "a call that skips the answer check",
			edits: []edit{{"testdata/script/tally.txtar", "\nncly standalone tally --json\n", "\nexec ncly standalone tally --json\n"}},
			want:  "testdata/script/tally.txtar:4: exec runs ncly, so the kit cannot check its answer. Fix: drop exec",
		},
		{
			name:  "an extension without a skill",
			edits: []edit{{"extension.go", "\tSkill:    skill,\n", ""}},
			want:  "skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named standalone",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS(filepath.Join(root, "testdata", "extensions", "standalone"))); err != nil {
				t.Fatal(err)
			}
			run(t, dir, "go", "mod", "edit", "-replace", "github.com/pascalandy/nicely="+root)
			for _, e := range c.edits {
				path := filepath.Join(dir, e.file)
				if e.old == "" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					continue
				}
				data, err := os.ReadFile(path)
				if err != nil || !strings.Contains(string(data), e.old) {
					t.Fatalf("%s lacks %q: %v", e.file, e.old, err)
				}
				if err := os.WriteFile(path, []byte(strings.Replace(string(data), e.old, e.new, 1)), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("go", "test", "-count=1", "./...")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("go test passed on the broken copy:\n%s", out)
			}
			if !strings.Contains(string(out), c.want) {
				t.Errorf("go test failed without %q:\n%s", c.want, out)
			}
		})
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
}
