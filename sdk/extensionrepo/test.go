package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const noticeFile = "NCLY_TEST_NOTICES"

// Notice records a test notice for Test to print after the packages finish.
// Outside Test, it writes to stdout, which go test can hide for passing packages.
func Notice(message string) error {
	path := os.Getenv(noticeFile)
	if path == "" {
		_, err := fmt.Fprintln(os.Stdout, message)
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(file, message)
	return errors.Join(err, file.Close())
}

// Test runs go test without its result cache, then prints notices outside Go's output filter.
// It accepts go test arguments, defaults to ./..., and preserves the test exit code.
func Test(args []string) int {
	file, err := os.CreateTemp("", "ncly-test-notices-*")
	if err != nil {
		return report("test", err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	if err := file.Close(); err != nil {
		return report("test", err)
	}
	if len(args) == 0 {
		args = []string{"./..."}
	}
	cmd := exec.Command("go", append([]string{"test", "-count=1"}, args...)...)
	cmd.Env = append(os.Environ(), noticeFile+"="+file.Name())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	testErr := cmd.Run()
	notices, err := os.ReadFile(file.Name())
	if err != nil {
		return report("test notices", err)
	}
	if _, err := os.Stdout.Write(notices); err != nil {
		return report("test notices", err)
	}
	var exit *exec.ExitError
	if errors.As(testErr, &exit) && exit.ExitCode() > 0 {
		return exit.ExitCode()
	}
	return report("test", testErr)
}
