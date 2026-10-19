package runner

import (
	"io"
	"os"
	"os/exec"
	"time"
)

// feed connects r to what the program reads: a file as it is, and
// otherwise a pipe, and returns the function that starts filling that pipe
// from r once the program has started. Run never waits for that copy,
// because a Read from r may block until its owner closes r, as the
// documentation of exec.Cmd warns of its own copy. Wait closes the pipe
// once the program exits.
func feed(cmd *exec.Cmd, r io.Reader) (func(), error) {
	if _, isFile := r.(*os.File); isFile || r == nil {
		cmd.Stdin = r
		return func() {}, nil
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	return func() {
		go func() {
			_, _ = io.Copy(w, r)
			_ = w.Close()
		}()
	}, nil
}

// stream copies what a program writes on a pipe that the runner owns, so
// that the runner, not a descendant that holds the pipe, decides when the
// copy ends.
type stream struct {
	to   io.Writer
	r, w *os.File
	done chan struct{}
}

// attach returns what the program writes to: a file as it is, so that the
// program sees a terminal when it is one, and otherwise the write end of a
// pipe whose copy to w the runner controls. A writer that wraps a file, such
// as the stdout of an SDK handler, passes its file. Two streams to the same
// writer share one pipe, as exec.Cmd does, so their writes never overlap.
func attach(w io.Writer, streams *[]*stream) (io.Writer, error) {
	if wrapper, wraps := w.(interface{ File() *os.File }); wraps {
		w = wrapper.File()
	}
	if _, isFile := w.(*os.File); isFile || w == nil {
		return w, nil
	}
	for _, s := range *streams {
		if sameWriter(s.to, w) {
			return s.w, nil
		}
	}
	r, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	s := &stream{to: w, r: r, w: pw, done: make(chan struct{})}
	go func() {
		_, _ = io.Copy(w, r)
		close(s.done)
	}()
	*streams = append(*streams, s)
	return pw, nil
}

// sameWriter reports whether a and b are the same writer. Comparing two
// values of a type that is not comparable panics, and then they differ.
func sameWriter(a, b io.Writer) (same bool) {
	defer func() { _ = recover() }()
	return a == b
}

// drain lets each copy run until its stream ends or until the deadline,
// then stops it. Output written after the deadline is lost.
func drain(streams []*stream, deadline time.Time) {
	for _, s := range streams {
		select {
		case <-s.done:
		case <-time.After(time.Until(deadline)):
		}
		_ = s.r.Close()
		<-s.done
	}
}
