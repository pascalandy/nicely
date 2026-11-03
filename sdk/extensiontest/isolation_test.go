package extensiontest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// TestTheScenariosIgnoreTheCallersExtensions puts an extension in a folder
// of the caller's PATH, which discovery would find. A scenario that lists
// the extensions must see only the one under test.
func TestTheScenariosIgnoreTheCallersExtensions(t *testing.T) {
	outside := t.TempDir()
	for name, content := range map[string]string{
		"ncly-outsider":                 "#!/bin/sh\nexit 0\n",
		"ncly-outsider.toml":            "format_version = 1\ncontract_version = 1\ndomain = \"outsider\"\n\n[[commands]]\npath = \"outsider\"\nsummary = \"outsider.summary\"\n\n[[commands]]\npath = \"outsider run\"\nsummary = \"outsider.run.summary\"\n",
		"ncly-outsider.locales/en.toml": "[\"outsider.summary\"]\ndescription = \"d\"\nother = \"Outside the test\"\n\n[\"outsider.run.summary\"]\ndescription = \"d\"\nother = \"Run outside\"\n",
	} {
		path := filepath.Join(outside, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	scenario := edit{file: "testdata/script/outsider.txtar", new: "ncly --help\nstdout 'standalone'\n! stdout 'outsider'\n"}
	if out, err := goTest(t, "standalone", outside+string(filepath.ListSeparator)+os.Getenv("PATH"), scenario); err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}

// callerState holds settings of the caller that changed past checks. At the
// parent of 09809608, the lifecycle check passed os.Environ(), and an invalid
// NCLY_CONFIG reached ncly. Its scratch HOME did not hide it.
func callerState(t *testing.T) map[string]string {
	state := map[string]string{
		"NCLY_CONFIG":        filepath.Join(t.TempDir(), "invalid.toml"),
		"NCLY_LANG":          "xx-YY",
		"NCLY_JSON":          "1",
		"NCLY_NO_INPUT":      "1",
		"NCLY_TEST_SCENARIO": "refusal",
		"NCLY_TEST_NOTICES":  filepath.Join(t.TempDir(), "notices"),
		"GIT_DIR":            filepath.Join(t.TempDir(), "hook.git"),
		"JUST_UNSTABLE":      "1",
		"GOFLAGS":            "-overlay=" + filepath.Join(t.TempDir(), "overlay.json"),
		"GOWORK":             filepath.Join(t.TempDir(), "go.work"),
	}
	for name, value := range state {
		t.Setenv(name, value)
	}
	return state
}

// environment returns the variables of env by name, and fails on a name that
// env gives twice.
func environment(t *testing.T, env []string) map[string]string {
	t.Helper()
	vars := map[string]string{}
	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		if _, twice := vars[name]; twice {
			t.Fatalf("the run gives %s twice: %q", name, env)
		}
		vars[name] = value
	}
	return vars
}

// TestAHostRunDropsTheCallersState checks the run of signoff and of the Git
// and gh steps of a release. It keeps the tools and accounts of the caller,
// and only the settings that the profile names.
func TestAHostRunDropsTheCallersState(t *testing.T) {
	state := callerState(t)
	t.Setenv("CALLER_TOOL", "kept")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	run, err := extensionrepo.NewOwnedRun(t.TempDir(), extensionrepo.Profile{Keep: []string{"GIT_CONFIG_GLOBAL"}})
	if err != nil {
		t.Fatal(err)
	}
	vars := environment(t, run.Env)
	for name := range state {
		if value, ok := vars[name]; ok && name != "GOFLAGS" && name != "GOWORK" && name != "NCLY_TEST_NOTICES" {
			t.Errorf("the host run keeps %s=%s of the caller", name, value)
		}
	}
	want := map[string]string{
		"GOFLAGS":           " ",
		"GOWORK":            "off",
		"NCLY_TEST_NOTICES": filepath.Join(run.Dir, "notices"),
		"CALLER_TOOL":       "kept",
		"GIT_CONFIG_GLOBAL": os.Getenv("GIT_CONFIG_GLOBAL"),
		"HOME":              os.Getenv("HOME"),
	}
	for name, value := range want {
		if vars[name] != value {
			t.Errorf("the host run gives %s=%q, want %q", name, vars[name], value)
		}
	}
}

// TestAScratchRunKeepsNothingElse checks the run of the kit, of acceptance,
// and of the install check of a release. Replacing HOME alone is no
// isolation, so the run keeps only the variables that it names.
func TestAScratchRunKeepsNothingElse(t *testing.T) {
	state := callerState(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "ncly-outsider"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	caller := t.TempDir()
	t.Setenv("PATH", outside+string(filepath.ListSeparator)+caller)
	t.Setenv("CALLER_TOOL", "dropped")
	t.Setenv("GOCOVERDIR", filepath.Join(t.TempDir(), "cover"))
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	run, err := extensionrepo.NewOwnedRun(dir, extensionrepo.Profile{Scratch: true, Path: []string{bin}, Keep: []string{"GOCOVERDIR"}, Set: map[string]string{"LC_ALL": "C"}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"PATH":                     bin + string(filepath.ListSeparator) + caller,
		"HOME":                     dir,
		"TMPDIR":                   filepath.Join(dir, ".tmp"),
		"XDG_CONFIG_HOME":          filepath.Join(dir, ".config"),
		"XDG_DATA_HOME":            filepath.Join(dir, ".local", "share"),
		"XDG_STATE_HOME":           filepath.Join(dir, ".local", "state"),
		"XDG_CACHE_HOME":           filepath.Join(dir, ".cache"),
		"DBUS_SESSION_BUS_ADDRESS": "disabled:",
		"GOFLAGS":                  " ",
		"GOWORK":                   "off",
		"GOCOVERDIR":               os.Getenv("GOCOVERDIR"),
		"LC_ALL":                   "C",
	}
	vars := environment(t, run.Env)
	var kept []string
	for name := range vars {
		if _, ok := want[name]; !ok {
			kept = append(kept, name)
		}
	}
	// A failure names the variables only, because the caller can hold keys.
	if len(kept) != 0 {
		slices.Sort(kept)
		t.Errorf("the scratch run keeps these variables of the caller: %s", strings.Join(kept, " "))
	}
	for name, value := range want {
		if vars[name] != value {
			t.Errorf("the scratch run gives %s=%q, want %q", name, vars[name], value)
		}
	}
	for name := range state {
		if _, ok := want[name]; !ok && vars[name] != "" {
			t.Errorf("the scratch run keeps %s of the caller", name)
		}
	}
	if info, err := os.Stat(filepath.Join(dir, ".tmp")); err != nil || !info.IsDir() {
		t.Errorf("the scratch run has no TMPDIR folder: %v", err)
	}
}

// TestAHostRunCarriesTheNoticesOfItsPrograms runs a program that records a
// notice, as the kit does for a skipped example. The run returns it to the
// one result of the action. At the parent of e35bd0bb, check ran go test
// with no notice file, and the passing package hid the notice.
func TestAHostRunCarriesTheNoticesOfItsPrograms(t *testing.T) {
	callerState(t)
	run, err := extensionrepo.NewOwnedRun(t.TempDir(), extensionrepo.Profile{})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestNoticeChild$")
	cmd.Env = append(run.Env, "NCLY_NOTICE_CHILD=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the notice child failed: %v\n%s", err, out)
	}
	notices, err := run.Notices()
	if err != nil {
		t.Fatal(err)
	}
	if string(notices) != "skipped: the notice of the child\n" {
		t.Fatalf("the run carries the notices %q, want the one of the child", notices)
	}
}

// TestNoticeChild records one notice when TestAHostRunCarriesTheNoticesOfItsPrograms
// starts it.
func TestNoticeChild(t *testing.T) {
	if os.Getenv("NCLY_NOTICE_CHILD") == "" {
		t.Skip("runs only as the child of TestAHostRunCarriesTheNoticesOfItsPrograms")
	}
	if err := extensionrepo.Notice("skipped: the notice of the child"); err != nil {
		t.Fatal(err)
	}
}
