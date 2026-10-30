package extensionrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const nicelyModule = "github.com/pascalandy/nicely"

func update(args []string) int {
	if len(args) < 2 || len(args) > 3 || !regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(args[0]) ||
		!filepath.IsAbs(args[1]) || (len(args) == 3 && args[2] != "--continue") {
		fmt.Fprintln(os.Stderr, "usage: update vX.Y.Z /absolute/repository [--continue]")
		return 2
	}
	if err := os.Chdir(args[1]); err != nil {
		return report("update", err)
	}
	return report("update", updateRepository(args[0], args[1], len(args) == 3))
}

func updateRepository(version, repo string, resume bool) error {
	if !resume {
		head, err := updateOutput("git", "rev-parse", "--verify", "HEAD")
		if err != nil {
			return errors.New("commit the repository before updating")
		}
		status, err := updateOutput("git", "status", "--porcelain")
		if err != nil {
			return err
		}
		if len(status) != 0 {
			return fmt.Errorf("the update needs a clean Git tree. Commit your changes, then run just update %s again", version)
		}
		if err := updateRun("uv", "tool", "run", "copier@9.19.1", "update", "--defaults", "--conflict", "inline", "--vcs-ref", version); err != nil {
			fmt.Fprintln(os.Stderr, "update: Copier stopped. Files can be incomplete without a conflict marker.")
			fmt.Fprintln(os.Stderr, "Return to the previous commit and inspect the added files before removing them:")
			fmt.Fprintf(os.Stderr, "  git restore --source=%s --staged --worktree -- .\n", strings.TrimSpace(string(head)))
			fmt.Fprintln(os.Stderr, "  git clean -d --dry-run")
			fmt.Fprintln(os.Stderr, "  git clean -d --force")
			fmt.Fprintln(os.Stderr, "Then retry:")
			fmt.Fprintln(os.Stderr, "  "+updateBootstrap(version, repo, false))
			return err
		}
	} else {
		data, err := os.ReadFile(".copier-answers.yml")
		if err != nil {
			return err
		}
		var answers struct {
			Commit string `yaml:"_commit"`
		}
		if err := yaml.Unmarshal(data, &answers); err != nil {
			return err
		}
		if answers.Commit != version {
			return fmt.Errorf("continuation needs _commit: %s in .copier-answers.yml", version)
		}
	}
	conflicts, err := updateConflicts()
	if err != nil {
		return err
	}
	if len(conflicts) != 0 {
		fmt.Fprintln(os.Stderr, "update: the template conflicts with your changes in these files:")
		for _, path := range conflicts {
			fmt.Fprintln(os.Stderr, "  "+path)
		}
		fmt.Fprintln(os.Stderr, "Resolve each conflict, mark its file with git add, and remove each .rej file.")
		fmt.Fprintln(os.Stderr, "Then finish the update:")
		fmt.Fprintln(os.Stderr, "  "+updateBootstrap(version, repo, true))
		return errors.New("resolve the template conflicts before updating dependencies")
	}
	previous, err := updateOutput("go", "list", "-m", "-json", nicelyModule)
	if err != nil {
		return err
	}
	var old struct {
		Version string
		GoMod   string
	}
	if err := json.Unmarshal(previous, &old); err != nil {
		return err
	}
	if old.GoMod == "" {
		old.GoMod, err = updateModuleFile(old.Version)
		if err != nil {
			return err
		}
	}
	target, err := updateModuleFile(version)
	if err != nil {
		return err
	}
	edits, err := updateReplacements(old.GoMod, target)
	if err != nil {
		return err
	}
	if err := updateRun("go", "get", nicelyModule+"@"+version); err != nil {
		return err
	}
	if len(edits) != 0 {
		if err := updateRun("go", append([]string{"mod", "edit"}, edits...)...); err != nil {
			return err
		}
	}
	if err := updateRun("go", "mod", "tidy"); err != nil {
		return err
	}
	return updateRun("just", "check")
}

func updateModuleFile(version string) (string, error) {
	data, err := updateOutput("go", "mod", "download", "-json", nicelyModule+"@"+version)
	if err != nil {
		return "", err
	}
	var module struct{ GoMod string }
	if err := json.Unmarshal(data, &module); err != nil {
		return "", err
	}
	if module.GoMod == "" {
		return "", errors.New("the downloaded Nicely module has no go.mod")
	}
	return module.GoMod, nil
}

func updateReplacements(previous, target string) ([]string, error) {
	type module struct{ Path, Version string }
	type replacement struct{ Old, New module }
	key := func(m module) string {
		if m.Version != "" {
			return m.Path + "@" + m.Version
		}
		return m.Path
	}
	edits := make(map[string]string)
	for i, path := range []string{previous, target} {
		data, err := updateOutput("go", "mod", "edit", "-json", path)
		if err != nil {
			return nil, err
		}
		var file struct{ Replace []replacement }
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, err
		}
		for _, r := range file.Replace {
			old := key(r.Old)
			edits[old] = "-dropreplace=" + old
			if i == 1 {
				edits[old] = "-replace=" + old + "=" + key(r.New)
			}
		}
	}
	var result []string
	for _, edit := range edits {
		result = append(result, edit)
	}
	slices.Sort(result)
	return result, nil
}

func updateConflicts() ([]string, error) {
	unmerged, err := updateOutput("git", "diff", "--name-only", "--diff-filter=U", "-z")
	if err != nil {
		return nil, err
	}
	paths := make(map[string]bool)
	for _, path := range strings.Split(string(unmerged), "\x00") {
		if path != "" {
			paths[path] = true
		}
	}
	if err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".rej") {
			paths[path] = true
		}
		return nil
	}); err != nil {
		return nil, err
	}
	marker := regexp.MustCompile(`(?m)^(<<<<<<<|>>>>>>>)( |$)`)
	for _, args := range [][]string{{"diff", "--name-only", "-z", "HEAD"}, {"ls-files", "--others", "--exclude-standard", "-z"}} {
		changed, err := updateOutput("git", args...)
		if err != nil {
			return nil, err
		}
		for _, path := range strings.Split(string(changed), "\x00") {
			if path == "" {
				continue
			}
			info, err := os.Lstat(path)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if !info.Mode().IsRegular() {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if marker.Match(data) {
				paths[path] = true
			}
		}
	}
	var result []string
	for path := range paths {
		result = append(result, path)
	}
	slices.Sort(result)
	return result, nil
}

func updateBootstrap(version, repo string, resume bool) string {
	command := `go -C "$(go mod download -json ` + nicelyModule + `@` + version + ` | sed -n 's/^[[:space:]]*"Dir": "\(.*\)",$/\1/p')" run ./sdk/extensionrepo/cmd/extensionrepo update ` + version + " '" + strings.ReplaceAll(repo, "'", "'\\''") + "'"
	if resume {
		command += " --continue"
	}
	return command
}

func updateRun(name string, args ...string) error {
	fmt.Fprintln(os.Stderr, strings.Join(append([]string{name}, args...), " "))
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func updateOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
