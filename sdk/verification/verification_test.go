package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// fakeMise stands for mise and the TLC that it starts. It logs each call,
// so a test sees whether Java started. For a model check, it prints the
// output and exits with the code of the files named after the configuration.
const fakeMise = `#!/bin/sh
printf '%s\n' "$*" >>"$FAKE_MISE_LOG"
config=
previous=
for arg in "$@"; do
	[ "$previous" = -config ] && config=$arg
	previous=$arg
done
if [ -z "$config" ]; then
	echo 'openjdk version "21.0.13"'
	exit 0
fi
name=$(basename "$config" .cfg)
cat "$FAKE_TLC/$name.out"
exit "$(cat "$FAKE_TLC/$name.exit")"
`

// repo is a repository with a suite of two checks, Pass and Bad, and the
// programs and folders of one run.
type repo struct {
	root, dir, tlc, log string
	env                 []string
	jar                 []byte
}

func newRepo(t *testing.T) repo {
	t.Helper()
	base := t.TempDir()
	r := repo{root: filepath.Join(base, "repo"), dir: filepath.Join(base, "run"), tlc: filepath.Join(base, "tlc"), log: filepath.Join(base, "mise.log"), jar: []byte("the pinned jar")}
	bin := filepath.Join(base, "bin")
	for _, dir := range []string{bin, r.dir, r.tlc} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	write(t, filepath.Join(bin, "mise"), fakeMise, 0o755)
	r.env = []string{"PATH=" + bin + ":/usr/bin:/bin", "FAKE_MISE_LOG=" + r.log, "FAKE_TLC=" + r.tlc}
	write(t, filepath.Join(r.root, "verification/m/Model.tla"), "---- MODULE Model ----\n====\n", 0o644)
	write(t, filepath.Join(r.root, "verification/m/Pass.cfg"), "SPECIFICATION Spec\n", 0o644)
	write(t, filepath.Join(r.root, "verification/m/Bad.cfg"), "SPECIFICATION Spec\n", 0o644)
	r.suite(t, "https://example.com/tla2tools.jar", digest(r.jar))
	r.tlcAnswers(t, "Pass", "pass.txt", 0)
	r.tlcAnswers(t, "Bad", "violation.txt", 12)
	return r
}

// suite writes verification/tlc.json with the pins url and sha.
func (r repo) suite(t *testing.T, url, sha string) {
	t.Helper()
	data, err := json.MarshalIndent(Suite{
		Java: "java@temurin-21.0.13+9.0.LTS", URL: url, SHA256: sha,
		Checks: []Check{
			{Name: "Pass", Model: "verification/m/Model.tla", Config: "verification/m/Pass.cfg"},
			{Name: "Bad", Model: "verification/m/Model.tla", Config: "verification/m/Bad.cfg", Violation: "NoOpenPass"},
		},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(r.root, SuiteFile), string(data)+"\n", 0o644)
}

// tlcAnswers makes the fake TLC print testdata/<output> and exit with code
// for the configuration name.
func (r repo) tlcAnswers(t *testing.T, name, output string, code int) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", output))
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(r.tlc, name+".out"), string(data), 0o644)
	write(t, filepath.Join(r.tlc, name+".exit"), strconv.Itoa(code)+"\n", 0o644)
}

// localJar writes data as a local jar, and returns its path.
func (r repo) localJar(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(r.root), "local.jar")
	write(t, path, string(data), 0o600)
	return path
}

func (r repo) options(jar string) Options {
	return Options{Root: r.root, Dir: r.dir, Jar: jar, Env: r.env}
}

// javaStarted reports whether the fake mise ran.
func (r repo) javaStarted(t *testing.T) bool {
	t.Helper()
	_, err := os.Stat(r.log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return err == nil
}

func (r repo) record(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(r.root, RecordFile))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(data)
}

// TestChangedJarBytesNeverProveAModel serves other bytes at the pinned URL,
// as the release of TLC v1.8.0 did on 2026-11-02, and gives the same bytes
// as a local jar. Neither starts Java, and the record stays as it was.
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
		{"no statistics", "Pass", "truncated.txt", 0, "graph statistics"},
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
		{"a new pin", func(t *testing.T, r repo) { r.suite(t, "https://example.com/tla2tools.jar", digest([]byte("another jar"))) }, "pins TLC " + digest([]byte("another jar"))},
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

func (r repo) editRecord(t *testing.T, old, replacement string) {
	t.Helper()
	text := r.record(t)
	if !strings.Contains(text, old) {
		t.Fatalf("the record holds no %q:\n%s", old, text)
	}
	write(t, filepath.Join(r.root, RecordFile), strings.Replace(text, old, replacement, 1), 0o644)
}

// TestTLCOutput reads excerpts of TLC v1.7.4. A check passes only with its
// expected exit, its message, and the statistics of the graph.
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
		{"crash", pass, "", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := parseResult(Check{Name: "c", Violation: tc.violation}, tc.code, tc.output)
			if (err == nil) != tc.valid {
				t.Fatalf("result %+v, error %v", r, err)
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

// TestCoreModels reads the record of core, as just check does. Its first
// model is the finish gate, and the gate at the parent of 59d33f38, where a
// finding of this round with no state passed, violates NoOpenPass.
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

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func write(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	return readFile(t, filepath.Join("testdata", name))
}
