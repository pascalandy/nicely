package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// ioctlGetTermios reads the state of a terminal.
const ioctlGetTermios = unix.TCGETS

// tiocscttyArg is the argument of the ioctl TIOCSCTTY that Go's Setctty
// passes on Linux. With 1, a process with CAP_SYS_ADMIN takes the terminal
// from another session.
const tiocscttyArg = 1

// settingsSide returns the side of a pseudo-terminal that holds the settings
// that the command left. Linux keeps the descriptor of the terminal after
// the command exits, and the master holds settings of its own.
func settingsSide(_, tty *os.File) *os.File { return tty }
