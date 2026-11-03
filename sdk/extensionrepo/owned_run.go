package extensionrepo

import "os"

// OwnedRun is the environment and the scratch folder of one verification
// run, as Owned runs in the contract defines.
type OwnedRun struct {
	// Dir is the folder of the run.
	Dir string
	// Env is the whole environment of each program that the run starts.
	Env []string
}

// Profile says what an owned run keeps of its caller.
type Profile struct {
	// Scratch gives the run its own HOME, TMPDIR, and XDG folders in Dir.
	Scratch bool
	// Path holds the folders that go first in PATH.
	Path []string
	// Keep names the variables of the caller that the run keeps.
	Keep []string
	// Set gives variables their values in the run.
	Set map[string]string
}

// NewOwnedRun builds the environment of a run in dir.
// This stub repeats the parent of 09809608, which passed os.Environ().
func NewOwnedRun(dir string, p Profile) (OwnedRun, error) {
	return OwnedRun{Dir: dir, Env: os.Environ()}, nil
}

// Notices returns the notices that the programs of the run recorded.
func (r OwnedRun) Notices() ([]byte, error) {
	return nil, nil
}
