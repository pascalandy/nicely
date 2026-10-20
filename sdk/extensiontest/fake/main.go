// Command fake is the fake child of the test kit. A scenario puts it in place
// of a program that the extension runs, with the scenario command fake
// <name>. The kit builds it from this source, because a Go module zip keeps
// no executable bit.
//
// It runs the actions of the file <name>.fake beside it, one per line, in
// order. A blank line and a line that starts with # do nothing.
//
//	stdout <text>    writes text and a newline to stdout
//	stderr <text>    writes text and a newline to stderr
//	stdin            copies stdin to stdout
//	env              writes the name of each variable of its environment, sorted, one per line, never a value
//	sleep <duration> waits, such as sleep 30s, until a signal ends it
//	exit <code>      exits with code
//
// It exits with 0 after the last action, and with 127 when its file is
// missing or holds an unknown action.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	self, err := os.Executable()
	if err == nil {
		err = run(self+".fake", os.Stdin, os.Stdout, os.Stderr)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "fake:", err)
		os.Exit(127)
	}
}

func run(path string, stdin io.Reader, stdout, stderr io.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	lines := bufio.NewScanner(file)
	for n := 1; lines.Scan(); n++ {
		line := strings.TrimSpace(lines.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		action, value, _ := strings.Cut(line, " ")
		switch action {
		case "stdout":
			_, err = fmt.Fprintln(stdout, value)
		case "stderr":
			_, err = fmt.Fprintln(stderr, value)
		case "stdin":
			_, err = io.Copy(stdout, stdin)
		case "env":
			var names []string
			for _, entry := range os.Environ() {
				name, _, _ := strings.Cut(entry, "=")
				names = append(names, name)
			}
			slices.Sort(names)
			_, err = fmt.Fprintln(stdout, strings.Join(names, "\n"))
		case "sleep":
			var d time.Duration
			if d, err = time.ParseDuration(value); err == nil {
				time.Sleep(d)
			}
		case "exit":
			var code int
			if code, err = strconv.Atoi(value); err == nil {
				os.Exit(code)
			}
		default:
			err = fmt.Errorf("unknown action %q", action)
		}
		if err != nil {
			return fmt.Errorf("%s:%d: %w", path, n, err)
		}
	}
	return lines.Err()
}
