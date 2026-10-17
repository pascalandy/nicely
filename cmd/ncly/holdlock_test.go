package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// holdLock holds a lock file as another ncly would, so a scenario proves
// what a command does while the lock is busy: holdlock <lock> <ready>.
// It takes the lock with flock, as sdk/platform does, writes the file
// ready once it holds it, then waits until a signal ends it, which
// releases the lock. It ends on its own after a minute, so a failed
// scenario leaves nothing running.
func holdLock() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: holdlock <lock> <ready>")
		os.Exit(2)
	}
	f, err := os.OpenFile(os.Args[1], os.O_RDWR|os.O_CREATE, 0o600)
	if err == nil {
		err = unix.Flock(int(f.Fd()), unix.LOCK_EX)
	}
	if err == nil {
		err = os.WriteFile(os.Args[2], nil, 0o644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "holdlock:", err)
		os.Exit(1)
	}
	time.Sleep(time.Minute)
}
