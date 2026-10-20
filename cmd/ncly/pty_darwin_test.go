package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// ioctlGetTermios reads the state of a terminal.
const ioctlGetTermios = unix.TIOCGETA

// tiocscttyArg is the argument of the ioctl TIOCSCTTY that Go's Setctty
// passes on macOS.
const tiocscttyArg = 0

// settingsSide returns the side of a pseudo-terminal that holds the settings
// that the command left. macOS revokes the descriptor of the terminal once
// the command, which leads its session, exits, while the master still reads
// the same settings.
func settingsSide(ptmx, _ *os.File) *os.File { return ptmx }
