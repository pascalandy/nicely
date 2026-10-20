package extension

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestExportWritesTheSidecarsOfTheValue proves that Export writes what the
// value embeds, and that an export over a previous one keeps no file that
// the new value lacks. The test kit's scratch install proves the rest:
// ncly finds the extension through these sidecars.
func TestExportWritesTheSidecarsOfTheValue(t *testing.T) {
	dir := t.TempDir()
	e := demo()
	e.Skill = fstest.MapFS{"skill/SKILL.md": {Data: []byte(validSkill)}, "skill/notes.md": {Data: []byte("old\n")}}
	if err := Export(e, dir); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"ncly-demo.toml":            validManifest,
		"ncly-demo.locales/en.toml": validCatalog,
		"ncly-demo.skill/SKILL.md":  validSkill,
		"ncly-demo.skill/notes.md":  "old\n",
	} {
		if got, err := os.ReadFile(filepath.Join(dir, path)); err != nil || string(got) != want {
			t.Errorf("%s holds %q (%v), want %q", path, got, err, want)
		}
	}

	if err := Export(demo(), dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ncly-demo.skill", "notes.md")); !os.IsNotExist(err) {
		t.Errorf("a second export kept a file of the first skill: %v", err)
	}
	e = demo()
	e.Skill = nil
	if err := Export(e, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ncly-demo.skill")); !os.IsNotExist(err) {
		t.Errorf("an export without a skill kept the skill folder: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Errorf("want only ncly-demo.toml and ncly-demo.locales, got %v (%v)", entries, err)
	}
}

func TestExportRefusesAnInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	e := demo()
	e.Manifest = []byte("format_version = 2\n")
	if err := Export(e, dir); err == nil {
		t.Fatal("Export accepted a manifest that it cannot read")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("Export wrote %v for an invalid manifest", entries)
	}
}
