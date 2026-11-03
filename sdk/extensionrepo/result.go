package extensionrepo

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// result is what one action of the export program gives back to its caller:
// the answer, the notices, and the exit code. Each action returns one, and
// deliver alone writes it. At the parent of e35bd0bb, check ran go test
// outside the notice file, and a passing package hid its notices.
type result struct {
	// action names the action in each line of a failure.
	action string
	// answer holds the lines for stdout, such as Signed off on <sha>.
	answer []string
	// notices holds what the programs of the run recorded with Notice.
	notices []byte
	// exit is the exit code of the child that decided the action, or 0.
	exit int
	// err is the refusal of the action, or nil.
	err error
}

// failed returns the result of an action that stops on err.
func failed(action string, err error) result {
	return result{action: action, err: err}
}

// child returns the result of an action that ran one child program.
// A child that exits with a code keeps it, and prints its own failure.
func child(action string, err error) result {
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() > 0 {
		return result{action: action, exit: exit.ExitCode()}
	}
	return result{action: action, err: err}
}

// result returns the result of an action that ran one child program in run,
// with the notices that its programs recorded there.
func (run OwnedRun) result(action string, err error) result {
	r := child(action, err)
	notices, readErr := run.Notices()
	r.notices = notices
	if readErr != nil {
		r.err = errors.Join(r.err, readErr)
	}
	return r
}

// deliver writes the answer to stdout, then the notices and any refusal to
// stderr, and returns the exit code. An answer or notices that a stream
// refuses fail the action, because the caller never received them.
func (r result) deliver(stdout, stderr io.Writer) int {
	var lost error
	for _, line := range r.answer {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			lost = err
			break
		}
	}
	if len(r.notices) != 0 {
		if _, err := stderr.Write(r.notices); err != nil {
			lost = errors.Join(lost, err)
		}
	}
	if r.err != nil {
		_, _ = fmt.Fprintln(stderr, r.action+":", r.err)
	}
	if lost != nil {
		_, _ = fmt.Fprintln(stderr, r.action+": output was not delivered:", lost)
	}
	switch {
	case r.exit > 0:
		return r.exit
	case r.err != nil || lost != nil:
		return 1
	}
	return 0
}
