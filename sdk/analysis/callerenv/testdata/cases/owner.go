package cases

import "os"

// NewOwnedRun is the one function of an owned package that reads the
// environment of the caller as a whole, in a closure too.
func NewOwnedRun() []string {
	read := func() []string { return os.Environ() }
	return read()
}

// home reads one variable, which the rule allows.
func home() string {
	return os.Getenv("HOME")
}
