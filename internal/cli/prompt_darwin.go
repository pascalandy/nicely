package cli

import "golang.org/x/sys/unix"

// The requests that read and set the state of a terminal.
const (
	getTermios = unix.TIOCGETA
	setTermios = unix.TIOCSETA
)
