package main

import "golang.org/x/sys/unix"

// arch is the AUDIT_ARCH value of the system calls that calls numbers.
const arch = unix.AUDIT_ARCH_X86_64

// calls holds each system call that makes a link or a rename.
var calls = map[uint32]call{
	unix.SYS_LINK:      {op: "link", oldDir: -1, oldPath: 0, dir: -1, path: 1, flagsIndex: -1},
	unix.SYS_LINKAT:    {op: "link", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: 4},
	unix.SYS_RENAME:    {op: "rename", oldDir: -1, oldPath: 0, dir: -1, path: 1, flagsIndex: -1},
	unix.SYS_RENAMEAT:  {op: "rename", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: -1},
	unix.SYS_RENAMEAT2: {op: "rename", oldDir: 0, oldPath: 1, dir: 2, path: 3, flagsIndex: 4},
}
