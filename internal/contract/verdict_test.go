package contract

import (
	"slices"
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
		{"Ctrl-C wins over item failures", Outcome{Results: []Item{item(), item(Runtime)}, Signal: Interrupt}, Verdict{130, Interrupted}},
		{"SIGTERM wins over command failures", Outcome{Errors: []Problem{fail(Runtime)}, Signal: Terminate}, Verdict{143, Terminated}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Decide(c.in); got != c.want {
				t.Errorf("Decide = %+v, want %+v", got, c.want)
			}
			reversed := c.in
			reversed.Errors = slices.Clone(c.in.Errors)
			slices.Reverse(reversed.Errors)
			reversed.Results = slices.Clone(c.in.Results)
			slices.Reverse(reversed.Results)
			if got := Decide(reversed); got != c.want {
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
			Finish(o, func(Code) Problem { return Problem{} })
		})
	}
}

func TestDetectMode(t *testing.T) {
	noEnv := func(string) (string, bool) { return "", false }
	emptyCI := func(name string) (string, bool) { return "", name == "CI" }
	cases := []struct {
		name                   string
		stdin, stdout, noInput bool
		env                    func(string) (string, bool)
		want                   Mode
	}{
		{"both streams are terminals", true, true, false, noEnv, Interactive},
		{"stdin is a pipe", false, true, false, noEnv, NonInteractive},
		{"stdout is a pipe", true, false, false, noEnv, NonInteractive},
		{"--no-input", true, true, true, noEnv, NonInteractive},
		{"CI is set, even empty", true, true, false, emptyCI, NonInteractive},
	}
	for _, c := range cases {
		if got := DetectMode(c.stdin, c.stdout, c.noInput, c.env); got != c.want {
			t.Errorf("%s: DetectMode = %v, want %v", c.name, got, c.want)
		}
	}
}
