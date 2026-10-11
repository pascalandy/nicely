package platform

import "golang.org/x/sys/unix"

// zombie is SZOMB of sys/proc.h.
const zombie = 5

// processes reads the process table from the kernel.
func processes() ([]process, error) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}
	table := make([]process, 0, len(procs))
	for _, p := range procs {
		table = append(table, process{pid: int(p.Proc.P_pid), ppid: int(p.Eproc.Ppid), zombie: p.Proc.P_stat == zombie})
	}
	return table, nil
}

// BecomeSubreaper does nothing: macOS has no subreaper, so a descendant
// whose parent exited before the signal escapes cleanup there.
func BecomeSubreaper() error { return nil }
