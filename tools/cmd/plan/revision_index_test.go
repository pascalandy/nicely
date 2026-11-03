package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRuleIndexResolvesSelectedTemplateRevision(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "sdk/extensionrepo/gate.go"), "package extensionrepo\n")
	file := filepath.Join(root, filepath.FromSlash(templateIndex))
	text := "# Rules\n\n| Rule | Enforced by |\n|---|---|\n| Check the gate | [Gate][gate] |\n\n[gate]: https://github.com/pascalandy/nicely/blob/{{ _commit }}/sdk/extensionrepo/gate.go\n"
	write(t, file, text)
	if problems, err := checkRuleIndexes(root); err != nil || len(problems) != 0 {
		t.Fatalf("selected revision link failed: %v %q", err, problems)
	}
	write(t, file, strings.ReplaceAll(text, "/gate.go", "/missing.go"))
	problems, err := checkRuleIndexes(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, "\n"), "reaches no file") {
		t.Fatalf("missing target passed: %q", problems)
	}
}
