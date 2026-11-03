package verification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestChangedJarBytesNeverProveAModel serves other bytes at the pinned URL,
// as the URL of TLC v1.8.0 did on 2026-11-02. It also gives these bytes as a
// local jar. Neither starts Java, and the record stays as it was.
func TestChangedJarBytesNeverProveAModel(t *testing.T) {
	changed := []byte("the jar that the URL serves today")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(changed) }))
	defer server.Close()
	for _, local := range []bool{false, true} {
		r := newRepo(t)
		r.suite(t, server.URL, digest(r.jar))
		write(t, filepath.Join(r.root, RecordFile), "the record of an earlier run\n", 0o644)
		jar := ""
		if local {
			jar = r.localJar(t, changed)
		}
		_, err := Run(r.options(jar))
		if err == nil || !strings.Contains(err.Error(), "SHA-256 "+digest(changed)) {
			t.Fatalf("local=%v: Run accepted a jar with other bytes, or gave no digest: %v", local, err)
		}
		if r.javaStarted(t) {
			t.Fatalf("local=%v: Java started before the digest check", local)
		}
		if got := r.record(t); got != "the record of an earlier run\n" {
			t.Fatalf("local=%v: the record changed to %q", local, got)
		}
	}
}

// TestThePinnedJarRunsEachCheck is the control of the test above: the same
// suite with the pinned bytes starts Java, and writes the record.
func TestThePinnedJarRunsEachCheck(t *testing.T) {
	r := newRepo(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(r.jar) }))
	defer server.Close()
	r.suite(t, server.URL, digest(r.jar))
	for _, jar := range []string{"", r.localJar(t, r.jar)} {
		if err := os.RemoveAll(r.log); err != nil {
			t.Fatal(err)
		}
		results, err := Run(r.options(jar))
		if err != nil {
			t.Fatalf("jar=%q: %v", jar, err)
		}
		want := []Result{{Name: "Pass", Exit: 0, Distinct: 467328, Depth: 13}, {Name: "Bad", Exit: 12, Violation: "NoOpenPass"}}
		if !slices.Equal(results, want) {
			t.Fatalf("jar=%q: results %+v, want %+v", jar, results, want)
		}
		if !r.javaStarted(t) {
			t.Fatalf("jar=%q: Java never started", jar)
		}
		if problems := CheckRecord(r.root); len(problems) != 0 {
			t.Fatalf("jar=%q: the new record fails: %q", jar, problems)
		}
	}
}

// TestOnlyExpectedOutcomesWriteTheRecord gives each check an outcome that
// its suite does not expect. No run writes a record.
func TestOnlyExpectedOutcomesWriteTheRecord(t *testing.T) {
	for _, tc := range []struct {
		name, check, output string
		code                int
		want                string
	}{
		{"a pass where a violation is expected", "Bad", "pass.txt", 0, "TLC Bad: exit 0, want 12"},
		{"a violation where a pass is expected", "Pass", "violation.txt", 12, "TLC Pass: exit 12, want 0"},
		{"a crash", "Pass", "pass.txt", 1, "TLC Pass: exit 1, want 0"},
		{"another invariant", "Bad", "other-invariant.txt", 12, `"Invariant NoOpenPass is violated."`},
		{"no statistics", "Pass", "truncated.txt", 0, "statistics of the graph"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newRepo(t)
			r.tlcAnswers(t, tc.check, tc.output, tc.code)
			_, err := Run(r.options(r.localJar(t, r.jar)))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Run gave %v, want an error with %q", err, tc.want)
			}
			if got := r.record(t); got != "" {
				t.Fatalf("Run wrote a record:\n%s", got)
			}
		})
	}
}

// TestCheckReadsOnlyTheRecord changes each input of a fresh record. The
// check needs no Java: its environment holds no PATH.
func TestCheckReadsOnlyTheRecord(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(t *testing.T, r repo)
		want string
	}{
		{"a fresh record", func(*testing.T, repo) {}, ""},
		{"a changed model", func(t *testing.T, r repo) {
			write(t, filepath.Join(r.root, "verification/m/Model.tla"), "---- MODULE Model ----\nX == 1\n====\n", 0o644)
		}, "verification/m/Model.tla changed"},
		{"a new module beside the model", func(t *testing.T, r repo) {
			write(t, filepath.Join(r.root, "verification/m/Helper.tla"), "---- MODULE Helper ----\n====\n", 0o644)
		}, "verification/m/Helper.tla is new"},
		{"a new pin", func(t *testing.T, r repo) {
			r.suite(t, "https://example.com/tla2tools.jar", digest([]byte("another jar")))
		}, "pins TLC " + digest([]byte("another jar"))},
		{"a forged outcome", func(t *testing.T, r repo) { r.editRecord(t, `"exit": 12`, `"exit": 0`) }, "the check Bad gives exit 0"},
		{"a missing record", func(t *testing.T, r repo) {
			if err := os.Remove(filepath.Join(r.root, RecordFile)); err != nil {
				t.Fatal(err)
			}
		}, "no record"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newRepo(t)
			if _, err := Run(r.options(r.localJar(t, r.jar))); err != nil {
				t.Fatal(err)
			}
			tc.edit(t, r)
			problems := CheckRecord(r.root)
			if tc.want == "" && len(problems) != 0 || tc.want != "" && !slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, tc.want) }) {
				t.Fatalf("CheckRecord gave %q, want a problem with %q", problems, tc.want)
			}
		})
	}
	if problems := CheckRecord(t.TempDir()); problems != nil {
		t.Fatalf("a repository without %s has the problems %q", SuiteFile, problems)
	}
}

// TestCoreModels reads the record of core, as just check does. Its first
// model is the finish gate. At the parent of 59d33f38, a finding of this
// round with no state passed, and that gate violates NoOpenPass.
func TestCoreModels(t *testing.T) {
	root := filepath.Join("..", "..")
	if problems := CheckRecord(root); len(problems) != 0 {
		t.Fatalf("the record of core fails:\n%s", strings.Join(problems, "\n"))
	}
	var record Record
	if err := json.Unmarshal([]byte(readFile(t, filepath.Join(root, RecordFile))), &record); err != nil {
		t.Fatal(err)
	}
	want := map[string]Result{"Current": {Name: "Current"}, "PreFix": {Name: "PreFix", Exit: 12, Violation: "NoOpenPass"}}
	for _, r := range record.Results {
		w, ok := want[r.Name]
		if !ok || r.Exit != w.Exit || r.Violation != w.Violation {
			t.Fatalf("the record of core holds %+v, want %+v", r, w)
		}
		delete(want, r.Name)
	}
	if len(want) != 0 {
		t.Fatalf("the record of core lacks %v", want)
	}
}
