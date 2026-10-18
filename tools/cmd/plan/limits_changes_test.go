package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLimitsChanges(t *testing.T) {
	long := strings.Repeat("word ", 25) + "last.\n"
	cases := []struct {
		name string
		edit func(*testing.T, string)
		args []string
		want []string
	}{
		{"all problems", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "AGENTS.md"), strings.Repeat("x", 5001))
			write(t, filepath.Join(root, "large.txt"), strings.Repeat("x", 10002))
			write(t, filepath.Join(root, "new.md"), long)
			write(t, filepath.Join(root, "second.md"), long)
		}, nil, []string{"AGENTS.md holds 1251", "large.txt holds 2501", "new.md:1", "second.md:1", "AGENTS.md only routes", "Split each file", "Rewrite each sentence"}},
		{"move between files and rewrap", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "old.md"), "Short.\n")
			write(t, filepath.Join(root, "new.md"), "# New\n\n"+strings.Repeat("word ", 15)+"\n"+strings.Repeat("word ", 10)+"last.\n")
		}, nil, nil},
		{"empty old file", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "empty.txt"), "")
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Empty")
			write(t, filepath.Join(root, "empty.txt"), strings.Repeat("x", 10001))
		}, nil, []string{"empty.txt holds 2501", "grows it by 10001 bytes"}},
		{"move from deleted file", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "old.md"), long+strings.Repeat("Earlier fact.\n", 30))
			limitsTestGit(t, root, "add", "old.md")
			limitsTestGit(t, root, "commit", "-qm", "Larger source")
			limitsTestGit(t, root, "rm", "old.md")
			write(t, filepath.Join(root, "new.md"), "# New\n\n"+long)
		}, nil, nil},
		{"rename", func(t *testing.T, root string) {
			limitsTestGit(t, root, "mv", "old.md", "new.md")
			write(t, filepath.Join(root, "new.md"), long+"\nShort.\n")
		}, nil, nil},
		{"rename size", func(t *testing.T, root string) {
			limitsTestGit(t, root, "mv", "large.txt", "renamed.txt")
			write(t, filepath.Join(root, "renamed.txt"), strings.Repeat("x", 10002))
		}, nil, []string{"renamed.txt holds 2501", "grows it by 1 bytes"}},
		{"rename beyond configured limit", func(t *testing.T, root string) {
			limitsTestGit(t, root, "config", "diff.renameLimit", "1")
			for _, name := range []string{"first", "second"} {
				write(t, filepath.Join(root, name+".txt"), strings.Repeat(name+"\n", 2200))
			}
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Files")
			for _, name := range []string{"first", "second"} {
				limitsTestGit(t, root, "mv", name+".txt", name+"-renamed.txt")
				write(t, filepath.Join(root, name+"-renamed.txt"), strings.Repeat(name+"\n", 2100))
			}
		}, nil, nil},
		{"quoted path", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "odd\tname.md"), long)
			write(t, filepath.Join(root, "odd name.md"), long)
		}, nil, []string{"odd\tname.md:1", "odd name.md:1", "26 words"}},
		{"staged only", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "new.md"), long)
			limitsTestGit(t, root, "add", ".")
			write(t, filepath.Join(root, "new.md"), "Short.\n")
		}, nil, []string{"new.md:1", "26 words"}},
		{"stage-like path", func(t *testing.T, root string) {
			write(t, filepath.Join(root, "short.md"), "Short.\n")
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Short")
			write(t, filepath.Join(root, "0:short.md"), long)
		}, nil, []string{"0:short.md:1", "26 words"}},
		{"submodule", func(t *testing.T, root string) {
			nestedRepository(t, root, "sub")
		}, nil, nil},
		{"submodule replaced by a file", func(t *testing.T, root string) {
			nestedRepository(t, root, "sub.md")
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Submodule")
			if err := os.RemoveAll(filepath.Join(root, "sub.md")); err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "sub.md"), long)
		}, nil, []string{"sub.md:1", "26 words"}},
		{"submodule replaced by a file since the base", func(t *testing.T, root string) {
			nestedRepository(t, root, "sub.md")
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Submodule")
			limitsTestGit(t, root, "branch", "base")
			if err := os.RemoveAll(filepath.Join(root, "sub.md")); err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "sub.md"), long)
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "File")
		}, []string{"--base", "base"}, []string{"sub.md:1", "26 words"}},
		{"base", func(t *testing.T, root string) {
			limitsTestGit(t, root, "branch", "base")
			write(t, filepath.Join(root, "new.md"), long)
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Change")
			limitsTestGit(t, root, "checkout", "-qb", "elsewhere", "base")
			write(t, filepath.Join(root, "other.txt"), "Other.\n")
			limitsTestGit(t, root, "add", ".")
			limitsTestGit(t, root, "commit", "-qm", "Other")
			limitsTestGit(t, root, "checkout", "-")
		}, []string{"--base", "elsewhere"}, []string{"new.md:1", "26 words"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := limitsRepository(t, map[string]string{"old.md": long, "large.txt": strings.Repeat("x", 10001)})
			tc.edit(t, root)
			if tc.name != "staged only" {
				limitsTestGit(t, root, "add", ".")
			}
			assertLimits(t, root, tc.want, tc.args...)
		})
	}
}

// nestedRepository makes name a repository with one commit, which git adds
// to root as a submodule.
func nestedRepository(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	limitsTestGit(t, dir, "init", "-q")
	limitsTestGit(t, dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "Sub")
}
