package extensiontest_test

import (
	"os"
	"path/filepath"
	"testing"
)

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
	if out, err := goTest(t, "standalone", outside+string(filepath.ListSeparator)+os.Getenv("PATH"), scenario); err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}
