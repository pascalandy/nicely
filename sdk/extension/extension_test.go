package extension

import (
	"context"
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
