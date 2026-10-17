package tui

import "golang.org/x/sys/unix"

// The requests that read the state of a terminal, and that set it once its
// output is written, dropping the input that it holds.
const (
	ioctlGetTermios      = unix.TIOCGETA
	ioctlSetTermiosFlush = unix.TIOCSETAF
)
