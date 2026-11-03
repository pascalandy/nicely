//go:build linux && !amd64 && !arm64

package main

// arch is unknown here, and calls is empty, so the stage refuses to start.
const arch = 0

var calls = map[uint32]call{}
