package extensionrepo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// signoff checks one frozen commit, without the caller's uncommitted files or task environment.
// It runs each just recipe of recipes in order, and stops at the first failure.
// Its answer, Signed off on <sha>, comes only after gh succeeds.
func signoff(recipes []string) result {
	sha, err := signoffCommit(recipes)
	if err != nil {
		return failed("signoff", err)
	}
	return result{action: "signoff", answer: []string{"Signed off on " + sha}}
}

// signoffCommit checks and signs the commit of HEAD, and returns its SHA.
func signoffCommit(recipes []string) (string, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env, err := signoffEnv()
	if err != nil {
		return "", err
	}
	root, err := signoffOutput(ctx, env, "", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	sha, err := signoffOutput(ctx, env, root, "git", "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return "", err
	}
	run, err := newSignoffRun(root, sha)
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr, "signoff: receipt in", filepath.Join(run, "receipt.json"))
	receipt := signoffReceipt{SHA: sha, OS: runtime.GOOS, Tools: map[string]string{}, Checks: []signoffCheck{}}
	if err := receipt.write(run); err != nil {
		return "", err
	}
	worktree := filepath.Join(run, "worktree")
	// Cleanup uses a fresh context because cancellation must still remove the worktree.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := signoffCommand(cleanupCtx, env, root, "git", "worktree", "remove", "--force", worktree)
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "signoff: worktree cleanup: %v: %s\n", err, output)
		}
	}()
	if _, err := signoffOutput(ctx, env, root, "git", "worktree", "add", "--quiet", "--detach", worktree, sha); err != nil {
		return "", err
	}
	for _, tool := range [][]string{{"git", "--version"}, {"go", "version"}, {"just", "--version"}, {"gh", "--version"}} {
		version, err := signoffOutput(ctx, env, worktree, tool[0], tool[1:]...)
		if err != nil {
			return "", err
		}
		receipt.Tools[tool[0]] = version
	}
	if err := receipt.write(run); err != nil {
		return "", err
	}
	log, err := os.OpenFile(filepath.Join(run, "check.log"), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	defer func() { _ = log.Close() }()
	if _, err := fmt.Fprintf(log, "SHA %s\n", sha); err != nil {
		return "", err
	}
	for _, recipe := range recipes {
		if _, err := fmt.Fprintf(log, "$ just %s\n", recipe); err != nil {
			return "", err
		}
		check := []string{"just", "--justfile", filepath.Join(worktree, "justfile"), "--working-directory", worktree, recipe}
		cmd := signoffCommand(ctx, env, worktree, check[0], check[1:]...)
		cmd.Stdout, cmd.Stderr = io.MultiWriter(os.Stdout, log), io.MultiWriter(os.Stderr, log)
		err := cmd.Run()
		receipt.Checks = append(receipt.Checks, signoffCheck{Command: check, Exit: signoffExit(err)})
		if writeErr := receipt.write(run); writeErr != nil {
			return "", writeErr
		}
		if err != nil {
			return "", fmt.Errorf("%s of %s failed: %w", recipe, sha, err)
		}
		if recipe == "finish" {
			// The ledger that finish accepted names the first failing run, which must fail again (#422).
			replay, err := replayFirstRun(ctx, env, root, worktree, run, log)
			receipt.FirstRun = replay
			if writeErr := receipt.write(run); writeErr != nil {
				return "", writeErr
			}
			if err != nil {
				return "", fmt.Errorf("the first failing run of %s: %w", sha, err)
			}
			fmt.Fprintf(os.Stderr, "signoff: the first scenario fails again at %s, on %q\n", replay.Commit, replay.Failed)
		}
	}
	if err := signoffUnchanged(ctx, env, worktree, sha); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	args := []string{"gh", "signoff", "--commit", sha}
	cmd := signoffCommand(ctx, env, worktree, args[0], args[1:]...)
	// Only this program prints the landing receipt line, after gh succeeds.
	cmd.Stdout, cmd.Stderr = log, io.MultiWriter(os.Stderr, log)
	err = cmd.Run()
	receipt.Signoff = &signoffCheck{Command: args, Exit: signoffExit(err)}
	if writeErr := receipt.write(run); writeErr != nil {
		return "", writeErr
	}
	if err != nil {
		return "", fmt.Errorf("gh signoff of %s failed: %w", sha, err)
	}
	return sha, nil
}

func signoffUnchanged(ctx context.Context, env []string, worktree, sha string) error {
	head, err := signoffOutput(ctx, env, worktree, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if head != sha {
		return errors.New("HEAD moved during the check; commit the intended files and repeat just signoff")
	}
	status, err := signoffOutput(ctx, env, worktree, "git", "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return err
	}
	if status != "" {
		return errors.New("tracked files changed during the check; commit the fixes and repeat just signoff")
	}
	return nil
}

// signoffEnv is the environment of an owned host run, without the caller
// state that Owned runs in the contract lists. Persisted Go settings, such
// as GOPRIVATE, stay.
func signoffEnv() ([]string, error) {
	run, err := NewOwnedRun("", Profile{})
	return run.Env, err
}

func signoffCommand(ctx context.Context, env []string, dir, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir, cmd.Env = dir, env
	// Cancel the group too, so a stopped check cannot leave a child using the worktree.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 5 * time.Second
	return cmd
}

func signoffOutput(ctx context.Context, env []string, dir, name string, args ...string) (string, error) {
	output, err := signoffCommand(ctx, env, dir, name, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func signoffExit(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 {
		return exitErr.ExitCode()
	}
	return 1
}
