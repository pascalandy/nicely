package platform

import (
	"bytes"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// processes reads the process table from /proc.
func processes() ([]process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var table []process
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		stat, err := os.ReadFile("/proc/" + entry.Name() + "/stat")
		if err != nil {
			continue // the process exited meanwhile
		}
		// The command name sits between parentheses and may hold any
		// character, so the fields start after the last parenthesis.
		fields := bytes.Fields(stat[bytes.LastIndexByte(stat, ')')+1:])
		if len(fields) < 2 {
			continue
		}
		ppid, _ := strconv.Atoi(string(fields[1]))
		table = append(table, process{pid: pid, ppid: ppid, zombie: string(fields[0]) == "Z"})
	}
	return table, nil
}

// BecomeSubreaper makes this process adopt every orphan among its
// descendants, so that cleanup still finds a descendant whose parent exited.
func BecomeSubreaper() error {
	return unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0)
}
