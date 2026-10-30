package extensiontest_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallLifecycleIgnoresCallerSettings(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(config, []byte("[invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NCLY_CONFIG", config)
	t.Setenv("NCLY_LANG", "xx-YY")
	if out, err := goTest(t, "standalone", os.Getenv("PATH")); err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}

// A sibling can keep a shared domain present after scenario cleanup.
func TestScenarioCanLeaveSiblingExtension(t *testing.T) {
	scenario := edit{file: "testdata/script/sibling.txtar", new: `
env EXTENSIONS=$XDG_DATA_HOME/nicely/extensions
fake standalone-other
cp bin/standalone-other $EXTENSIONS/ncly-standalone-other
cp sibling.toml $EXTENSIONS/ncly-standalone-other.toml
mkdir $EXTENSIONS/ncly-standalone-other.locales
cp en.toml $EXTENSIONS/ncly-standalone-other.locales/en.toml

# The whole-domain extension hides the sibling until its executable goes.
ncly describe standalone --json
! stdout '"path":"standalone other"'
rm $EXTENSIONS/ncly-standalone
ncly describe standalone other --json
stdout '"path":"standalone other"'

-- sibling.toml --
format_version = 1
contract_version = 1
domain = "standalone"
name = "other"
[[commands]]
path = "standalone other"
summary = "standalone-other.summary"
-- en.toml --
["standalone-other.summary"]
description = "Summary of the sibling fixture."
other = "Run the sibling"
`}
	if out, err := goTest(t, "standalone", os.Getenv("PATH"), scenario); err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}
