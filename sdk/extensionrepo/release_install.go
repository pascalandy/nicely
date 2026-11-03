package extensionrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
)

// releaseGit names the Git settings that a release keeps. Git URL rewrites
// also let the published install run against a local test remote.
var releaseGit = []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM"}

// releaseEnv is the environment of the owned host run of a release.
func releaseEnv() ([]string, error) {
	run, err := NewOwnedRun("", Profile{Keep: releaseGit})
	return run.Env, err
}

func releaseInstallBlock(readme []byte) (string, error) {
	install, fenced := false, false
	var lines []string
	for _, line := range strings.Split(string(readme), "\n") {
		if strings.HasPrefix(line, "## ") {
			if install {
				break
			}
			install = line == "## Install"
		}
		if !install {
			continue
		}
		if fenced {
			if line == "```" {
				if len(lines) == 0 {
					break
				}
				return strings.Join(lines, "\n"), nil
			}
			lines = append(lines, line)
		} else if line == "```sh" {
			fenced = true
		}
	}
	return "", errors.New("README.md needs a fenced sh block under ## Install")
}

func (r *releaseReceipt) publishedInstall(ctx context.Context, env []string, root string, m contract.Manifest, fixture string) error {
	readme, err := releaseOutput(ctx, env, root, "git", "show", r.SHA+":README.md")
	if err != nil {
		return err
	}
	block, err := releaseInstallBlock([]byte(readme))
	if err != nil {
		return err
	}
	work := filepath.Join(r.dir, "install")
	if err := os.MkdirAll(filepath.Join(work, "bin"), 0o700); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(work) }()
	// Build the pinned host before HOME and the XDG cache move.
	if err := r.require(ctx, env, root, "build-host", "go", "build", "-o", filepath.Join(work, "bin", "ncly"), "github.com/pascalandy/nicely/cmd/ncly"); err != nil {
		return err
	}
	isolated, err := releaseScratchEnv(ctx, env, root, work, m)
	if err != nil {
		return err
	}
	cloneFile := filepath.Join(work, "clone-path")
	script := strings.ReplaceAll(block, "TAG", r.Tag) + "\npwd -P > '" + strings.ReplaceAll(cloneFile, "'", "'\\''") + "'\n"
	if err := r.require(ctx, isolated, work, "readme-install", "sh", "-e", "-c", script); err != nil {
		return err
	}
	data, err := os.ReadFile(cloneFile)
	if err != nil {
		return err
	}
	clone := strings.TrimSpace(string(data))
	rel, err := filepath.Rel(work, clone)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("README install must finish in its scratch clone")
	}
	r.ClonedSHA, err = releaseOutput(ctx, isolated, clone, "git", "rev-parse", "HEAD^{commit}")
	if err != nil {
		return err
	}
	if r.ClonedSHA != r.SHA {
		return fmt.Errorf("README install cloned SHA %s, want tagged SHA %s; correct the install block", r.ClonedSHA, r.SHA)
	}
	if err := r.require(ctx, isolated, clone, "describe", append([]string{"ncly", "describe"}, append(m.Root(), "--json")...)...); err != nil {
		return err
	}
	if err := r.economy(ctx, isolated, clone, m, fixture); err != nil {
		return err
	}
	if err := r.require(ctx, isolated, clone, "uninstall", "just", "uninstall"); err != nil {
		return err
	}
	row, err := r.measure(ctx, isolated, clone, "not-found", append([]string{"ncly", "describe"}, append(m.Root(), "--json")...)...)
	r.Steps = append(r.Steps, row)
	if err != nil {
		return err
	}
	stdout, err := os.ReadFile(filepath.Join(r.dir, row.Stdout))
	if err != nil {
		return err
	}
	stderr, err := os.ReadFile(filepath.Join(r.dir, row.Stderr))
	if err != nil {
		return err
	}
	var outcome contract.Outcome
	if row.Exit != 2 || len(stdout) != 0 || json.Unmarshal(stderr, &outcome) != nil || len(outcome.Errors) != 1 || outcome.Errors[0].Code != contract.NotFound {
		return errors.New("after uninstall, ncly describe must exit 2 with NOT_FOUND; read not-found.stderr")
	}
	return nil
}

// releaseScratchEnv builds the owned scratch run of the install check in
// work. Its bin folder holds the programs of the caller and the built ncly.
func releaseScratchEnv(ctx context.Context, env []string, root, work string, m contract.Manifest) ([]string, error) {
	bin := filepath.Join(work, "bin")
	if err := scratchTools(ctx, env, root, bin, []string{"go", "just", "git", "sh"}, m.Programs); err != nil {
		return nil, err
	}
	set, err := goSettings(ctx, env, root)
	if err != nil {
		return nil, err
	}
	set["NCLY_NO_INPUT"], set["LC_ALL"] = "1", "C"
	run, err := NewOwnedRun(work, Profile{Scratch: true, Path: []string{bin}, Keep: releaseGit, Set: set})
	return run.Env, err
}

// scratchTools links each program of required and optional into bin, as the
// caller resolves it. A shim of a version manager reads its configuration
// under HOME and the XDG folders, which a scratch run moves, so mise which
// answers first. A program of optional can be missing.
func scratchTools(ctx context.Context, env []string, root, bin string, required, optional []string) error {
	if err := os.MkdirAll(bin, 0o700); err != nil {
		return err
	}
	for i, name := range append(append([]string{}, required...), optional...) {
		if _, err := os.Lstat(filepath.Join(bin, name)); err == nil {
			continue
		}
		found, err := releaseOutput(ctx, env, root, "mise", "which", name)
		if err != nil {
			found, err = releaseProgram(env, name)
		}
		if err != nil {
			if i >= len(required) {
				continue
			}
			return err
		}
		if err := os.Symlink(found, filepath.Join(bin, name)); err != nil {
			return err
		}
	}
	return nil
}

// goSettings returns the Go settings of the caller that a scratch run keeps:
// its caches, its module folder, and where modules come from. GOFLAGS and
// GOWORK stay masked.
func goSettings(ctx context.Context, env []string, root string) (map[string]string, error) {
	keys := []string{"GOCACHE", "GOMODCACHE", "GOPATH", "GOPROXY", "GOPRIVATE", "GONOPROXY", "GONOSUMDB", "GOSUMDB", "GOINSECURE", "GOTOOLCHAIN"}
	out, err := releaseOutput(ctx, env, root, "go", append([]string{"env", "-json"}, keys...)...)
	if err != nil {
		return nil, err
	}
	set := map[string]string{}
	if err := json.Unmarshal([]byte(out), &set); err != nil {
		return nil, errors.Join(errors.New("go env -json gives no object"), err)
	}
	for name, value := range set {
		if value == "" {
			delete(set, name)
		}
	}
	return set, nil
}
