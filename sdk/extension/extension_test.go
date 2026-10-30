package extension

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/pascalandy/nicely/sdk/contract"
)

const validManifest = `format_version = 1
contract_version = 1
domain = "demo"

[[commands]]
path = "demo"
summary = "demo.summary"

[[commands]]
path = "demo run"
summary = "demo.run.summary"

[[errors]]
code = "DEMO_FAILED"
exit = 1
`

const validCatalog = `["demo.summary"]
description = "Summary of the demo extension."
other = "Demo"

["demo.run.summary"]
description = "Summary of demo run."
other = "Run the demo"
`

const validSkill = `---
name: demo
description: Use when running the demo.
---
`

func noop(context.Context, Call) contract.Outcome { return contract.Outcome{} }

// demo returns a valid extension, which each case breaks in one way.
func demo() Extension {
	return Extension{
		Manifest: []byte(validManifest),
		Locales:  fstest.MapFS{"locales/en.toml": {Data: []byte(validCatalog)}},
		Skill:    fstest.MapFS{"skill/SKILL.md": {Data: []byte(validSkill)}},
		Handlers: map[string]Handler{"demo run": noop},
	}
}

func TestCheckAcceptsAnExtensionWithoutSkill(t *testing.T) {
	e := demo()
	e.Skill = nil
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAcceptsAValidExtension(t *testing.T) {
	if err := Check(demo()); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsInvalidCodeDeclarations(t *testing.T) {
	prefix := validManifest[:strings.Index(validManifest, "[[errors]]")]
	cases := []struct{ name, field, want string }{
		{"empty commands", `commands = []`, "commands must name at least one command that runs"},
		{"empty modes", `modes = []`, "modes must contain run or dry_run"},
		{"unknown command", `commands = ["demo missing"]`, `command "demo missing" must name a declared command that runs`},
		{"group command", `commands = ["demo"]`, `command "demo" must name a declared command that runs`},
		{"duplicate command", `commands = ["demo run", "demo run"]`, `command "demo run" appears twice`},
		{"unknown mode", `modes = ["typo"]`, `mode "typo" must be run or dry_run`},
		{"duplicate mode", `modes = ["run", "run"]`, `mode "run" appears twice`},
		{"unsupported dry run", `modes = ["dry_run"]`, `command "demo run" does not offer dry_run`},
		{"empty exemption", `untested = ""`, "untested must give a reason and an issue URL"},
		{"blank exemption", `untested = "  "`, "untested must give a reason and an issue URL"},
		{"exemption without issue", `untested = "Needs a controlled host"`, "untested must give a reason and an issue URL"},
		{"exemption without reason", `untested = "https://github.com/OWNER/REPO/issues/123"`, "untested must give a reason and an issue URL"},
		{"exemption without issue number", `untested = "Needs a controlled host: https://github.com/OWNER/REPO/issues/0"`, "untested must give a reason and an issue URL"},
	}
	for _, kind := range []string{"errors", "warnings"} {
		for _, c := range cases {
			t.Run(kind+"/"+c.name, func(t *testing.T) {
				e := demo()
				declaration := "[[" + kind + "]]\ncode = \"DEMO_FAILED\"\n"
				if kind == "errors" {
					declaration += "exit = 1\n"
				}
				e.Manifest = []byte(prefix + declaration + c.field + "\n")
				err := Check(e)
				if err == nil {
					t.Fatal("Check accepted an invalid code declaration")
				}
				for _, want := range []string{"manifest.toml:13:3:", "DEMO_FAILED", c.want} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("Check = %v, want %q", err, want)
					}
				}
			})
		}
	}
}

func TestCheckUsesTOMLCodePositions(t *testing.T) {
	cases := []struct {
		name, manifest string
		line, column   int
	}{
		{
			name: "quoted header after a multiline string and a comment",
			manifest: "notes = '''\n[[warnings]]\n'''\n# [[warnings]]\n" + validManifest +
				"  [[ \"warnings\" ]] # a real declaration\ncode = 'DEMO_NOTICE'\nmodes = ['typo']\n",
			line: 20, column: 6,
		},
		{
			name:     "second inline declaration",
			manifest: "warnings = [\n {code = 'DEMO_FIRST'},\n {code = 'DEMO_NOTICE', modes = ['typo']}\n]\n" + validManifest,
			line:     3, column: 2,
		},
		{
			name: "escaped key and CRLF",
			manifest: strings.ReplaceAll(validManifest+
				"[[\"warn\\u0069ngs\"]]\ncode = 'DEMO_NOTICE'\nmodes = ['typo']\n", "\n", "\r\n"),
			line: 16, column: 3,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := demo()
			e.Manifest = []byte(c.manifest)
			err := Check(e)
			position := fmt.Sprintf("manifest.toml:%d:%d:", c.line, c.column)
			if err == nil || !strings.Contains(err.Error(), position) || !strings.Contains(err.Error(), `mode "typo"`) {
				t.Fatalf("Check = %v, want %s and the invalid mode", err, position)
			}
		})
	}
}

func TestCheckRejectsWhatMainWouldRefuse(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*Extension)
		want  string
	}{
		{"an own code without the domain prefix", func(e *Extension) {
			e.Manifest = []byte(strings.Replace(validManifest, "DEMO_FAILED", "FAILED", 1))
		}, "manifest.toml declares the error code FAILED, which does not start with DEMO_"},
		{"invalid TOML in the manifest", func(e *Extension) {
			e.Manifest = []byte("format_version = [")
		}, "manifest.toml has a TOML syntax error on line 1"},
		{"an invalid manifest field type", func(e *Extension) {
			e.Manifest = []byte(strings.Replace(validManifest, "format_version = 1", `format_version = "one"`, 1))
		}, "manifest.toml is not in the expected format"},
		{"an invalid domain name", func(e *Extension) {
			e.Manifest = []byte(strings.Replace(validManifest, `domain = "demo"`, `domain = "my-tool"`, 1))
		}, `manifest.toml declares the invalid domain "my-tool". Use lowercase letters and digits, starting with a letter`},
		{"a command that runs without a handler", func(e *Extension) {
			e.Handlers = map[string]Handler{}
		}, `"demo run" runs but has no handler`},
		{"a command that runs with a nil handler", func(e *Extension) {
			e.Handlers["demo run"] = nil
		}, `"demo run" runs but has no handler`},
		{"a handler for no command", func(e *Extension) {
			e.Handlers["demo stop"] = noop
		}, `"demo stop" has no command that runs`},
		{"a handler for a group that only prints its help", func(e *Extension) {
			e.Handlers["demo"] = noop
		}, `"demo" has no command that runs`},
		{"a catalog ID without the domain prefix", func(e *Extension) {
			e.Locales = fstest.MapFS{"locales/en.toml": {Data: []byte(validCatalog + "\n[\"other.text\"]\ndescription = \"d\"\nother = \"o\"\n")}}
		}, `other.text does not start with "demo."`},
		{"an entry without a description", func(e *Extension) {
			e.Locales = fstest.MapFS{"locales/en.toml": {Data: []byte(strings.Replace(validCatalog, `description = "Summary of demo run."`, "", 1))}}
		}, "demo.run.summary has no description"},
		{"a translation that lacks an ID of en", func(e *Extension) {
			e.Locales = fstest.MapFS{
				"locales/en.toml":    {Data: []byte(validCatalog)},
				"locales/fr-CA.toml": {Data: []byte("[\"demo.summary\"]\ndescription = \"d\"\nother = \"Démo\"\n")},
			}
		}, "fr-CA.toml lacks demo.run.summary"},
		{"a skill named after another domain", func(e *Extension) {
			e.Skill = fstest.MapFS{"skill/SKILL.md": {Data: []byte(strings.Replace(validSkill, "name: demo", "name: other", 1))}}
		}, `named "other", not after the slug "demo"`},
		{"a skill without description", func(e *Extension) {
			e.Skill = fstest.MapFS{"skill/SKILL.md": {Data: []byte("---\nname: demo\n---\n")}}
		}, "lacks description"},
		{"a skill without SKILL.md", func(e *Extension) {
			e.Skill = fstest.MapFS{"SKILL.md": {Data: []byte(validSkill)}}
		}, "skill/SKILL.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := demo()
			c.spoil(&e)
			err := Check(e)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Check = %v, want an error about %s", err, c.want)
			}
			if err != nil && strings.Contains(err.Error(), "map[") {
				t.Errorf("Check exposes rule data: %v", err)
			}
		})
	}
}
