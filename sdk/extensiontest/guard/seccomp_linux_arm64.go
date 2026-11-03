package main

import "golang.org/x/sys/unix"

// arch is the AUDIT_ARCH value of the system calls that calls numbers.
const arch = unix.AUDIT_ARCH_AARCH64

// calls holds each system call that makes a link or a rename. arm64 has
// no link and no rename without a folder.
var calls = map[uint32]call{
	unix.SYS_LINKAT:    {op: "link", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: 4},
	unix.SYS_RENAMEAT:  {op: "rename", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: -1},
	unix.SYS_RENAMEAT2: {op: "rename", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: 4},
}
