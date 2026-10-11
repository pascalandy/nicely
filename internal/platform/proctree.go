// Package platform holds what ncly needs from the operating system: for
// now, the process tree that cleanup reaches, as Programs that ncly runs in
// contract.md defines.
package platform

import "slices"

// process is one entry of the process table.
type process struct {
	pid, ppid int
	zombie    bool
}

// Descendants returns the running processes whose chain of parents leads
// to pid, found through their parent process, so a descendant that left
// the process group of its parent is still found.
func Descendants(pid int) []int {
	table, err := processes()
	if err != nil {
		return nil
	}
	children := map[int][]process{}
	for _, p := range table {
		children[p.ppid] = append(children[p.ppid], p)
	}
	var found []int
	queue := []int{pid}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		for _, child := range children[parent] {
			queue = append(queue, child.pid)
			if !child.zombie {
				found = append(found, child.pid)
			}
		}
	}
	return found
}

// Running reports whether a process runs. A zombie, which has exited and
// waits for its parent to read its status, no longer runs.
func Running(pid int) bool {
	table, err := processes()
	if err != nil {
		return false
	}
	return slices.ContainsFunc(table, func(p process) bool { return p.pid == pid && !p.zombie })
}
