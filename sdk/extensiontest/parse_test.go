package extensiontest

import (
	"slices"
	"testing"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// demo has a group that runs, demo run, which takes an argument and holds
// demo run fast, and a group that only groups, demo pick.
const demo = `format_version = 1
contract_version = 1
domain = "demo"

[[commands]]
path = "demo"
summary = "demo.summary"

[[commands]]
path = "demo run"
summary = "demo.run.summary"

[[commands.args]]
name = "name"
summary = "demo.run.name"

[[commands]]
path = "demo run fast"
summary = "demo.run.fast.summary"

[[commands]]
path = "demo pick"
summary = "demo.pick.summary"

[[commands]]
path = "demo pick one"
summary = "demo.pick.one.summary"
`

func demoKit(t *testing.T) *kit {
	t.Helper()
	m, err := contract.ParseManifest([]byte(demo), "")
	if err != nil {
		t.Fatal(err)
	}
	catalogs, err := i18n.Load(i18n.Locales())
	if err != nil {
		t.Fatal(err)
	}
	return &kit{manifest: m, printer: catalogs.Printer("en")}
}

// TestMentionsTellACommandFromAnArgument reads the word after demo run,
// which runs, as an argument unless it names a child. The word after demo
// pick, which only groups, is always a command, so a mistyped one shows.
func TestMentionsTellACommandFromAnArgument(t *testing.T) {
	k := demoKit(t)
	for text, want := range map[string][]string{
		"`ncly demo run folder`":     {"demo run"},
		"`ncly demo run fast`":       {"demo run fast"},
		"`ncly demo pick one`":       {"demo pick one"},
		"`ncly demo pick two`":       {"demo pick two"},
		"`ncly demo runs`":           {"demo runs"},
		"ncly demo, then ncly other": {"demo"},
		"`ncly-demo run`":            nil,
	} {
		if got := k.mentions(text); !slices.Equal(got, want) {
			t.Errorf("mentions(%q) = %q, want %q", text, got, want)
		}
	}
}

// TestFindReadsEverySwitch reads the switches as ncly resolves them. A flag
// may come before the domain, and an invalid value of one flag hides no
// later switch.
func TestFindReadsEverySwitch(t *testing.T) {
	k := demoKit(t)
	cases := []struct {
		words []string
		path  string
		json  bool
	}{
		{[]string{"demo", "run", "Ada", "--json"}, "demo run", true},
		{[]string{"--json", "demo", "run", "Ada"}, "demo run", true},
		{[]string{"demo", "run", "Ada", "--no-input=bad", "--json"}, "demo run", true},
		{[]string{"demo", "run", "--json", "--json=bad"}, "demo run", true},
		{[]string{"demo", "run", "--json", "--json=false"}, "demo run", false},
		{[]string{"--lang", "fr", "demo", "pick", "one", "--json"}, "demo pick one", true},
	}
	for _, c := range cases {
		call, found := k.find(c.words)
		if !found || call.path != c.path || call.switches["json"] != c.json {
			t.Errorf("find(%q) = %q with json %t (found %t), want %q with json %t", c.words, call.path, call.switches["json"], found, c.path, c.json)
		}
	}
	if _, found := k.find([]string{"doctor", "--json"}); found {
		t.Error("find took a call of core for one of the extension")
	}
}
