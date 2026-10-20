package extensiontest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// TestScenarioLinksReadExecAsTestscript reads each exec line word by word,
// as testscript runs it. Quotes may wrap the program, and a quoted & still
// sends the program to the background.
func TestScenarioLinksReadExecAsTestscript(t *testing.T) {
	k := demoKit(t)
	k.root = t.TempDir()
	lines := []struct {
		line    string
		refused bool
	}{
		{"exec ncly demo run folder", true},
		{"[linux] ! exec ncly-demo run", true},
		{"exec 'bin dir/ncly' --version", true},
		{"exec bin/'ncly-demo' --version", true},
		{"exec 'it''s/ncly' --version", true},
		{"exec ncly --version # not in the background &", true},
		{"exec ncly --version & # in the background", false},
		{"exec ncly --version '&'", false},
		{"exec ncly --version &named&", false},
		{"exec 'ncly #1' --version", false},
		{"ncly demo run folder", false},
		{"stdout 'exec ncly'", false},
	}
	var script strings.Builder
	var want []string
	for i, l := range lines {
		script.WriteString(l.line + "\n")
		if l.refused {
			want = append(want, fmt.Sprintf("exec.txtar:%d:", i+1))
		}
	}
	dir := filepath.Join(k.root, "testdata", "script")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "exec.txtar"), []byte(script.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, problem := range k.scenarioLinks() {
		if strings.Contains(problem, "exec runs") {
			got = append(got, strings.TrimPrefix(problem[:strings.Index(problem, " ")], filepath.Join("testdata", "script")+string(filepath.Separator)))
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("scenarioLinks refused %q, want %q", got, want)
	}
}

// TestFindReadsEverySwitch reads the switches as ncly resolves them. A flag
// may come before the domain, and an invalid value or a malformed flag
// hides no later switch.
func TestFindReadsEverySwitch(t *testing.T) {
	k := demoKit(t)
	cases := []struct {
		words      []string
		path       string
		json, help bool
	}{
		{[]string{"demo", "run", "Ada", "--json"}, "demo run", true, false},
		{[]string{"--json", "demo", "run", "Ada"}, "demo run", true, false},
		{[]string{"demo", "run", "Ada", "--no-input=bad", "--json"}, "demo run", true, false},
		{[]string{"demo", "run", "--json", "--json=bad"}, "demo run", true, false},
		{[]string{"demo", "run", "--json", "--json=false"}, "demo run", false, false},
		{[]string{"demo", "run", "--json", "--=bad", "--help"}, "demo run", true, true},
		{[]string{"demo", "run", "--=bad", "-h"}, "demo run", false, true},
		{[]string{"--lang", "fr", "demo", "pick", "one", "--json"}, "demo pick one", true, false},
	}
	for _, c := range cases {
		call, found := k.find(c.words)
		json, _ := call.line.Switch("json")
		help, _ := call.line.Switch("help")
		if !found || call.path != c.path || json != c.json || help != c.help {
			t.Errorf("find(%q) = %q with json %t and help %t (found %t), want %q with json %t and help %t", c.words, call.path, json, help, found, c.path, c.json, c.help)
		}
	}
	if _, found := k.find([]string{"doctor", "--json"}); found {
		t.Error("find took a call of core for one of the extension")
	}
}
