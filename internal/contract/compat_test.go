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
	a := Finish(o, summary)
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
		a := Finish(Outcome{Data: map[string]int{key: 1}}, summary)
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
		if err := Finish(c.in, summary).WriteJSON(&stdout, &stderr); err != nil {
			t.Fatal(err)
		}
		if stdout.String() != c.out || stderr.String() != "diagnostic line\n"+c.error {
			t.Errorf("%s: stdout %q, stderr %q", c.name, stdout.String(), stderr.String())
		}
	}
}
