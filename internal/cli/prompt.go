package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

// The plain prompts of ncly auth, until the masked field and the
// confirmation form of M07 replace them. Each one reads from stdin, a
// terminal, and writes to stdout, the terminal too, since interactive mode
// needs both. A signal ends the wait at once: the read that it abandons
// writes nothing, and the process ends right after.

// readSecret shows the prompt with echo off, so the key never shows, and
// returns the line that the human types. It turns echo off before it shows
// the prompt, and restores the terminal's state when the prompt ends, a
// signal included.
func readSecret(ctx context.Context, in *os.File, out io.Writer, prompt string) (string, error) {
	fd := int(in.Fd())
	state, err := unix.IoctlGetTermios(fd, getTermios)
	if err != nil {
		return "", err
	}
	quiet := *state
	quiet.Lflag &^= unix.ECHO
	quiet.Lflag |= unix.ICANON | unix.ISIG
	quiet.Iflag |= unix.ICRNL
	if err := unix.IoctlSetTermios(fd, setTermios, &quiet); err != nil {
		return "", err
	}
	defer func() { _ = unix.IoctlSetTermios(fd, setTermios, state) }()
	if _, err := io.WriteString(out, prompt); err != nil {
		return "", err
	}
	line, err := readLine(ctx, in)
	// Enter never shows with echo off, so the prompt ends its own line.
	_, _ = io.WriteString(out, "\n")
	return line, err
}

// confirm asks a y/N question and reports whether the answer is one of the
// words that mean yes. Any other answer, an empty one, and the end of input
// mean no.
func confirm(ctx context.Context, in *os.File, out io.Writer, question string, yes []string) (bool, error) {
	if _, err := io.WriteString(out, question); err != nil {
		return false, err
	}
	line, err := readLine(ctx, in)
	if errors.Is(err, io.EOF) {
		_, _ = io.WriteString(out, "\n")
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return slices.Contains(yes, strings.ToLower(strings.TrimSpace(line))), nil
}

// readLine reads one line from in without its line ending, or returns the
// error of ctx when ctx ends first.
func readLine(ctx context.Context, in io.Reader) (string, error) {
	type answer struct {
		line string
		err  error
	}
	done := make(chan answer, 1)
	go func() {
		line, err := bufio.NewReader(in).ReadString('\n')
		if errors.Is(err, io.EOF) && line != "" {
			err = nil
		}
		done <- answer{strings.TrimRight(line, "\r\n"), err}
	}()
	select {
	case a := <-done:
		return a.line, a.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// readAll reads in to its end, or returns the error of ctx when ctx ends
// first.
func readAll(ctx context.Context, in io.Reader) (string, error) {
	type answer struct {
		data []byte
		err  error
	}
	done := make(chan answer, 1)
	go func() {
		data, err := io.ReadAll(in)
		done <- answer{data, err}
	}()
	select {
	case a := <-done:
		return string(a.data), a.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// yesWords splits the catalog's comma-separated words that confirm.
func yesWords(list string) []string {
	var words []string
	for word := range strings.SplitSeq(list, ",") {
		if word = strings.ToLower(strings.TrimSpace(word)); word != "" {
			words = append(words, word)
		}
	}
	if len(words) == 0 {
		panic(fmt.Sprintf("cli: no word confirms in %q", list))
	}
	return words
}
