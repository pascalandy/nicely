package extension

import (
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/sdk/contract"
)

func TestWarningsAppearInDeclaredCodes(t *testing.T) {
	e := demo()
	e.Manifest = []byte(validManifest + `
[[warnings]]
code = "DEMO_NOTICE"
commands = ["demo run"]
modes = ["run"]
untested = "Needs a controlled host: https://github.com/OWNER/REPO/issues/123"
`)
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(m.Registry().Codes(), contract.Code("DEMO_NOTICE")) {
		t.Fatal("the registry's code list omits the declared warning DEMO_NOTICE")
	}
	if _, errorCode := m.Registry().Exit("DEMO_NOTICE"); errorCode {
		t.Fatal("a declared warning must not map to an exit code")
	}
	if len(m.DeclaredCodes) != 2 {
		t.Fatalf("declared codes = %+v, want the error and the warning", m.DeclaredCodes)
	}
	legacy, warning := m.DeclaredCodes[0], m.DeclaredCodes[1]
	if legacy.Code != "DEMO_FAILED" || legacy.Warning || legacy.Commands != nil || legacy.Modes != nil || legacy.Untested != "" {
		t.Errorf("the legacy error lost its default scope: %+v", legacy)
	}
	if warning.Code != "DEMO_NOTICE" || !warning.Warning || !slices.Equal(warning.Commands, []string{"demo run"}) ||
		!slices.Equal(warning.Modes, []string{"run"}) || warning.Untested != "Needs a controlled host: https://github.com/OWNER/REPO/issues/123" ||
		warning.Line != 17 || warning.Column != 3 {
		t.Errorf("the warning lost its scope, reason, or TOML position: %+v", warning)
	}
}

func TestCheckRejectsInvalidWarnings(t *testing.T) {
	cases := []struct{ name, declaration, want string }{
		{"wrong prefix", "[[warnings]]\ncode = 'OTHER_NOTICE'\n", "must use uppercase letters, digits, and underscores, starting with DEMO_"},
		{"invalid spelling", "[[warnings]]\ncode = 'DEMO_notice'\n", "must use uppercase letters, digits, and underscores"},
		{"exit on a warning", "[[warnings]]\ncode = 'DEMO_NOTICE'\nexit = 0\n", "must not declare exit"},
		{"duplicate warning", "[[warnings]]\ncode = 'DEMO_NOTICE'\n[[warnings]]\ncode = 'DEMO_NOTICE'\n", "appears twice across errors and warnings"},
		{"error as a warning", "[[warnings]]\ncode = 'DEMO_FAILED'\n", "appears twice across errors and warnings"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := demo()
			e.Manifest = []byte(validManifest + c.declaration)
			err := Check(e)
			if err == nil || !strings.Contains(err.Error(), "manifest.toml:") || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Check = %v, want a TOML position and %q", err, c.want)
			}
		})
	}
}

func TestCheckAcceptsCodeScopes(t *testing.T) {
	for _, fields := range []string{
		"",
		"commands = ['demo run']\n",
		"modes = ['run']\n",
		"commands = ['demo run']\nmodes = ['dry_run']\n",
		"commands = ['demo run']\nmodes = ['run', 'dry_run']\n",
	} {
		t.Run(fields, func(t *testing.T) {
			e := demo()
			manifest := strings.Replace(validManifest, `summary = "demo.run.summary"`, `summary = "demo.run.summary"`+"\nmodes = {dry_run = true}", 1)
			manifest = strings.Replace(manifest, "[[errors]]", "[[commands]]\npath = 'demo read'\nsummary = 'demo.run.summary'\n\n[[errors]]", 1)
			e.Handlers["demo read"] = noop
			e.Manifest = []byte(manifest + fields)
			if err := Check(e); err != nil {
				t.Fatal(err)
			}
		})
	}
}
