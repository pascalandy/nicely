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
func signoff() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env := signoffEnv()
	root, err := signoffOutput(ctx, env, "", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	sha, err := signoffOutput(ctx, env, root, "git", "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return err
	}
	run, err := newSignoffRun(root, sha)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "signoff: receipt in", filepath.Join(run, "receipt.json"))
	receipt := signoffReceipt{SHA: sha, OS: runtime.GOOS, Tools: map[string]string{}, Checks: []signoffCheck{}}
	if err := receipt.write(run); err != nil {
		return err
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
		return err
	}
	for _, tool := range [][]string{{"git", "--version"}, {"go", "version"}, {"just", "--version"}, {"gh", "--version"}} {
		version, err := signoffOutput(ctx, env, worktree, tool[0], tool[1:]...)
		if err != nil {
			return err
		}
		receipt.Tools[tool[0]] = version
	}
	if err := receipt.write(run); err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(run, "check.log"), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = log.Close() }()
	if _, err := fmt.Fprintf(log, "SHA %s\n$ just check\n", sha); err != nil {
		return err
	}
	check := []string{"just", "--justfile", filepath.Join(worktree, "justfile"), "--working-directory", worktree, "check"}
	cmd := signoffCommand(ctx, env, worktree, check[0], check[1:]...)
	cmd.Stdout, cmd.Stderr = io.MultiWriter(os.Stdout, log), io.MultiWriter(os.Stderr, log)
	err = cmd.Run()
	receipt.Checks = append(receipt.Checks, signoffCheck{Command: check, Exit: signoffExit(err)})
	if writeErr := receipt.write(run); writeErr != nil {
		return writeErr
	}
	if err != nil {
		return fmt.Errorf("check of %s failed: %w", sha, err)
	}
	if err := signoffUnchanged(ctx, env, worktree, sha); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	args := []string{"gh", "signoff", "--commit", sha}
	cmd = signoffCommand(ctx, env, worktree, args[0], args[1:]...)
	// Only this program prints the landing receipt line, after gh succeeds.
	cmd.Stdout, cmd.Stderr = log, io.MultiWriter(os.Stderr, log)
	err = cmd.Run()
	receipt.Signoff = &signoffCheck{Command: args, Exit: signoffExit(err)}
	if writeErr := receipt.write(run); writeErr != nil {
		return writeErr
	}
	if err != nil {
		return fmt.Errorf("gh signoff of %s failed: %w", sha, err)
	}
	fmt.Println("Signed off on", sha)
	return nil
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

func signoffEnv() []string {
	var env []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, "JUST_") || strings.HasPrefix(name, "GIT_") {
			continue
		}
		switch name {
		case "GOFLAGS", "GOWORK", "GOENV":
			continue
		}
		env = append(env, entry)
	}
	return append(env, "GOWORK=off", "GOENV=off")
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
