package extensiontest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// An edit changes the first old of a file of the fixture into new. An edit
// without old removes its file, and one with new alone writes it.
type edit struct{ file, old, new string }

// TestEachCheckFailsAndNamesItsFix breaks a copy of the standalone fixture
// in one way per case, then runs its go test. That test must fail with the
// fix of the check that the break trips. The fixture's own go test, which
// just test runs, proves that the unbroken repository passes. The cases run
// one at a time, because each go test builds and links. In parallel, they
// slowed the pty scenarios of cmd/ncly past their limit on a 12-core Mac.
func TestEachCheckFailsAndNamesItsFix(t *testing.T) {
	breakSchema := edit{"manifest.toml", `count = { type = "integer" }`, `count = { type = "string" }`}
	cases := []struct {
		name  string
		edits []edit
		want  []string
		// absent holds what the output must not hold.
		absent []string
	}{
		{
			name:   "a manifest that names a missing catalog ID",
			edits:  []edit{{"manifest.toml", "standalone.tally.summary", "standalone.missing.summary"}},
			want:   []string{`locales/en.toml lacks the ID "standalone.missing.summary" that the manifest names.`},
			absent: []string{"map[", "<no value>"},
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
				"testdata/script/keys.txtar:10: exec runs ncly, so the kit cannot check its answer",
			},
			absent: []string{"keys.txtar:11:"},
		},
		{
			name:  "an extension without a skill",
			edits: []edit{{"extension.go", "\tSkill:    skill,\n", ""}},
			want:  []string{"skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named standalone"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := goTest(t, os.Getenv("PATH"), c.edits...)
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

// TestTheScenariosIgnoreTheCallersExtensions puts an extension in a folder
// of the caller's PATH, which discovery would find. A scenario that lists
// the extensions must see only the one under test.
func TestTheScenariosIgnoreTheCallersExtensions(t *testing.T) {
	outside := t.TempDir()
	for name, content := range map[string]string{
		"ncly-outsider":                 "#!/bin/sh\nexit 0\n",
		"ncly-outsider.toml":            "format_version = 1\ncontract_version = 1\ndomain = \"outsider\"\n\n[[commands]]\npath = \"outsider\"\nsummary = \"outsider.summary\"\n\n[[commands]]\npath = \"outsider run\"\nsummary = \"outsider.run.summary\"\n",
		"ncly-outsider.locales/en.toml": "[\"outsider.summary\"]\ndescription = \"d\"\nother = \"Outside the test\"\n\n[\"outsider.run.summary\"]\ndescription = \"d\"\nother = \"Run outside\"\n",
	} {
		path := filepath.Join(outside, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	scenario := edit{file: "testdata/script/outsider.txtar", new: "ncly --help\nstdout 'standalone'\n! stdout 'outsider'\n"}
	if out, err := goTest(t, outside+string(filepath.ListSeparator)+os.Getenv("PATH"), scenario); err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}

// goTest copies the standalone fixture, points its replace at this checkout,
// applies the edits, and runs its go test with path as PATH.
func goTest(t *testing.T, path string, edits ...edit) (string, error) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join(root, "testdata", "extensions", "standalone"))); err != nil {
		t.Fatal(err)
	}
	if out, err := command(dir, os.Getenv("PATH"), "go", "mod", "edit", "-replace", "github.com/pascalandy/nicely="+root); err != nil {
		t.Fatalf("go mod edit: %v\n%s", err, out)
	}
	for _, e := range edits {
		file := filepath.Join(dir, e.file)
		var err error
		switch data, readErr := os.ReadFile(file); {
		case e.old == "" && e.new == "":
			err = os.Remove(file)
		case e.old == "":
			err = os.WriteFile(file, []byte(e.new), 0o644)
		case readErr != nil || !strings.Contains(string(data), e.old):
			t.Fatalf("%s lacks %q: %v", e.file, e.old, readErr)
		default:
			err = os.WriteFile(file, []byte(strings.Replace(string(data), e.old, e.new, 1)), 0o644)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return command(dir, path, "go", "test", "-count=1", "./...")
}

// command runs a program in dir with path as PATH, and returns its output.
func command(dir, path, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
