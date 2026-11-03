//go:build !linux

package main

import "runtime"

// supervise refuses: only Linux gives a filter to a child alone.
func supervise(options) Report {
	return Report{Operations: []Operation{}, Children: []string{}, Error: runtime.GOOS + " has no seccomp filter for a child alone"}
}

func stage([]string) int { return 125 }
