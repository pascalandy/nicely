package contract

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// These cases read rendered answers the way a consumer of contract version 1
// does. Each expected value is written out here, not computed by the code
// under test.

func summary(code Code) Problem { return Problem{Code: code, Message: "summary", Hint: "next"} }

func render(t *testing.T, o Outcome) (Answer, string) {
	t.Helper()
	a := Core.Finish(o, summary)
	line, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return a, string(line)
}

type outputDir struct {
	OutputDir string `json:"output_dir"`
}

func TestAnswersAsAConsumerReadsThem(t *testing.T) {
	cases := []struct {
		name string
		in   Outcome
		want string
	}{
		{
			"a success holds ok, the version, and its keys, without errors",
			Outcome{Data: outputDir{"/Users/me/out"}},
			`{"ok":true,"contract_version":1,"output_dir":"/Users/me/out"}`,
		},
		{
			"a failure keeps the keys that still apply",
			Outcome{Data: outputDir{"/Users/me/out"}, Errors: []Problem{fail(Runtime)}},
			`{"ok":false,"contract_version":1,"output_dir":"/Users/me/out","errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`,
		},
		{
			"the error that matches the exit code leads, then the others in their order",
			Outcome{Errors: []Problem{fail(Temporary), fail(UsageInvalid), fail(Runtime)}},
			`{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":"h"},{"code":"TEMPORARY","message":"m","hint":"h"},{"code":"USAGE_INVALID","message":"m","hint":"h"}]}`,
		},
		{
			"a batch keeps its items in input order, and its successful data",
			Outcome{Results: []Item{
				{Data: map[string]string{"url": "a"}},
				{Data: map[string]string{"url": "b"}, Errors: []Problem{fail(Temporary)}},
				{Data: map[string]string{"url": "c"}},
			}},
			`{"ok":false,"contract_version":1,"results":[{"ok":true,"url":"a"},{"ok":false,"url":"b","errors":[{"code":"TEMPORARY","message":"m","hint":"h"}]},{"ok":true,"url":"c"}],"errors":[{"code":"TEMPORARY","message":"summary","hint":"next"}]}`,
		},
		{
			"an interruption keeps the data and only the attempted items",
			Outcome{Data: outputDir{"/Users/me/out"}, Results: []Item{{Data: map[string]string{"url": "a"}}}, Signal: Interrupt},
			`{"ok":false,"contract_version":1,"output_dir":"/Users/me/out","results":[{"ok":true,"url":"a"}],"errors":[{"code":"INTERRUPTED","message":"summary","hint":"next"}]}`,
		},
		{
			"a warning never turns a success into a failure",
			Outcome{Warnings: []Problem{fail(ConfigUnknownKey)}},
			`{"ok":true,"contract_version":1,"warnings":[{"code":"CONFIG_UNKNOWN_KEY","message":"m","hint":"h"}]}`,
		},
		{
			"null stays null, and an absent key stays absent",
			Outcome{Data: map[string]any{"run_id": nil}},
			`{"ok":true,"contract_version":1,"run_id":null}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, got := render(t, c.in); got != c.want {
				t.Errorf("got  %s\nwant %s", got, c.want)
			}
		})
	}
}

func TestTypesAreThoseAConsumerExpects(t *testing.T) {
	_, line := render(t, Outcome{Errors: []Problem{fail(UsageInvalid)}})
	dec := json.NewDecoder(strings.NewReader(line))
	dec.UseNumber()
	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		t.Fatal(err)
	}
	if ok, isBool := raw["ok"].(bool); !isBool || ok {
		t.Errorf("ok = %#v, want the JSON boolean false", raw["ok"])
	}
	if v, isNumber := raw["contract_version"].(json.Number); !isNumber || v.String() != "1" {
		t.Errorf("contract_version = %#v, want the JSON number 1", raw["contract_version"])
	}
	if _, isArray := raw["errors"].([]any); !isArray {
		t.Errorf("errors = %#v, want an array", raw["errors"])
	}
}

func TestAnOlderConsumerIgnoresNewOptionalKeys(t *testing.T) {
	// A consumer written before warnings and output_dir existed.
	var older struct {
		OK     bool `json:"ok"`
		Errors []struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	_, line := render(t, Outcome{
		Data:     outputDir{"/Users/me/out"},
		Errors:   []Problem{fail(TerminalRequired)},
		Warnings: []Problem{fail(ConfigUnknownKey)},
	})
	if err := json.Unmarshal([]byte(line), &older); err != nil {
		t.Fatal(err)
	}
	if older.OK || len(older.Errors) != 1 || older.Errors[0].Code != "TERMINAL_REQUIRED" {
		t.Errorf("older consumer read %+v from %s", older, line)
	}
}

func TestDataCannotTakeAnEnvelopeKey(t *testing.T) {
	for _, key := range []string{"ok", "contract_version", "results", "errors", "warnings"} {
		a := Core.Finish(Outcome{Data: map[string]int{key: 1}}, summary)
		if _, err := json.Marshal(a); err == nil {
			t.Errorf("data with the key %q rendered", key)
		}
	}
}

func TestWriteJSONPicksTheStream(t *testing.T) {
	cases := []struct {
		name       string
		in         Outcome
		out, error string
	}{
		{"a success goes to stdout", Outcome{}, "{\"ok\":true,\"contract_version\":1}\n", ""},
		{"a failure goes to stderr", Outcome{Errors: []Problem{fail(UsageInvalid)}}, "", "{\"ok\":false,\"contract_version\":1,\"errors\":[{\"code\":\"USAGE_INVALID\",\"message\":\"m\",\"hint\":\"h\"}]}\n"},
	}
	for _, c := range cases {
		var stdout, stderr bytes.Buffer
		stderr.WriteString("diagnostic line\n")
		if err := Core.Finish(c.in, summary).WriteJSON(&stdout, &stderr); err != nil {
			t.Fatal(err)
		}
		if stdout.String() != c.out || stderr.String() != "diagnostic line\n"+c.error {
			t.Errorf("%s: stdout %q, stderr %q", c.name, stdout.String(), stderr.String())
		}
	}
}

// TestAConsumerActsOnTheExitCodeOfACodeItDoesNotKnow passes answers through
// core's answer check unchanged, then reads them as a consumer of contract
// version 1 that knows no code beyond USAGE_INVALID, as Error codes in
// contract.md asks: the set of codes is open, and the exit code says what to
// do.
func TestAConsumerActsOnTheExitCodeOfACodeItDoesNotKnow(t *testing.T) {
	act := func(exit int, line string) string {
		var answer struct {
			Errors []struct{ Code, Hint string } `json:"errors"`
		}
		if err := json.Unmarshal([]byte(line), &answer); err != nil {
			t.Fatal(err)
		}
		lead := answer.Errors[0]
		switch {
		case lead.Code == "USAGE_INVALID", exit == 2:
			return "fix the call"
		case exit == 75:
			return "repeat"
		case exit == 78:
			return "relay " + lead.Hint
		}
		return "inspect, never repeat"
	}
	cases := []struct {
		exit       int
		line, want string
	}{
		{1, `{"ok":false,"contract_version":1,"errors":[{"code":"TRANSCRIPT_DOWNLOAD_FAILED","message":"m","hint":"ncly run view r1"}]}`, "inspect, never repeat"},
		{78, `{"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"m","hint":"ncly auth login deepgram"}]}`, "relay ncly auth login deepgram"},
		{2, `{"ok":false,"contract_version":1,"errors":[{"code":"NOT_FOUND","message":"m","hint":"ncly skill list"}]}`, "fix the call"},
		{75, `{"ok":false,"contract_version":1,"errors":[{"code":"TEMPORARY","message":"m","hint":"h"}]}`, "repeat"},
	}
	for _, c := range cases {
		a, err := CheckAnswer(Manifest{Domain: "hello"}, ResultObject, Reply{Exit: c.exit, Last: []byte(c.line)})
		if err != nil {
			t.Fatalf("exit %d: core rejected the answer: %v", c.exit, err)
		}
		if string(a.Line) != c.line {
			t.Errorf("exit %d: core changed the answer to %s", c.exit, a.Line)
		}
		if got := act(c.exit, string(a.Line)); got != c.want {
			t.Errorf("exit %d: the consumer would %q, want %q", c.exit, got, c.want)
		}
	}
}

// TestAConsumerRunsNextOnlyWhenPresent reads a problem with next and one
// without it, as a consumer of contract version 1 does: it runs next as is
// when the key is present, and relays the hint otherwise. A consumer
// written before next existed reads the same answers.
func TestAConsumerRunsNextOnlyWhenPresent(t *testing.T) {
	cases := []struct {
		name string
		in   Problem
		line string
		want []string
	}{
		{
			"next names the command of the hint as words",
			Problem{Code: NotFound, Message: "m", Hint: "ncly describe", Next: []string{"ncly", "describe"}},
			`{"ok":false,"contract_version":1,"errors":[{"code":"NOT_FOUND","message":"m","hint":"ncly describe","next":["ncly","describe"]}]}`,
			[]string{"ncly", "describe"},
		},
		{
			"a problem without next leaves the key out",
			Problem{Code: UsageInvalid, Message: "m", Hint: "ncly completion --help"},
			`{"ok":false,"contract_version":1,"errors":[{"code":"USAGE_INVALID","message":"m","hint":"ncly completion --help"}]}`,
			nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, line := render(t, Outcome{Errors: []Problem{c.in}})
			if line != c.line {
				t.Fatalf("got  %s\nwant %s", line, c.line)
			}
			var answer struct {
				Errors []map[string]json.RawMessage `json:"errors"`
			}
			if err := json.Unmarshal([]byte(line), &answer); err != nil {
				t.Fatal(err)
			}
			raw, present := answer.Errors[0]["next"]
			if present != (c.want != nil) {
				t.Fatalf("next present = %t, want %t", present, c.want != nil)
			}
			if present {
				var next []string
				if err := json.Unmarshal(raw, &next); err != nil || strings.Join(next, " ") != strings.Join(c.want, " ") {
					t.Errorf("the consumer would run %s, want %q", raw, c.want)
				}
			}
			var older struct {
				Errors []struct{ Code, Hint string } `json:"errors"`
			}
			if err := json.Unmarshal([]byte(line), &older); err != nil || older.Errors[0].Code != string(c.in.Code) || older.Errors[0].Hint != c.in.Hint {
				t.Errorf("older consumer read %+v from %s", older, line)
			}
		})
	}
}

func TestAConsumerReadsAnArgumentWithAndWithoutValuesFrom(t *testing.T) {
	// Arguments as ncly describe gives them, written out: values_from is
	// present only when the declaration has one.
	with := `{"name":"name","summary":"s","type":"string","required":true,"variadic":false,"values":[],"values_from":{"command":"skill list","key":"skills","field":"name"}}`
	without := `{"name":"name","summary":"s","type":"string","required":true,"variadic":false,"values":[]}`
	type valuesFrom struct{ Command, Key, Field string }
	for _, c := range []struct {
		line string
		want *valuesFrom
	}{{with, &valuesFrom{"skill list", "skills", "name"}}, {without, nil}} {
		var arg struct {
			Name       string      `json:"name"`
			Required   bool        `json:"required"`
			ValuesFrom *valuesFrom `json:"values_from"`
		}
		if err := json.Unmarshal([]byte(c.line), &arg); err != nil {
			t.Fatal(err)
		}
		if arg.Name != "name" || !arg.Required || (arg.ValuesFrom == nil) != (c.want == nil) || c.want != nil && *arg.ValuesFrom != *c.want {
			t.Errorf("consumer read %+v from %s", arg, c.line)
		}
		// A consumer written before values_from reads the same argument.
		var older struct {
			Name     string `json:"name"`
			Required bool   `json:"required"`
			Values   []any  `json:"values"`
		}
		if err := json.Unmarshal([]byte(c.line), &older); err != nil || older.Name != "name" || !older.Required || len(older.Values) != 0 {
			t.Errorf("older consumer read %+v from %s", older, c.line)
		}
	}

	// The manifest key that declares it, and an argument without one.
	manifest := `format_version = 1
contract_version = 1
domain = "demo"

[[commands]]
path = "demo"
summary = "demo.summary"

[[commands]]
path = "demo view"
summary = "demo.summary"

[[commands.args]]
name = "name"
summary = "demo.summary"
values_from = { command = "demo item list", key = "items", field = "name" }

[[commands.args]]
name = "other"
summary = "demo.summary"

[[commands]]
path = "demo item"
summary = "demo.summary"

[[commands]]
path = "demo item list"
summary = "demo.summary"
effects = {}
`
	m, err := ParseManifest([]byte(manifest), "demo")
	if err != nil {
		t.Fatal(err)
	}
	view, _ := m.Command("demo view")
	got := view.Args[0].ValuesFrom
	if got == nil || strings.Join(got.Command, " ") != "demo item list" || got.Key != "items" || got.Field != "name" {
		t.Errorf("values_from of name is %+v", got)
	}
	if view.Args[1].ValuesFrom != nil {
		t.Errorf("an argument without values_from has %+v", view.Args[1].ValuesFrom)
	}
}

func TestAConsumerReadsACheckWithAndWithoutTheKeysOfAnExtension(t *testing.T) {
	// Checks as ncly doctor gives them, written out: domain, source, and
	// path appear in an entry of the extensions component alone, path is
	// null for a bundled extension, and hidden_by only when another one wins.
	type hiddenBy struct {
		Source string  `json:"source"`
		Path   *string `json:"path"`
	}
	type check struct {
		ID       string    `json:"id"`
		Status   string    `json:"status"`
		Domain   string    `json:"domain"`
		Source   string    `json:"source"`
		Path     *string   `json:"path"`
		HiddenBy *hiddenBy `json:"hidden_by"`
	}
	winner := "/Users/me/.local/share/nicely/extensions/ncly-hello"
	for _, c := range []struct {
		line     string
		want     check
		hasPath  bool
		wantPath string
	}{
		{`{"id":"core.config","component":"core","status":"pass","message":"m"}`, check{ID: "core.config", Status: "pass"}, false, ""},
		{`{"id":"extensions.skill","component":"extensions","status":"pass","message":"m","domain":"skill","source":"bundled","path":null}`, check{ID: "extensions.skill", Status: "pass", Domain: "skill", Source: "bundled"}, true, ""},
		{`{"id":"extensions.hello","component":"extensions","status":"pass","message":"m","domain":"hello","source":"extension","path":"/Users/me/bin/ncly-hello","hidden_by":{"source":"extension","path":"` + winner + `"}}`, check{ID: "extensions.hello", Status: "pass", Domain: "hello", Source: "extension", HiddenBy: &hiddenBy{"extension", &winner}}, true, "/Users/me/bin/ncly-hello"},
	} {
		var got check
		var keys map[string]json.RawMessage
		if err := json.Unmarshal([]byte(c.line), &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(c.line), &keys); err != nil {
			t.Fatal(err)
		}
		_, hasPath := keys["path"]
		path := ""
		if got.Path != nil {
			path = *got.Path
		}
		sameHidden := (got.HiddenBy == nil) == (c.want.HiddenBy == nil) &&
			(got.HiddenBy == nil || got.HiddenBy.Source == c.want.HiddenBy.Source && *got.HiddenBy.Path == *c.want.HiddenBy.Path)
		if got.ID != c.want.ID || got.Status != c.want.Status || got.Domain != c.want.Domain || got.Source != c.want.Source ||
			hasPath != c.hasPath || path != c.wantPath || !sameHidden {
			t.Errorf("consumer read %+v from %s", got, c.line)
		}
		// A consumer written before the extensions component reads the same check.
		var older struct {
			ID        string `json:"id"`
			Component string `json:"component"`
			Status    string `json:"status"`
			Message   string `json:"message"`
		}
		if err := json.Unmarshal([]byte(c.line), &older); err != nil || older.ID != c.want.ID || older.Status != "pass" || older.Message != "m" {
			t.Errorf("older consumer read %+v from %s", older, c.line)
		}
	}
}

func TestAConsumerReadsAProgramWithAndWithoutItsEntry(t *testing.T) {
	// Entries of requires.bins as ncly describe gives them, written out:
	// conditional, version, probe, and install appear only when the
	// manifest declares them, and probe then defaults to --version.
	with := `{"name":"ffmpeg","conditional":true,"version":"6.0","probe":["-version"],"install":{"brew":["brew","install","ffmpeg"]}}`
	without := `{"name":"ffmpeg"}`
	for _, c := range []struct {
		line string
		want Requirement
	}{
		{with, Requirement{Conditional: true, Version: "6.0", Probe: []string{"-version"}, Install: map[string][]string{"brew": {"brew", "install", "ffmpeg"}}}},
		{without, Requirement{}},
	} {
		var bin struct {
			Name        string              `json:"name"`
			Conditional bool                `json:"conditional"`
			Version     string              `json:"version"`
			Probe       []string            `json:"probe"`
			Install     map[string][]string `json:"install"`
		}
		if err := json.Unmarshal([]byte(c.line), &bin); err != nil {
			t.Fatal(err)
		}
		got := Requirement{Conditional: bin.Conditional, Version: bin.Version, Probe: bin.Probe, Install: bin.Install}
		if bin.Name != "ffmpeg" || got.Conditional != c.want.Conditional || got.Version != c.want.Version ||
			strings.Join(got.ProbeArgs(), " ") != strings.Join(c.want.ProbeArgs(), " ") ||
			strings.Join(got.Install["brew"], " ") != strings.Join(c.want.Install["brew"], " ") {
			t.Errorf("consumer read %+v from %s", bin, c.line)
		}
		// A consumer written before M05 reads the name alone.
		var older struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(c.line), &older); err != nil || older.Name != "ffmpeg" {
			t.Errorf("older consumer read %+v from %s", older, c.line)
		}
	}
}

func TestAConsumerReadsAServiceWithAndWithoutGrants(t *testing.T) {
	// Entries of services as ncly auth status gives them, written out: an
	// entry of M06 holds no grants, and from M08 grants lists one object per
	// domain whose extension declares the service, with granted and
	// optional.
	type grant struct {
		Domain   string `json:"domain"`
		Granted  bool   `json:"granted"`
		Optional bool   `json:"optional"`
	}
	type service struct {
		Service    string  `json:"service"`
		Configured bool    `json:"configured"`
		Source     string  `json:"source"`
		Code       Code    `json:"code"`
		Grants     []grant `json:"grants"`
	}
	for _, c := range []struct {
		line      string
		want      service
		hasGrants bool
	}{
		{`{"service":"deepgram","configured":true,"source":"keychain"}`, service{Service: "deepgram", Configured: true, Source: "keychain"}, false},
		{`{"service":"deepgram","configured":true,"source":"keychain","grants":[{"domain":"transcript","granted":true,"optional":false}]}`, service{Service: "deepgram", Configured: true, Source: "keychain", Grants: []grant{{"transcript", true, false}}}, true},
		{`{"service":"deepgram","configured":true,"source":"env","code":"KEY_NOT_GRANTED","message":"m","hint":"ncly auth grant deepgram transcript","next":["ncly","auth","grant","deepgram","transcript"],"grants":[{"domain":"agent","granted":false,"optional":true},{"domain":"transcript","granted":false,"optional":false}]}`, service{Service: "deepgram", Configured: true, Source: "env", Code: KeyNotGranted, Grants: []grant{{"agent", false, true}, {"transcript", false, false}}}, true},
	} {
		var got service
		var keys map[string]json.RawMessage
		if err := json.Unmarshal([]byte(c.line), &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(c.line), &keys); err != nil {
			t.Fatal(err)
		}
		_, hasGrants := keys["grants"]
		sameGrants := len(got.Grants) == len(c.want.Grants)
		for i := range got.Grants {
			sameGrants = sameGrants && got.Grants[i] == c.want.Grants[i]
		}
		if got.Service != c.want.Service || got.Configured != c.want.Configured || got.Source != c.want.Source ||
			got.Code != c.want.Code || hasGrants != c.hasGrants || !sameGrants {
			t.Errorf("consumer read %+v from %s", got, c.line)
		}
		// KEY_NOT_GRANTED asks a human to act, as every code of exit 78 does.
		if exit, known := got.Code.Exit(); got.Code != "" && (!known || exit != ExitHuman) {
			t.Errorf("%s maps to exit %d, %v", got.Code, exit, known)
		}
		// A consumer written before M08 reads the same verdict, and ignores grants.
		var older struct {
			Service    string `json:"service"`
			Configured bool   `json:"configured"`
			Source     string `json:"source"`
		}
		if err := json.Unmarshal([]byte(c.line), &older); err != nil || older.Service != "deepgram" || older.Configured != c.want.Configured || older.Source != c.want.Source {
			t.Errorf("older consumer read %+v from %s", older, c.line)
		}
	}
}
