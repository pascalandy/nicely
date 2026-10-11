package contract

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

func fail(code Code) Problem { return Problem{Code: code, Message: "m", Hint: "h"} }

func item(codes ...Code) Item {
	it := Item{Data: map[string]string{"url": "u"}}
	for _, c := range codes {
		it.Errors = append(it.Errors, fail(c))
	}
	return it
}

func TestDecide(t *testing.T) {
	cases := []struct {
		name string
		in   Outcome
		want Verdict
	}{
		{"success", Outcome{}, Verdict{0, ""}},
		{"usage error", Outcome{Errors: []Problem{fail(UsageInvalid)}}, Verdict{2, UsageInvalid}},
		{"a human must act before the call is fixed", Outcome{Errors: []Problem{fail(UsageInvalid), fail(ConfigInvalid)}}, Verdict{78, ConfigInvalid}},
		{"runtime leads a temporary failure", Outcome{Errors: []Problem{fail(Temporary), fail(Runtime)}}, Verdict{1, Runtime}},
		{"temporary before any effect", Outcome{Errors: []Problem{fail(Temporary)}}, Verdict{75, Temporary}},
		{"temporary after a proven idempotent write", Outcome{Errors: []Problem{fail(Temporary)}, Effect: RepeatableEffect}, Verdict{75, Temporary}},
		{"temporary after a non-repeatable write", Outcome{Errors: []Problem{fail(Temporary)}, Effect: NonRepeatableEffect}, Verdict{1, Runtime}},
		{"temporary after an unknown effect", Outcome{Errors: []Problem{fail(Temporary)}, Effect: UnknownEffect}, Verdict{1, Runtime}},
		{"temporary after a paid request started", Outcome{Errors: []Problem{fail(Temporary)}, Effect: PaidEffect}, Verdict{1, Runtime}},
		{"batch that succeeded", Outcome{Results: []Item{item(), item()}}, Verdict{0, ""}},
		{"batch with a temporary item", Outcome{Results: []Item{item(), item(Temporary)}}, Verdict{75, Temporary}},
		{"batch whose finished item wrote a file", Outcome{Results: []Item{item(), item(Temporary)}, Effect: NonRepeatableEffect}, Verdict{1, Runtime}},
		{"batch with a runtime item", Outcome{Results: []Item{item(Temporary), item(Runtime)}}, Verdict{1, Runtime}},
		{"batch with an item that needs a human", Outcome{Results: []Item{item(), item(TerminalRequired)}}, Verdict{1, Runtime}},
		{"a failed item makes the verdict the batch's, before a command error", Outcome{Results: []Item{item(), item(Runtime)}, Errors: []Problem{fail(ConfigInvalid)}}, Verdict{1, Runtime}},
		{"a signal's error from a nested command wins over item failures", Outcome{Results: []Item{item(), item(Runtime)}, Errors: []Problem{fail(Terminated)}}, Verdict{143, Terminated}},
		{"a temporary batch with a temporary command error", Outcome{Results: []Item{item(), item(Temporary)}, Errors: []Problem{fail(Temporary)}}, Verdict{75, Temporary}},
		{"Ctrl-C wins over item failures", Outcome{Results: []Item{item(), item(Runtime)}, Signal: Interrupt}, Verdict{130, Interrupted}},
		{"SIGTERM wins over command failures", Outcome{Errors: []Problem{fail(Runtime)}, Signal: Terminate}, Verdict{143, Terminated}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Core.Decide(c.in); got != c.want {
				t.Errorf("Decide = %+v, want %+v", got, c.want)
			}
			reversed := c.in
			reversed.Errors = slices.Clone(c.in.Errors)
			slices.Reverse(reversed.Errors)
			reversed.Results = slices.Clone(c.in.Results)
			slices.Reverse(reversed.Results)
			if got := Core.Decide(reversed); got != c.want {
				t.Errorf("Decide with failures in reverse order = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestFinishPanicsOnAnUnregisteredCode(t *testing.T) {
	cases := map[string]Outcome{
		"error":            {Errors: []Problem{fail("NOT_IN_THE_SPEC")}},
		"warning as error": {Errors: []Problem{fail(ConfigUnknownKey)}},
		"error as warning": {Warnings: []Problem{fail(Runtime)}},
		"item error":       {Results: []Item{item("NOT_IN_THE_SPEC")}},
	}
	for name, o := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Finish accepted a code that contract.md does not register")
				}
			}()
			Core.Finish(o, func(Code) Problem { return Problem{} })
		})
	}
}

func TestDetectMode(t *testing.T) {
	noEnv := func(string) (string, bool) { return "", false }
	emptyCI := func(name string) (string, bool) { return "", name == "CI" }
	jsonVariable := func(name string) (string, bool) { return "1", name == "NCLY_JSON" }
	cases := []struct {
		name                         string
		stdin, stdout, noInput, json bool
		env                          func(string) (string, bool)
		want                         Mode
	}{
		{"both streams are terminals", true, true, false, false, noEnv, Interactive},
		{"stdin is a pipe", false, true, false, false, noEnv, NonInteractive},
		{"stdout is a pipe", true, false, false, false, noEnv, NonInteractive},
		{"--no-input", true, true, true, false, noEnv, NonInteractive},
		{"--json never asks", true, true, false, true, noEnv, NonInteractive},
		{"--json=false wins over NCLY_JSON=1", true, true, false, false, jsonVariable, Interactive},
		{"CI is set, even empty", true, true, false, false, emptyCI, NonInteractive},
	}
	for _, c := range cases {
		if got := DetectMode(c.stdin, c.stdout, c.noInput, c.json, c.env); got != c.want {
			t.Errorf("%s: DetectMode = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestAnExtensionsOwnCodesCompeteInTheOrderOfOutput(t *testing.T) {
	m := Manifest{Domain: "hello", Errors: []OwnCode{{"HELLO_GRUMPY", ExitRuntime}, {"HELLO_LOCKED", ExitHuman}, {"HELLO_ODD", ExitUsage}}}
	r := m.Registry()
	cases := []struct {
		name  string
		codes []Code
		want  Verdict
	}{
		{"an own code of exit 1 leads RUNTIME", []Code{Runtime, "HELLO_GRUMPY"}, Verdict{1, "HELLO_GRUMPY"}},
		{"core's codes of exit 78 lead an own one", []Code{"HELLO_LOCKED", TerminalRequired}, Verdict{78, TerminalRequired}},
		{"core's codes of exit 2 lead an own one", []Code{"HELLO_ODD", UsageInvalid}, Verdict{2, UsageInvalid}},
		{"exit 1 leads exit 78", []Code{"HELLO_LOCKED", "HELLO_GRUMPY"}, Verdict{1, "HELLO_GRUMPY"}},
	}
	for _, c := range cases {
		var o Outcome
		for _, code := range c.codes {
			o.Errors = append(o.Errors, fail(code))
		}
		if got := r.Decide(o); got != c.want {
			t.Errorf("%s: Decide = %+v, want %+v", c.name, got, c.want)
		}
	}
	a := r.Finish(Outcome{Warnings: []Problem{fail("HELLO_UNNAMED")}}, summary)
	if !a.OK() {
		t.Errorf("a warning of the extension changed the verdict to %+v", a.Verdict)
	}
}

func TestAReportThatNeedsAHumanAnswersOnStdout(t *testing.T) {
	a := Core.Finish(Outcome{Data: map[string]any{"checks": []string{"fail"}}, HumanNeeded: true}, summary)
	if a.Exit != ExitHuman || a.OK() {
		t.Fatalf("verdict %+v, want exit 78 and ok false", a.Verdict)
	}
	var stdout, stderr strings.Builder
	if err := a.WriteJSON(&stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if want := "{\"ok\":false,\"contract_version\":1,\"checks\":[\"fail\"]}\n"; stdout.String() != want || stderr.Len() > 0 {
		t.Errorf("stdout %q, stderr %q, want stdout %q", stdout.String(), stderr.String(), want)
	}
}

// TestCoreAcceptsTheSDKsAnswerOfAFailedBatch renders a batch whose failed
// item comes with an error of the command itself, and checks the answer as
// core does.
func TestCoreAcceptsTheSDKsAnswerOfAFailedBatch(t *testing.T) {
	m := Manifest{Domain: "hello", Errors: []OwnCode{{"HELLO_GRUMPY", ExitRuntime}}}
	for _, code := range []Code{"HELLO_GRUMPY", ConfigInvalid, Interrupted, Terminated} {
		o := Outcome{Results: []Item{item(), item(Runtime)}, Errors: []Problem{fail(code)}}
		a := m.Registry().Finish(o, summary)
		var stdout, stderr bytes.Buffer
		if err := a.WriteJSON(&stdout, &stderr); err != nil {
			t.Fatal(err)
		}
		reply := Reply{Exit: int(a.Exit), Stdout: stdout.Bytes(), Last: stderr.Bytes()}
		if _, err := CheckAnswer(m, ResultBatch, reply); err != nil {
			t.Errorf("core rejects the answer %s: %v", stderr.Bytes(), err)
		}
	}
}
