package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLimitsWorktree(t *testing.T) {
	long := strings.Repeat("word ", 25) + "last.\n"
	large := strings.Repeat("x", 10001)
	cases := []struct {
		name string
		edit func(*testing.T, string)
		want []string
	}{
		{"unstaged size", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "tracked.txt"), large)
		}, []string{"tracked.txt holds 2501"}},
		{"staged sentence", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "new.md"), long)
			limitsTestGit(t, root, "add", "new.md")
		}, []string{"new.md:1", "26 words"}},
		{"untracked size and sentence", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "odd\tname.md"), long)
			write(t, filepath.Join(root, "new.txt"), large)
		}, []string{"odd\tname.md:1", "26 words", "new.txt holds 2501"}},
		{"committed and uncommitted", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "committed.md"), long)
			limitsTestGit(t, root, "add", "committed.md")
			limitsTestGit(t, root, "commit", "-qm", "Change")
			write(t, filepath.Join(root, "tracked.txt"), large)
		}, []string{"committed.md:1", "26 words", "tracked.txt holds 2501"}},
		{"unstaged content after staging", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "new.md"), "Short.\n")
			limitsTestGit(t, root, "add", "new.md")
			write(t, filepath.Join(root, "new.md"), long)
		}, []string{"new.md:1", "26 words"}},
		{"working tree repairs staged content", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "new.md"), long)
			limitsTestGit(t, root, "add", "new.md")
			assertLimits(t, root, []string{"new.md:1", "26 words"})
			write(t, filepath.Join(root, "new.md"), "Short.\n")
		}, nil},
		{"deletion", func(t *testing.T, root string) {
			limitsTestGit(t, root, "rm", "tracked.txt")
		}, nil},
		{"ignored file", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "ignored.txt"), large)
		}, nil},
		{"symlinks", func(t *testing.T, root string) {
			for _, name := range []string{"staged.md", "untracked.md"} {
				if err := os.Symlink("missing.txt", filepath.Join(root, name)); err != nil {
					t.Fatal(err)
				}
			}
			limitsTestGit(t, root, "add", "staged.md")
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := limitsRepository(t, map[string]string{
				"tracked.txt": "Small.\n",
				".gitignore":  "ignored.txt\n",
			})
			limitsTestGit(t, root, "branch", "base")
			tc.edit(t, root)
			assertLimits(t, root, tc.want, "--base", "base")
		})
	}
}
