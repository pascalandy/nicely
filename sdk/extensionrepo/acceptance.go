package extensionrepo

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// acceptance runs acceptance.sh of the current folder in an owned host run,
// with the fixture of args. No Nicely setting of the caller, such as
// NCLY_CONFIG, reaches the command of ## Acceptance. The run keeps HOME, so
// acceptance.sh still finds the programs of a version manager before it
// moves the XDG folders. Its TMPDIR, where acceptance.sh makes its scratch
// folder, is the folder of the run.
func acceptance(args []string) result {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dir, err := os.MkdirTemp("", "ncly-acceptance-run-")
	if err != nil {
		return failed("acceptance", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	run, err := NewOwnedRun(dir, Profile{Set: map[string]string{"TMPDIR": dir}})
	if err != nil {
		return failed("acceptance", err)
	}
	bash, err := releaseProgram(run.Env, "bash")
	if err != nil {
		return failed("acceptance", err)
	}
	cmd := exec.CommandContext(ctx, bash, append([]string{"acceptance.sh"}, args...)...)
	cmd.Env = run.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return run.result("acceptance", cmd.Run())
}
