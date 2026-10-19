package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// limitsGit returns the standard output of git alone, because a warning on
// stderr, such as a skipped rename detection, would corrupt a parsed diff.
func limitsGit(root string, args ...string) ([]byte, error) {
	data, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, exit.Stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", args[0], err)
	}
	return data, nil
}

// gitlink is the git mode of a submodule entry.
const gitlink = "160000"

// readChanges reads the raw diff, whose modes tell a gitlink apart. A gitlink
// holds no file to read, because a submodule's commits live in another
// repository. A submodule replaced by a file counts as a new file. Each path
// reads as ./path, so a name such as 0:notes.md never selects an index stage.
func readChanges(root string, refs []string, old string, worktree bool) ([]change, error) {
	data, err := limitsGit(root, append([]string{"diff", "--raw", "-z", "-M", "-l0", "--diff-filter=ACMRDT"}, refs...)...)
	if err != nil {
		return nil, err
	}
	fields := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
	var changes []change
	for i := 0; i+1 < len(fields); {
		meta := strings.Fields(strings.TrimPrefix(fields[i], ":"))
		if len(meta) != 5 {
			return nil, fmt.Errorf("git diff --raw: unexpected entry %q", fields[i])
		}
		oldMode, newMode, status, before := meta[0], meta[1], meta[4], fields[i+1]
		i += 2
		after := before
		if status[0] == 'R' || status[0] == 'C' {
			after = fields[i]
			i++
		}
		if newMode == gitlink || (status == "D" && oldMode == gitlink) {
			continue
		}
		c := change{path: after}
		if status != "A" && oldMode != gitlink {
			c.old, err = limitsGit(root, "show", old+":./"+before)
			if err != nil {
				return nil, err
			}
			if c.old == nil {
				c.old = []byte{}
			}
		}
		if status != "D" {
			if worktree {
				c.new, err = readWorktreeFile(root, after)
			} else {
				c.new, err = limitsGit(root, "show", ":./"+after)
			}
			if err != nil {
				return nil, err
			}
		}
		changes = append(changes, c)
	}
	if worktree {
		data, err := limitsGit(root, "ls-files", "--others", "--exclude-standard", "-z")
		if err != nil {
			return nil, err
		}
		for name := range strings.SplitSeq(strings.TrimSuffix(string(data), "\x00"), "\x00") {
			// Untracked repositories are directories, with no file to check.
			if name == "" || strings.HasSuffix(name, "/") {
				continue
			}
			content, err := readWorktreeFile(root, name)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change{path: name, new: content})
		}
	}
	return changes, nil
}

// readWorktreeFile reads a symlink's target text, as git stores it, without following it.
func readWorktreeFile(root, name string) ([]byte, error) {
	file := filepath.Join(root, filepath.FromSlash(name))
	info, err := os.Lstat(file)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(file)
		return []byte(target), err
	}
	return os.ReadFile(file)
}
