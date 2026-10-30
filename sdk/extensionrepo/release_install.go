package extensionrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
)

func releaseEnv() []string {
	env := signoffEnv()
	// Git URL rewrites also let the published install run against a local test remote.
	for _, key := range []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
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

func releaseScratchEnv(ctx context.Context, env []string, root, work string, m contract.Manifest) ([]string, error) {
	bin := filepath.Join(work, "bin")
	path := []string{bin}
	// Resolve version-manager shims while they can still read the caller's configuration.
	for i, name := range append([]string{"go", "just", "git", "sh"}, m.Programs...) {
		if _, err := os.Lstat(filepath.Join(bin, name)); err == nil {
			continue
		}
		found, err := releaseOutput(ctx, env, root, "mise", "which", name)
		if err != nil {
			found, err = exec.LookPath(name)
		}
		if err != nil {
			if i >= 4 {
				continue
			}
			return nil, err
		}
		if err := os.Symlink(found, filepath.Join(bin, name)); err != nil {
			return nil, err
		}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if !filepath.IsAbs(dir) {
			continue
		}
		if found, _ := filepath.Glob(filepath.Join(dir, "ncly-*")); len(found) == 0 {
			path = append(path, dir)
		}
	}
	isolated := []string{
		"HOME=" + work, "PATH=" + strings.Join(slices.Compact(path), string(filepath.ListSeparator)),
		"TMPDIR=" + work, "XDG_CONFIG_HOME=" + filepath.Join(work, "config"), "XDG_DATA_HOME=" + filepath.Join(work, "data"),
		"XDG_STATE_HOME=" + filepath.Join(work, "state"), "XDG_CACHE_HOME=" + filepath.Join(work, "cache"),
		"DBUS_SESSION_BUS_ADDRESS=disabled:", "NCLY_NO_INPUT=1", "GOWORK=off", "GOFLAGS= ", "LC_ALL=C",
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
		value, err := releaseOutput(ctx, env, root, "go", "env", key)
		if err != nil {
			return nil, err
		}
		isolated = append(isolated, key+"="+value)
	}
	for _, key := range []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM"} {
		if value, ok := os.LookupEnv(key); ok {
			isolated = append(isolated, key+"="+value)
		}
	}
	return isolated, nil
}
