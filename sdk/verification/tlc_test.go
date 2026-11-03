package verification

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestTLCOutput reads excerpts of TLC v1.7.4. A check passes only with its
// expected exit, its message, and the statistics of the graph. A pass also
// leaves no state on the queue.
func TestTLCOutput(t *testing.T) {
	pass, violation := readTestdata(t, "pass.txt"), readTestdata(t, "violation.txt")
	for _, tc := range []struct {
		name, output, violation string
		code                    int
		valid                   bool
	}{
		{"complete", pass, "", 0, true},
		{"counterexample", violation, "NoOpenPass", 12, true},
		{"wrong exit", violation, "NoOpenPass", 0, false},
		{"wrong invariant", violation, "NoWorkPass", 12, false},
		{"truncated", "Model checking completed. No error has been found.\n", "", 0, false},
		{"progress only", strings.ReplaceAll(pass, "7566145 states generated, 467328 distinct states found, 0 states left on queue.\n", ""), "", 0, false},
		{"states left", strings.ReplaceAll(pass, "found, 0 states left", "found, 3 states left"), "", 0, false},
		{"crash", pass, "", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := parseResult(Check{Name: "c", Violation: tc.violation}, tc.code, tc.output)
			if (err == nil) != tc.valid {
				t.Fatalf("result %+v, error %v", r, err)
			}
			// The count of the last line, not that of a progress line.
			if tc.name == "complete" && (r.Distinct != 467328 || r.Depth != 13) {
				t.Fatalf("result %+v, want 467328 distinct states at depth 13", r)
			}
		})
	}
}

// TestSuiteRefusals refuses a suite that could read outside the repository
// or that names one check twice.
func TestSuiteRefusals(t *testing.T) {
	for _, tc := range []struct{ name, old, replacement, want string }{
		{"a path outside", `"model": "verification/m/Model.tla"`, `"model": "../Model.tla"`, "../Model.tla"},
		{"a name twice", `"name": "Bad"`, `"name": "Pass"`, "Pass twice"},
		{"no digest", `"tlc_sha256": "`, `"tlc_sha256": "x`, "tlc_sha256"},
		{"an unknown key", `"java":`, `"workers": 4, "java":`, "workers"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newRepo(t)
			path := filepath.Join(r.root, SuiteFile)
			text := readFile(t, path)
			write(t, path, strings.Replace(text, tc.old, tc.replacement, 1), 0o644)
			_, err := Run(r.options(r.localJar(t, r.jar)))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Run gave %v, want an error with %q", err, tc.want)
			}
			if r.javaStarted(t) {
				t.Fatal("Java started for a refused suite")
			}
		})
	}
}
