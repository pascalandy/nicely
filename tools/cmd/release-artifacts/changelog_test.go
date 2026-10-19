package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangelog(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	tree := t.TempDir()
	for _, name := range []string{"justfile", "cliff.toml"} {
		testWrite(t, filepath.Join(tree, name), testRead(t, filepath.Join(root, name)))
	}
	testWrite(t, filepath.Join(tree, ".gitignore"), []byte("/.release/\n"))
	env := []string{
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Test", "GIT_COMMITTER_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_EMAIL=test@example.com",
	}
	git := func(args ...string) {
		t.Helper()
		args = append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)
		testCommand(t, tree, env, "git", args...)
	}
	git("-c", "init.templateDir=", "init", "-q")
	git("add", "justfile", "cliff.toml", ".gitignore")
	git("commit", "-qm", "📚 docs: old: already released")
	git("tag", "v0.2.1")
	for _, message := range []string{
		"✨ feat: transcript: add cookies",
		"🐛 fix: transcript: refuse a pipe",
		"⚡️ perf: runner: reduce waits",
		"🧰 skill: release-nicely: teach publication",
		"✨ feat: keep an older unscoped feature",
		"Keep a legacy commit visible",
	} {
		git("commit", "--allow-empty", "-qm", message)
	}
	testCommand(t, tree, env, "just", "changelog", "v0.16.0")
	notes := string(testRead(t, filepath.Join(tree, ".release/changelog.md")))
	if !strings.Contains(notes, "## v0.16.0\n") {
		t.Fatalf("notes must name the requested version:\n%s", notes)
	}
	for group, entries := range map[string][]string{
		"Features":    {"transcript: add cookies", "keep an older unscoped feature"},
		"Fixes":       {"transcript: refuse a pipe"},
		"Performance": {"runner: reduce waits"},
		"Skills":      {"release-nicely: teach publication"},
		"Other":       {"Keep a legacy commit visible"},
	} {
		_, section, found := strings.Cut(notes, "### "+group+"\n\n")
		section, _, _ = strings.Cut(section, "\n### ")
		for _, entry := range entries {
			if !found || !strings.Contains(section, "- "+entry+" ([") {
				t.Fatalf("missing %q under %s:\n%s", entry, group, notes)
			}
		}
	}
	if strings.Contains(notes, "already released") || strings.Count(notes, "](https://github.com/pascalandy/nicely/commit/") != 6 {
		t.Fatalf("notes must contain exactly the six unreleased commits:\n%s", notes)
	}
	git("tag", "v0.16.0")
	testCommand(t, tree, env, "just", "changelog", "v0.16.0")
	if tagged := string(testRead(t, filepath.Join(tree, ".release/changelog.md"))); tagged != notes {
		t.Fatalf("tagging changed the release notes:\n%s", tagged)
	}
	testCommand(t, tree, env, "just", "changelog", "v0.16.0", "v0.2.1")
	if explicit := string(testRead(t, filepath.Join(tree, ".release/changelog.md"))); explicit != notes {
		t.Fatalf("explicit previous tag changed the notes:\n%s", explicit)
	}
	invalid := exec.CommandContext(t.Context(), "just", "changelog", "v01.2.3")
	invalid.Dir = tree
	if out, err := invalid.CombinedOutput(); err == nil || !strings.Contains(string(out), "Usage: just changelog") {
		t.Fatalf("invalid version must fail: %s (%v)", out, err)
	}
}
