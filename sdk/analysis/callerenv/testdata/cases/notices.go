package cases

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// test repeats the go test of sdk/extensionrepo/test.go, which kept the
// environment of the caller beside its notice file. At the parent of
// e35bd0bb, check ran go test with no notice file at all, so a passing
// package hid its notices.
func test(notices string) *exec.Cmd {
	cmd := exec.Command("go", "test", "./...")
	cmd.Env = append(os.Environ(), "NCLY_TEST_NOTICES="+notices) // want `os.Environ: build the environment`
	return cmd
}

// signoffEnv repeats the filter of sdk/extensionrepo/signoff.go, which each
// new caller setting could pass.
func signoffEnv() []string {
	var env []string
	for _, entry := range os.Environ() { // want `os.Environ: build the environment`
		if !strings.HasPrefix(entry, "JUST_") {
			env = append(env, entry)
		}
	}
	return env
}

// A function value reads the environment too, and so does syscall.
var (
	environ    = os.Environ      // want `os.Environ: build the environment`
	rawEnviron = syscall.Environ // want `syscall.Environ: build the environment`
)

// method has the name of the constructor, but a receiver.
type run struct{}

func (run) NewOwnedRun() []string {
	return os.Environ() // want `os.Environ: build the environment`
}
