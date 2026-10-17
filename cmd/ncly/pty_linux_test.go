package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// ioctlGetTermios reads the state of a terminal.
const ioctlGetTermios = unix.TCGETS

// settingsSide returns the side of a pseudo-terminal that holds the settings
// that the command left. Linux keeps the descriptor of the terminal after
// the command exits, and the master holds settings of its own.
func settingsSide(_, tty *os.File) *os.File { return tty }
