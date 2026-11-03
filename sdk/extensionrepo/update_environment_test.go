package extensionrepo

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A Copier subprocess uses Git too. Under a hook's environment, even its -C
// can address the caller's repository. Both repositories here belong to this
// test, and the control reader never inherits the poisoned environment.
func TestUpdateDropsCallerGitEnvironment(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is required")
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home, bin := filepath.Join(root, "home"), filepath.Join(root, "bin")
	for _, path := range []string{home, bin, filepath.Join(root, "hooks")} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	controlEnv := []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + home,
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "LC_ALL=C",
	}
	control := func(repo string, args ...string) []byte {
		t.Helper()
		cmd := exec.Command(git, append([]string{"-C", repo}, args...)...)
		cmd.Env = controlEnv
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("control git %v: %v\n%s", args, err, out)
		}
		return out
	}
	a, b := filepath.Join(root, "target"), filepath.Join(root, "other")
	for _, repo := range []string{a, b} {
		control(root, "init", "--quiet", repo)
		control(repo, "config", "core.hooksPath", filepath.Join(root, "hooks"))
		control(repo, "config", "user.name", "Update fixture")
		control(repo, "config", "user.email", "fixture@example.com")
		control(repo, "config", "commit.gpgsign", "false")
		updateFixtureWrite(t, filepath.Join(repo, "tracked.txt"), filepath.Base(repo)+"\n", 0o600)
		control(repo, "add", "tracked.txt")
		control(repo, "commit", "--quiet", "-m", "fixture")
	}
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	snapshot := func() map[string][]byte {
		return map[string][]byte{
			"refs": control(b, "for-each-ref"), "HEAD": read(filepath.Join(b, ".git", "HEAD")),
			"index": read(filepath.Join(b, ".git", "index")), "config": read(filepath.Join(b, ".git", "config")),
			"tracked": read(filepath.Join(b, "tracked.txt")),
		}
	}
	before := snapshot()
	targetHead := strings.TrimSpace(string(control(a, "rev-parse", "HEAD")))
	oldMod, newMod := filepath.Join(root, "old.mod"), filepath.Join(root, "new.mod")
	for _, path := range []string{oldMod, newMod} {
		updateFixtureWrite(t, path, "module github.com/pascalandy/nicely\n", 0o600)
	}
	oldJSON, err := json.Marshal(map[string]string{"Version": "v0.25.0", "GoMod": oldMod})
	if err != nil {
		t.Fatal(err)
	}
	newJSON, err := json.Marshal(map[string]string{"GoMod": newMod})
	if err != nil {
		t.Fatal(err)
	}
	// The stubs record names only. Their output lives outside either repository,
	// so checking the initial clean tree cannot mask the regression.
	preamble := `#!/bin/sh
set -eu
env | sed -n 's/^\(GIT_[^=]*\)=.*/\1/p' >> "$NCLY_UPDATE_REPORT"
test "$NCLY_UPDATE_KEEP" = kept
`
	updateFixtureWrite(t, filepath.Join(bin, "uv"), preamble+`
test "$*" = 'tool run copier@9.19.1 update --defaults --conflict inline --vcs-ref v0.99.0'
"$NCLY_UPDATE_GIT" -C "$NCLY_UPDATE_TARGET" update-ref refs/heads/copier-probe HEAD
"$NCLY_UPDATE_GIT" -C "$NCLY_UPDATE_TARGET" rev-parse --show-toplevel > "$NCLY_UPDATE_TOP"
printf 'copied\n' > "$NCLY_UPDATE_TARGET/copied.txt"
`, 0o700)
	updateFixtureWrite(t, filepath.Join(bin, "go"), preamble+`
case "$*" in
  'list -m -json github.com/pascalandy/nicely') printf '%s\n' "$NCLY_UPDATE_OLD" ;;
  'mod download -json github.com/pascalandy/nicely@v0.99.0') printf '%s\n' "$NCLY_UPDATE_NEW" ;;
  'mod edit -json '*) printf '{"Replace":[]}\n' ;;
  'get github.com/pascalandy/nicely@v0.99.0') printf 'v0.99.0\n' > "$NCLY_UPDATE_TARGET/pin.txt" ;;
  'mod tidy') ;;
  *) exit 91 ;;
esac
`, 0o700)
	updateFixtureWrite(t, filepath.Join(bin, "just"), preamble+`
test "$*" = check
printf 'checked\n' > "$NCLY_UPDATE_TARGET/checked.txt"
`, 0o700)
	values := map[string]string{
		"HOME": home, "XDG_CONFIG_HOME": home, "PATH": bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"GIT_DIR": filepath.Join(b, ".git"), "GIT_WORK_TREE": b, "GIT_INDEX_FILE": filepath.Join(b, ".git", "index"),
		"GIT_POISON": "sentinel", "NCLY_UPDATE_KEEP": "kept", "NCLY_UPDATE_GIT": git,
		"NCLY_UPDATE_TARGET": a, "NCLY_UPDATE_OLD": string(oldJSON), "NCLY_UPDATE_NEW": string(newJSON),
		"NCLY_UPDATE_REPORT": filepath.Join(root, "git-names"),
		"NCLY_UPDATE_TOP":    filepath.Join(root, "top"),
	}
	for name, value := range values {
		t.Setenv(name, value)
	}
	t.Chdir(a)
	if code := update([]string{"v0.99.0", a}); code != 0 {
		t.Errorf("update exit = %d", code)
	}
	for name, data := range snapshot() {
		if !bytes.Equal(data, before[name]) {
			t.Errorf("second repository changed: %s", name)
		}
	}
	if names := string(read(values["NCLY_UPDATE_REPORT"])); names != "" {
		t.Errorf("update programs inherited Git variable names:\n%s", names)
	}
	if top := strings.TrimSpace(string(read(values["NCLY_UPDATE_TOP"]))); top != a {
		t.Errorf("Copier's Git addressed %q, want %q", top, a)
	}
	if refs := string(control(a, "for-each-ref", "--format=%(objectname)", "refs/heads/copier-probe")); strings.TrimSpace(refs) != targetHead {
		t.Errorf("Copier did not update the target ref: %q", refs)
	}
	for name, want := range map[string]string{"copied.txt": "copied\n", "pin.txt": "v0.99.0\n", "checked.txt": "checked\n"} {
		if got := string(read(filepath.Join(a, name))); got != want {
			t.Errorf("target %s = %q, want %q", name, got, want)
		}
	}
}

func updateFixtureWrite(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}
