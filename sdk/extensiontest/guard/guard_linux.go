package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/pascalandy/nicely/sdk/platform"
	"golang.org/x/sys/unix"
)

// settle is how long a descendant may run after the call exits, such as a
// child that the call killed just before its exit.
const settle = 2 * time.Second

// supervise runs the call through the stage, which installs the filter and
// sends its listener here. It answers each notification until the call and
// its descendants have ended.
func supervise(o options) Report {
	r := Report{Operations: []Operation{}, Children: []string{}}
	if err := platform.BecomeSubreaper(); err != nil {
		r.Error = "subreaper: " + err.Error()
		return r
	}
	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		r.Error = "socketpair: " + err.Error()
		return r
	}
	defer func() { _ = unix.Close(pair[0]) }()
	child := os.NewFile(uintptr(pair[1]), "guard-stage")
	self, err := os.Executable()
	if err != nil {
		r.Error = err.Error()
		return r
	}
	cmd := exec.Command(self, append([]string{"-stage", "--"}, o.call...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.ExtraFiles = []*os.File{child}
	err = cmd.Start()
	_ = child.Close()
	if err != nil {
		r.Error = "start: " + err.Error()
		return r
	}
	listener, err := receive(pair[0])
	if err != nil {
		_ = cmd.Wait()
		r.Error = err.Error()
		return r
	}
	w := watcher{listener: listener, scope: folder(o.scope), refuse: o.refuse, done: make(chan struct{})}
	for _, dir := range o.skip {
		w.skip = append(w.skip, folder(dir))
	}
	var answered sync.WaitGroup
	answered.Go(w.answer)
	_ = cmd.Wait()
	r.Exit = exitOf(cmd.ProcessState)
	r.Children = reap()
	close(w.done)
	answered.Wait()
	_ = unix.Close(listener)
	r.Operations = w.operations()
	return r
}

// receive returns the listener that the stage sends, or the error of the
// stage.
func receive(sock int) (int, error) {
	buf, oob := make([]byte, 4096), make([]byte, unix.CmsgSpace(4))
	n, oobn, _, _, err := unix.Recvmsg(sock, buf, oob, 0)
	if err != nil {
		return -1, fmt.Errorf("receive the listener: %w", err)
	}
	if oobn == 0 {
		if n == 0 {
			return -1, errors.New("the stage exited before it installed the filter")
		}
		return -1, errors.New(string(buf[:n]))
	}
	messages, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, fmt.Errorf("read the listener: %w", err)
	}
	var fds []int
	if len(messages) == 1 {
		fds, err = unix.ParseUnixRights(&messages[0])
	}
	if err != nil || len(fds) != 1 {
		return -1, fmt.Errorf("read the listener from %d messages: %w", len(messages), err)
	}
	return fds[0], nil
}

// stage installs the filter on its own thread, sends its listener to the
// guard through descriptor 3, and runs the call in its place. Only the call
// and its descendants get the filter.
func stage(call []string) int {
	runtime.LockOSThread()
	const sock = 3
	unix.CloseOnExec(sock)
	fail := func(format string, args ...any) int {
		_, _ = unix.Write(sock, []byte(fmt.Sprintf(format, args...)))
		return 125
	}
	if len(calls) == 0 {
		return fail("%s/%s has no filter of the guard", runtime.GOOS, runtime.GOARCH)
	}
	path, err := exec.LookPath(call[0])
	if err != nil {
		return fail("%v", err)
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fail("prctl no_new_privs: %v", err)
	}
	program := filter()
	prog := unix.SockFprog{Len: uint16(len(program)), Filter: &program[0]}
	fd, _, errno := unix.RawSyscall(unix.SYS_SECCOMP, unix.SECCOMP_SET_MODE_FILTER, unix.SECCOMP_FILTER_FLAG_NEW_LISTENER, uintptr(unsafe.Pointer(&prog)))
	if errno != 0 {
		return fail("seccomp user notification: %v", errno)
	}
	if err := unix.Sendmsg(sock, []byte{0}, unix.UnixRights(int(fd)), nil, 0); err != nil {
		return fail("send the listener: %v", err)
	}
	_ = unix.Close(int(fd))
	// The kit gives the guard the environment of the scenario, which the
	// call keeps.
	err = syscall.Exec(path, call, os.Environ())
	_, _ = fmt.Fprintln(os.Stderr, "guard: exec:", err)
	return 127
}

// filter sends each call of calls to the listener, and allows any other.
func filter() []unix.SockFilter {
	const (
		ld  = unix.BPF_LD | unix.BPF_W | unix.BPF_ABS
		jeq = unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K
		ret = unix.BPF_RET | unix.BPF_K
	)
	numbers := make([]uint32, 0, len(calls))
	for nr := range calls {
		numbers = append(numbers, nr)
	}
	slices.Sort(numbers)
	n := uint8(len(numbers))
	// The checks of numbers sit between the load of nr and the two returns.
	program := []unix.SockFilter{
		{Code: ld, K: 4}, // seccomp_data.arch
		{Code: jeq, K: arch, Jt: 0, Jf: n + 1},
		{Code: ld, K: 0}, // seccomp_data.nr
	}
	for i, nr := range numbers {
		program = append(program, unix.SockFilter{Code: jeq, K: nr, Jt: n - uint8(i), Jf: 0})
	}
	return append(program,
		unix.SockFilter{Code: ret, K: unix.SECCOMP_RET_ALLOW},
		unix.SockFilter{Code: ret, K: unix.SECCOMP_RET_USER_NOTIF})
}

// exitOf returns the exit code of the call, as a shell reports it.
func exitOf(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}

// reap waits for each descendant that the guard adopted, up to settle. It
// returns each one that still runs then, with its command, and kills it.
func reap() []string {
	found := []string{}
	deadline := time.Now().Add(settle)
	for rounds := 0; rounds < 5; {
		var status unix.WaitStatus
		pid, err := unix.Wait4(-1, &status, unix.WNOHANG, nil)
		switch {
		case errors.Is(err, unix.ECHILD):
			return found
		case err == nil && pid > 0:
			continue
		case time.Now().Before(deadline):
			time.Sleep(10 * time.Millisecond)
			continue
		}
		for _, pid := range platform.Descendants(os.Getpid()) {
			command, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			entry := fmt.Sprintf("%d %s", pid, strings.TrimSpace(strings.ReplaceAll(string(command), "\x00", " ")))
			if !slices.Contains(found, entry) {
				found = append(found, entry)
			}
			_ = unix.Kill(pid, unix.SIGKILL)
		}
		// Each kill makes more orphans, which the guard adopts and reaps.
		deadline = time.Now().Add(settle)
		rounds++
	}
	return found
}
