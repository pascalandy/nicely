package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestLimitsBaseContext(t *testing.T) {
	for _, tc := range []struct {
		name     string
		commits  int
		diverged bool
	}{
		{"uncommitted only", 0, false},
		{"stale base", 3, false},
		{"diverged ref", 3, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			large := strings.Repeat("x", 10001)
			root := limitsRepository(t, map[string]string{"large.txt": large})
			limitsTestGit(t, root, "branch", "base")
			base, err := limitsGit(root, "rev-parse", "--short", "base")
			if err != nil {
				t.Fatal(err)
			}
			for i := range tc.commits {
				write(t, filepath.Join(root, "change.txt"), fmt.Sprintf("Change %d.\n", i))
				limitsTestGit(t, root, "add", "change.txt")
				limitsTestGit(t, root, "commit", "-qm", "Change")
			}
			ref := "base"
			if tc.diverged {
				ref = "elsewhere"
				limitsTestGit(t, root, "checkout", "-qb", ref, "base")
				write(t, filepath.Join(root, "other.txt"), "Other.\n")
				limitsTestGit(t, root, "add", "other.txt")
				limitsTestGit(t, root, "commit", "-qm", "Other")
				limitsTestGit(t, root, "checkout", "-")
			}
			assertLimits(t, root, nil, "--base", ref)
			write(t, filepath.Join(root, "large.txt"), large+"xxx")
			limitsTestGit(t, root, "add", "large.txt")
			assertLimits(t, root, []string{
				fmt.Sprintf("limits: compared working tree with %s (%s), %d commits", strings.TrimSpace(string(base)), ref, tc.commits),
				"large.txt holds 2501", "grows it by 3 bytes",
			}, "--base", ref)
			var out strings.Builder
			err = run([]string{"limits", root}, &out)
			if err == nil || !strings.HasPrefix(err.Error(), "large.txt holds 2501") || out.Len() != 0 {
				t.Fatalf("staged refusal changed: error = %v, output = %q", err, out.String())
			}
		})
	}
}
