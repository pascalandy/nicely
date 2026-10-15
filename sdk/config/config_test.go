package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func env(vars map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, set := vars[name]
		return value, set
	}
}

func TestLocate(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Paths
	}{
		{"defaults under home", nil, Paths{
			Shared: "/Users/me/.config/nicely/config.toml",
			Local:  "/Users/me/.config/nicely/config.local.toml",
			Data:   "/Users/me/.local/share/nicely",
			State:  "/Users/me/.local/state/nicely",
			Cache:  "/Users/me/.cache/nicely",
		}},
		{"absolute XDG folders win", map[string]string{
			"XDG_CONFIG_HOME": "/xdg/config", "XDG_DATA_HOME": "/xdg/data",
			"XDG_STATE_HOME": "/xdg/state", "XDG_CACHE_HOME": "/xdg/cache",
		}, Paths{
			Shared: "/xdg/config/nicely/config.toml",
			Local:  "/xdg/config/nicely/config.local.toml",
			Data:   "/xdg/data/nicely",
			State:  "/xdg/state/nicely",
			Cache:  "/xdg/cache/nicely",
		}},
		{"a relative XDG folder is ignored", map[string]string{"XDG_CONFIG_HOME": "relative"}, Paths{
			Shared: "/Users/me/.config/nicely/config.toml",
			Local:  "/Users/me/.config/nicely/config.local.toml",
			Data:   "/Users/me/.local/share/nicely",
			State:  "/Users/me/.local/state/nicely",
			Cache:  "/Users/me/.cache/nicely",
		}},
		{"NCLY_CONFIG moves both config files", map[string]string{"NCLY_CONFIG": "/Users/me/dotfiles/ncly.toml"}, Paths{
			Shared: "/Users/me/dotfiles/ncly.toml",
			Local:  "/Users/me/dotfiles/config.local.toml",
			Data:   "/Users/me/.local/share/nicely",
			State:  "/Users/me/.local/state/nicely",
			Cache:  "/Users/me/.cache/nicely",
		}},
	}
	for _, c := range cases {
		if got := Locate("/Users/me", env(c.env)); got != c.want {
			t.Errorf("%s:\n got %+v\nwant %+v", c.name, got, c.want)
		}
	}
	if got := Locate("", env(nil)); got != (Paths{}) {
		t.Errorf("without a home, Locate = %+v, want no paths", got)
	}
	got := Locate("", env(map[string]string{"XDG_CONFIG_HOME": "/xdg/config"}))
	if got.Shared != "/xdg/config/nicely/config.toml" || got.Data != "" {
		t.Errorf("without a home but with XDG_CONFIG_HOME, Locate = %+v", got)
	}
}

func files(t *testing.T, shared, local string) Paths {
	t.Helper()
	dir := t.TempDir()
	p := Paths{Shared: filepath.Join(dir, "config.toml"), Local: filepath.Join(dir, "config.local.toml")}
	for path, text := range map[string]string{p.Shared: shared, p.Local: local} {
		if text != "" {
			if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	return p
}

func TestLoad(t *testing.T) {
	cases := []struct {
		name          string
		shared, local string
		want          Config
	}{
		{"no file gives the defaults", "", "", Config{}},
		{"the shared file sets a key", `lang = "fr-CA"`, "", Config{Lang: "fr-CA"}},
		{"the local file wins", `lang = "fr-CA"`, `lang = "en"`, Config{Lang: "en"}},
		{"a local file alone works", "", `lang = "en"`, Config{Lang: "en"}},
	}
	for _, c := range cases {
		got, unknown, err := Load(files(t, c.shared, c.local))
		if err != nil || got != c.want || unknown != nil {
			t.Errorf("%s: Load = %+v, %v, %v; want %+v", c.name, got, unknown, err, c.want)
		}
	}
}

func TestLoadCollectsUnknownKeysPerFile(t *testing.T) {
	p := files(t, "lang = \"en\"\n[skill]\npaths = [\"~/code/skills\"]\n", "future = 1\n")
	_, unknown, err := Load(p)
	want := []Key{{p.Shared, "skill"}, {p.Shared, "skill.paths"}, {p.Local, "future"}}
	if err != nil || !reflect.DeepEqual(unknown, want) {
		t.Errorf("unknown keys %v, %v; want %v", unknown, err, want)
	}
}

func TestLoadRejectsAnInvalidFile(t *testing.T) {
	cases := []struct {
		name          string
		shared, local string
		inLocal       bool
		line          int
	}{
		{"a syntax error", "\nlang =\n", "", false, 2},
		{"a value of the wrong type", "lang = 1\n", "", false, 1},
		{"an error in the local file names that file", `lang = "en"`, "x = [\n", true, 1},
		{"a number that is not finite", "x = nan\n", "", false, 0},
		{"infinity inside a list", "[a]\nb = [1.5, inf]\n", "", false, 0},
		{"infinity in a table of an array", "[[a]]\nb = { c = -inf }\n", "", false, 0},
		{"a number that is not finite under an empty key", "\"\" = nan\n", "", false, 0},
	}
	for _, c := range cases {
		p := files(t, c.shared, c.local)
		_, _, err := Load(p)
		var invalid *InvalidError
		if !errors.As(err, &invalid) {
			t.Errorf("%s: Load error %v, want an InvalidError", c.name, err)
			continue
		}
		file := p.Shared
		if c.inLocal {
			file = p.Local
		}
		if invalid.File != file || invalid.Line != c.line {
			t.Errorf("%s: InvalidError in %s at line %d, want %s at line %d", c.name, invalid.File, invalid.Line, file, c.line)
		}
	}
}

func TestMergeJoinsTablesAndReplacesLists(t *testing.T) {
	shared := map[string]any{
		"lang":  "fr-CA",
		"skill": map[string]any{"paths": []any{"~/a", "~/b"}},
		"agent": map[string]any{"profiles": map[string]any{
			"everyday": map[string]any{"harness": "claude", "effort": "medium"},
		}},
	}
	local := map[string]any{
		"skill": map[string]any{"paths": []any{"~/c"}},
		"agent": map[string]any{"profiles": map[string]any{
			"everyday": map[string]any{"effort": "high"},
		}},
	}
	merge(shared, local)
	want := map[string]any{
		"lang":  "fr-CA",
		"skill": map[string]any{"paths": []any{"~/c"}},
		"agent": map[string]any{"profiles": map[string]any{
			"everyday": map[string]any{"harness": "claude", "effort": "high"},
		}},
	}
	if !reflect.DeepEqual(shared, want) {
		t.Errorf("merged\n%v\nwant\n%v", shared, want)
	}
}

func TestLoadValuesKeepsTheSourceOfEachValue(t *testing.T) {
	p := files(t,
		"lang = \"fr-CA\"\n[skill]\npaths = [\"~/a\"]\n[agent.profiles.everyday]\nharness = \"claude\"\neffort = \"medium\"\n[agent.models]\nfast = \"a\"\n",
		"[skill]\npaths = []\n[agent.profiles.everyday]\neffort = \"high\"\n[agent]\nmodels = \"none\"\n",
	)
	got, err := LoadValues(p)
	if err != nil {
		t.Fatal(err)
	}
	want := Values{
		"lang":                            {"lang", "fr-CA", p.Shared},
		"skill.paths":                     {"skill.paths", []any{}, p.Local},
		"agent.profiles.everyday.harness": {"agent.profiles.everyday.harness", "claude", p.Shared},
		"agent.profiles.everyday.effort":  {"agent.profiles.everyday.effort", "high", p.Local},
		"agent.models":                    {"agent.models", "none", p.Local},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadValues\n%v\nwant\n%v", got, want)
	}
	under := got.Under("agent.profiles")
	if len(under) != 2 || under[0].Key != "agent.profiles.everyday.effort" || under[1].Key != "agent.profiles.everyday.harness" {
		t.Errorf("Under(agent.profiles) = %v, want both keys of everyday in order", under)
	}
	if unset := got.Under("agent.default_profile"); !reflect.DeepEqual(unset, []Value{{Key: "agent.default_profile", Source: Default}}) {
		t.Errorf("Under(agent.default_profile) = %v, want the default", unset)
	}
	if _, err := LoadValues(files(t, "lang = 1\n", "")); err == nil {
		t.Error("LoadValues read a value of the wrong type")
	}
}

func TestLangNamesItsSource(t *testing.T) {
	file := Values{"lang": {"lang", "fr-CA", "/Users/me/.config/nicely/config.toml"}}
	cases := []struct {
		name  string
		flag  string
		env   map[string]string
		files Values
		want  Value
	}{
		{"the flag wins", "en", map[string]string{"NCLY_LANG": "fr"}, file, Value{"lang", "en", "--lang"}},
		{"then the variable", "", map[string]string{"NCLY_LANG": "fr", "LANG": "en_US.UTF-8"}, file, Value{"lang", "fr", "NCLY_LANG"}},
		{"then the files", "", map[string]string{"LANG": "en_US.UTF-8"}, file, file["lang"]},
		{"then the locale variables in order", "", map[string]string{"LC_MESSAGES": "fr_CA.UTF-8", "LANG": "en_US.UTF-8"}, nil, Value{"lang", "fr_CA.UTF-8", "LC_MESSAGES"}},
		{"and en by default", "", nil, nil, Value{"lang", "en", Default}},
	}
	for _, c := range cases {
		read := false
		got := Lang(c.flag, func(name string) string { return c.env[name] }, func() Values { read = true; return c.files })
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: Lang = %v, want %v", c.name, got, c.want)
		}
		if read && (c.flag != "" || c.env["NCLY_LANG"] != "") {
			t.Errorf("%s: Lang read the files although a flag or NCLY_LANG sets the language", c.name)
		}
	}
}
