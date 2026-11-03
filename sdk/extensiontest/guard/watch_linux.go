package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

// A call says how to read one system call: the operation it makes, and the
// arguments of its old and new folders and paths. A folder of -1 is the
// current folder of the process, and flagsIndex is -1 without flags.
type call struct {
	op                         string
	oldDir, oldPath, dir, path int
	flagsIndex                 int
}

// notification and response are struct seccomp_notif and struct
// seccomp_notif_resp of linux/seccomp.h.
type notification struct {
	id    uint64
	pid   uint32
	flags uint32
	nr    int32
	arch  uint32
	ip    uint64
	args  [6]uint64
}

type response struct {
	id    uint64
	val   int64
	error int32
	flags uint32
}

// refusals is the error of each refused operation, as a file system without
// that ability gives it. vfat has no hard link. A file system without
// RENAME_NOREPLACE refuses the flag.
var refusals = map[string]unix.Errno{"link": unix.EPERM, "rename-noreplace": unix.EINVAL, "rename": unix.EPERM}

// watcher answers the notifications of the filter.
type watcher struct {
	listener      int
	scope, refuse string
	skip          []string
	done          chan struct{}
	mu            sync.Mutex
	seen          []Operation
}

func (w *watcher) operations() []Operation {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]Operation{}, w.seen...)
}

// answer serves notifications until done closes. Each notification blocks
// its process until the guard answers.
func (w *watcher) answer() {
	for {
		select {
		case <-w.done:
			return
		default:
		}
		fds := []unix.PollFd{{Fd: int32(w.listener), Events: unix.POLLIN}}
		if n, err := unix.Poll(fds, 50); err != nil || n == 0 || fds[0].Revents&unix.POLLIN == 0 {
			if fds[0].Revents&unix.POLLHUP != 0 {
				// No process uses the filter any more.
				<-w.done
				return
			}
			continue
		}
		var req notification
		if err := ioctl(w.listener, unix.SECCOMP_IOCTL_NOTIF_RECV, unsafe.Pointer(&req)); err != nil {
			continue
		}
		w.handle(req)
	}
}

// handle answers one notification. The kernel runs a call outside the
// scope. In the scope, the guard refuses the operation of -refuse, and
// makes any other itself, so the report says whether it succeeded.
func (w *watcher) handle(req notification) {
	resp := response{id: req.id, flags: unix.SECCOMP_USER_NOTIF_FLAG_CONTINUE}
	c, known := calls[uint32(req.nr)]
	var op Operation
	var from, to string
	var flags uint64
	if known {
		op.Name = c.op
		if c.flagsIndex >= 0 {
			flags = req.args[c.flagsIndex]
		}
		if c.op == "rename" && flags&unix.RENAME_NOREPLACE != 0 {
			op.Name = "rename-noreplace"
		}
		from, to = path(int(req.pid), c.oldDir, req.args[c.oldPath], req.args), path(int(req.pid), c.dir, req.args[c.path], req.args)
		op.Target = real(filepath.Clean(to))
	}
	// The process may have exited, and its memory gone to another, while the
	// guard read it.
	if ioctl(w.listener, unix.SECCOMP_IOCTL_NOTIF_ID_VALID, unsafe.Pointer(&req.id)) != nil {
		return
	}
	inScope := to != "" && within(op.Target, w.scope) && !slices.ContainsFunc(w.skip, func(dir string) bool { return within(op.Target, dir) })
	switch {
	case !inScope || from == "" || c.op == "link" && flags&unix.AT_EMPTY_PATH != 0:
	case op.Name == w.refuse:
		op.Refused = true
		resp.flags, resp.error = 0, -int32(refusals[op.Name])
	default:
		// A rename of a file onto itself succeeds and changes nothing, so it
		// publishes nothing.
		same := sameFile(from, to)
		err := perform(c.op, from, to, flags)
		resp.flags = 0
		var errno unix.Errno
		if errors.As(err, &errno) {
			resp.error = -int32(errno)
		} else if err != nil {
			resp.error = -int32(unix.EIO)
		}
		op.Done = err == nil && !same
	}
	_ = ioctl(w.listener, unix.SECCOMP_IOCTL_NOTIF_SEND, unsafe.Pointer(&resp))
	if inScope {
		w.mu.Lock()
		w.seen = append(w.seen, op)
		w.mu.Unlock()
	}
}

// perform makes a link or a rename for a process, with the paths that the
// guard resolved from its folders.
func perform(op, from, to string, flags uint64) error {
	if op == "link" {
		return unix.Linkat(unix.AT_FDCWD, from, unix.AT_FDCWD, to, int(flags&unix.AT_SYMLINK_FOLLOW))
	}
	return unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, uint(flags))
}

// sameFile reports whether two paths name one file, without a link
// followed.
func sameFile(a, b string) bool {
	x, errA := os.Lstat(a)
	y, errB := os.Lstat(b)
	return errA == nil && errB == nil && os.SameFile(x, y)
}

func ioctl(fd int, request uint, arg unsafe.Pointer) error {
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(request), uintptr(arg)); errno != 0 {
		return errno
	}
	return nil
}

// path returns the absolute path that a call names at addr. A relative path
// starts at the folder of the argument dir. It returns "" when the guard
// cannot read the path.
func path(pid, dir int, addr uint64, args [6]uint64) string {
	name, err := readString(pid, uintptr(addr))
	if err != nil || name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		return name
	}
	base := fmt.Sprintf("/proc/%d/cwd", pid)
	if dir >= 0 {
		if dirfd := int32(uint32(args[dir])); dirfd != unix.AT_FDCWD {
			base = fmt.Sprintf("/proc/%d/fd/%d", pid, dirfd)
		}
	}
	folder, err := os.Readlink(base)
	if err != nil || !filepath.IsAbs(folder) {
		return ""
	}
	return folder + string(filepath.Separator) + name
}

// readString reads a string that ends with NUL in the memory of pid, page
// by page, because a read past a mapped page fails.
func readString(pid int, addr uintptr) (string, error) {
	mem, err := os.Open(fmt.Sprintf("/proc/%d/mem", pid))
	if err != nil {
		return "", err
	}
	defer func() { _ = mem.Close() }()
	page := uintptr(os.Getpagesize())
	var text []byte
	for len(text) < unix.PathMax {
		chunk := make([]byte, page-addr%page)
		n, err := mem.ReadAt(chunk, int64(addr))
		if i := slices.Index(chunk[:n], 0); i >= 0 {
			return string(append(text, chunk[:i]...)), nil
		}
		if err != nil {
			return "", err
		}
		text, addr = append(text, chunk[:n]...), addr+uintptr(n)
	}
	return "", errors.New("the path is too long")
}

// real returns path with the symbolic links of its folder resolved, as
// /proc gives the folder of a process. A link at path itself stays.
func real(path string) string {
	if parent, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
		return filepath.Join(parent, filepath.Base(path))
	}
	return filepath.Clean(path)
}

// folder returns dir with all its symbolic links resolved.
func folder(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return real(dir)
}

// within reports whether path is dir or lies inside it.
func within(path, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+string(filepath.Separator))
}
