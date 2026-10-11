// Package runner runs the programs that ncly and its extensions start, as
// the Programs that ncly runs section of contract.md defines: their
// environment, their process group, and their exit code.
package runner

import (
	"errors"
	"io"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
)

// Program is one program to run.
type Program struct {
	Path string
	Args []string
	Env  []string
	// Stdin, Stdout, and Stderr reach the program. An *os.File passes as it
	// is, so the program sees the terminal when it is one.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Terminal keeps the program in the process group of ncly, because it
	// may read the terminal, as an extension does. Any other program starts
	// in its own process group.
	Terminal bool
}

// Run starts the program and waits for it to exit. It returns the
// program's exit code, and 128 plus the number of the signal that ended it,
// as a shell reports it.
func Run(p Program) (int, error) {
	cmd := exec.Command(p.Path, p.Args...)
	cmd.Env = p.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = p.Stdin, p.Stdout, p.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: !p.Terminal}
	err := cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return 0, err
	}
	return exitCode(cmd.ProcessState), nil
}

func exitCode(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}

// Environ returns the environment of a program: its parent's, without the
// key of any service, which needs a grant, and with each variable of set
// given its value, or removed when that value is empty.
func Environ(parent []string, set map[string]string) []string {
	env := slices.DeleteFunc(slices.Clone(parent), func(entry string) bool {
		name, _, _ := strings.Cut(entry, "=")
		_, changed := set[name]
		return changed || strings.HasSuffix(name, "_API_KEY")
	})
	for _, name := range slices.Sorted(maps.Keys(set)) {
		if value := set[name]; value != "" {
			env = append(env, name+"="+value)
		}
	}
	return env
}
