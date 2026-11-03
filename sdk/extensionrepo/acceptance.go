package extensionrepo

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// acceptance runs acceptance.sh of the current folder in an owned scratch
// run, with the fixture of args. No setting of the caller, such as
// NCLY_CONFIG, reaches the command of ## Acceptance.
func acceptance(e extension.Extension, args []string) result {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return failed("acceptance", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return failed("acceptance", err)
	}
	host, err := NewOwnedRun("", Profile{})
	if err != nil {
		return failed("acceptance", err)
	}
	dir, err := os.MkdirTemp("", "ncly-acceptance-run-")
	if err != nil {
		return failed("acceptance", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	bin := filepath.Join(dir, "bin")
	if err := scratchTools(ctx, host.Env, root, bin, []string{"bash", "go", "just"}, append([]string{"git", "sh"}, m.Programs...)); err != nil {
		return failed("acceptance", err)
	}
	set, err := goSettings(ctx, host.Env, root)
	if err != nil {
		return failed("acceptance", err)
	}
	run, err := NewOwnedRun(dir, Profile{Scratch: true, Path: []string{bin}, Set: set})
	if err != nil {
		return failed("acceptance", err)
	}
	cmd := exec.CommandContext(ctx, filepath.Join(bin, "bash"), append([]string{"acceptance.sh"}, args...)...)
	cmd.Dir, cmd.Env = root, run.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return child("acceptance", cmd.Run())
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
