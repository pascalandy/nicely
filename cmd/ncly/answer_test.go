package main

import (
	"strings"
	"testing"
)

// TestCheckAnswer proves what the answer helper rejects, which no scenario
// can prove, because answer takes no negation. A failed report, as doctor
// gives it with exit 78, passes answer stdout report only on one line, with
// the integer contract_version 1, ok false, and no errors key.
func TestCheckAnswer(t *testing.T) {
	const report = `{"ok":false,"contract_version":1,"checks":[{"id":"core.config","component":"core","status":"fail","code":"CONFIG_INVALID","message":"m","hint":"h"}]}`
	edit := func(old, replacement string) string { return strings.Replace(report, old, replacement, 1) + "\n" }
	cases := []struct {
		name, text, want string
		// failure is a part of the error, or empty when the answer passes.
		failure string
	}{
		{"a failed report", report + "\n", "report", ""},
		{"a failed report with a warning", edit(`"checks"`, `"warnings":[{"code":"KEYRING_ENV_ONLY","message":"m","hint":"h"}],"checks"`), "report", ""},
		{"two lines", report + "\n" + report + "\n", "report", "exactly one line"},
		{"no final newline", report, "report", "exactly one line"},
		{"a contract_version of 2", edit(`"contract_version":1`, `"contract_version":2`), "report", "contract_version"},
		{"a contract_version that is a string", edit(`"contract_version":1`, `"contract_version":"1"`), "report", "contract_version"},
		{"ok true", edit(`"ok":false`, `"ok":true`), "report", "ok is"},
		{"an empty errors", edit(`"checks"`, `"errors":[],"checks"`), "report", "a report holds errors"},
		{"an error", edit(`"checks"`, `"errors":[{"code":"RUNTIME","message":"m","hint":"h"}],"checks"`), "report", "a report holds errors"},
		{"a broken warning", edit(`"checks"`, `"warnings":[],"checks"`), "report", "warnings"},
		{"a failed report is no success", report + "\n", "ok", "ok is"},
		{"a failed report leads no code", report + "\n", "CONFIG_INVALID", "errors"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkAnswer(c.text, c.want)
			switch {
			case c.failure == "" && err != nil:
				t.Errorf("answer %s rejects %q: %v", c.want, c.text, err)
			case c.failure != "" && err == nil:
				t.Errorf("answer %s accepts %q", c.want, c.text)
			case c.failure != "" && !strings.Contains(err.Error(), c.failure):
				t.Errorf("answer %s rejects %q with %q, want %q", c.want, c.text, err, c.failure)
			}
		})
	}
}
