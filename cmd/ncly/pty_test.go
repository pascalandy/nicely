package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

// ptyRun runs a command with stdin, stdout, and stderr on a pseudo-terminal
// of 80 columns and 24 rows, as a human at a terminal runs it:
// pty <script> <command> [args...]. Each line of the script holds a prompt
// and an answer, separated by a tab. pty waits until the terminal shows the
// prompt, then types the answer and Enter, or one control key alone when the
// answer is a caret and a letter: Ctrl-C for ^C, Ctrl-D for ^D, the end of
// input, Ctrl-E for ^E, which accepts a suggestion, Enter for ^M, and
// Escape for ^[. An answer that ends with \c, as in echo, is
// typed without Enter, so the next prompt can be what the typing shows.
// Once the command ends, pty prints what the terminal showed on stdout, with
// each CR LF as LF, and exits with the command's code. On stderr, it then
// says echo on or echo off, the state that the command left the terminal in,
// and terminal as found, or terminal changed and what changed: the settings
// of the terminal, or each mode that the command changed and left changed,
// such as a hidden cursor.
//
// testscript's ttyin leaves stdout a pipe, so the call would be
// non-interactive, and script writes its input before ncly turns echo off,
// so a typed key would show.
func ptyRun() {
	os.Exit(runPty(os.Args[1:]))
}

func runPty(args []string) int {
	fail := func(err error) int {
		fmt.Fprintln(os.Stderr, "pty:", err)
		return 1
	}
	if len(args) < 2 {
		return fail(errors.New("usage: pty <script> <command> [args...]"))
	}
	steps, err := ptyScript(args[0])
	if err != nil {
		return fail(err)
	}
	ptmx, tty, err := pty.Open()
	if err != nil {
		return fail(err)
	}
	defer func() { _ = ptmx.Close() }()
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		return fail(err)
	}
	found, err := unix.IoctlGetTermios(int(tty.Fd()), ioctlGetTermios)
	if err != nil {
		return fail(err)
	}
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	// The command leads its own session, with the terminal as its
	// controlling one, so Ctrl-C reaches it as SIGINT.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	screen := &ptyScreen{changed: make(chan struct{}, 1)}
	go screen.read(ptmx)
	for _, s := range steps {
		if !screen.waitFor(s.prompt, 10*time.Second) {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			fmt.Print(screen.text())
			return fail(fmt.Errorf("the prompt %q never appeared", s.prompt))
		}
		answer, typed := strings.CutSuffix(s.answer, `\c`)
		switch {
		case len(s.answer) == 2 && s.answer[0] == '^' && '@' <= s.answer[1] && s.answer[1] <= '_':
			answer = string(rune(s.answer[1] - '@'))
		case !typed:
			answer += "\r"
		}
		if _, err := ptmx.WriteString(answer); err != nil {
			return fail(err)
		}
	}
	err = cmd.Wait()
	code := 0
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		code = exit.ExitCode()
		if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			code = 128 + int(status.Signal())
		}
	case err != nil:
		return fail(err)
	}
	left, err := unix.IoctlGetTermios(int(settingsSide(ptmx, tty).Fd()), ioctlGetTermios)
	if err != nil {
		return fail(err)
	}
	echo := "echo off"
	if left.Lflag&unix.ECHO != 0 {
		echo = "echo on"
	}
	// Once the terminal closes, the reader sees the end of what it showed.
	_ = tty.Close()
	screen.waitDone(2 * time.Second)
	text := screen.text()
	fmt.Print(text)
	fmt.Fprintln(os.Stderr, echo)
	changed := changedModes(text)
	if *left != *found {
		changed = append([]string{"settings"}, changed...)
	}
	if len(changed) == 0 {
		fmt.Fprintln(os.Stderr, "terminal as found")
	} else {
		fmt.Fprintln(os.Stderr, "terminal changed:", strings.Join(changed, ", "))
	}
	return code
}

// privateMode is a sequence that sets or resets modes of a terminal, such as
// ESC [ ? 25 l, which hides the cursor.
var privateMode = regexp.MustCompile(`\x1b\[\?([0-9;]+)([hl])`)

// changedModes lists each mode that what the terminal showed changed and
// left changed. A program changes a mode from the state that it found, so
// the mode is as found when its last change undoes its first one.
func changedModes(text string) []string {
	first, last := map[string]string{}, map[string]string{}
	for _, m := range privateMode.FindAllStringSubmatch(text, -1) {
		for mode := range strings.SplitSeq(m[1], ";") {
			if _, seen := first[mode]; !seen {
				first[mode] = m[2]
			}
			last[mode] = m[2]
		}
	}
	var changed []string
	for mode, set := range first {
		if last[mode] == set {
			changed = append(changed, "mode "+mode)
		}
	}
	slices.Sort(changed)
	return changed
}

type ptyStep struct{ prompt, answer string }

func ptyScript(path string) ([]ptyStep, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var steps []ptyStep
	lines := bufio.NewScanner(bytes.NewReader(data))
	for lines.Scan() {
		prompt, answer, found := strings.Cut(lines.Text(), "\t")
		if !found {
			return nil, fmt.Errorf("%s: want a prompt, a tab, and an answer on %q", path, lines.Text())
		}
		steps = append(steps, ptyStep{prompt, answer})
	}
	return steps, lines.Err()
}

// ptyScreen collects what the terminal shows.
type ptyScreen struct {
	mu      sync.Mutex
	shown   []byte
	seen    int
	done    bool
	changed chan struct{}
}

func (s *ptyScreen) read(ptmx *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := ptmx.Read(buf)
		s.mu.Lock()
		s.shown = append(s.shown, buf[:n]...)
		s.done = err != nil
		s.mu.Unlock()
		select {
		case s.changed <- struct{}{}:
		default:
		}
		if err != nil {
			return
		}
	}
}

// waitFor waits until the terminal shows prompt after what earlier prompts
// matched.
func (s *ptyScreen) waitFor(prompt string, limit time.Duration) bool {
	deadline := time.After(limit)
	for {
		s.mu.Lock()
		i := bytes.Index(s.shown[s.seen:], []byte(prompt))
		if i >= 0 {
			s.seen += i + len(prompt)
		}
		done := s.done
		s.mu.Unlock()
		if i >= 0 {
			return true
		}
		if done {
			return false
		}
		select {
		case <-s.changed:
		case <-deadline:
			return false
		}
	}
}

func (s *ptyScreen) waitDone(limit time.Duration) {
	deadline := time.After(limit)
	for {
		s.mu.Lock()
		done := s.done
		s.mu.Unlock()
		if done {
			return
		}
		select {
		case <-s.changed:
		case <-deadline:
			return
		}
	}
}

func (s *ptyScreen) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.ReplaceAll(string(s.shown), "\r\n", "\n")
}
