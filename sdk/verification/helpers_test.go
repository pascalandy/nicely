package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
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

func (r repo) editRecord(t *testing.T, old, replacement string) {
	t.Helper()
	text := r.record(t)
	if !strings.Contains(text, old) {
		t.Fatalf("the record holds no %q:\n%s", old, text)
	}
	write(t, filepath.Join(r.root, RecordFile), strings.Replace(text, old, replacement, 1), 0o644)
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
