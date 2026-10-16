package main

import "golang.org/x/sys/unix"

// ioctlGetTermios reads the state of a terminal.
const ioctlGetTermios = unix.TCGETS
