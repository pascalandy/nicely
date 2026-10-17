package cli

import (
	"context"
	"io"
)

// readAll reads in to its end, or returns the error of ctx when ctx ends
// first. A signal ends the wait at once: the read that it abandons writes
// nothing, and the process ends right after.
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
