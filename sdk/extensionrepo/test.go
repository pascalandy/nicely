package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Notice records a test notice for Test to print after the packages finish.
// Outside the owned run of Test, it writes to stdout, which go test can hide
// for passing packages.
func Notice(message string) error {
	path := os.Getenv(noticeFile)
	if path == "" {
		_, err := fmt.Fprintln(os.Stdout, message)
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(file, message)
	return errors.Join(err, file.Close())
}

// Test runs go test without its result cache, then prints notices to stderr outside Go's output filter.
// It accepts go test arguments, defaults to ./..., and preserves the test exit code.
func Test(args []string) int {
	return test("test", args, nil).deliver(os.Stdout, os.Stderr)
}

// test runs go test in an owned host run, whose notices go into the result.
// The run keeps GOFLAGS and GOWORK of the caller, which signoff masks
// before. set gives variables to go test, such as the scenario to run alone.
func test(action string, args []string, set map[string]string) result {
	dir, err := os.MkdirTemp("", "ncly-test-")
	if err != nil {
		return failed(action, err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	run, err := NewOwnedRun(dir, Profile{Keep: []string{"GOFLAGS", "GOWORK"}, Set: set})
	if err != nil {
		return failed(action, err)
	}
	if len(args) == 0 {
		args = []string{"./..."}
	}
	cmd := exec.Command("go", append([]string{"test", "-count=1"}, args...)...)
	cmd.Env = run.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	r := child(action, cmd.Run())
	notices, err := run.Notices()
	r.notices = notices
	if err != nil {
		r.err = errors.Join(r.err, err)
	}
	return r
}
