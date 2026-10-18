package command

import (
	"os"
	"sync"
	"testing"
)

// Err reads the first write error while a failed write records it, as
// when a handler that outlives its time writes while the extension answers
// (#47). just race runs this test under the race detector, which fails it
// when the mutex of Output is gone (#52). Each round uses a new Output, so
// the detector sees each write beside the read that races with it.
func TestOutputErrDuringFailedWrites(t *testing.T) {
	// A file opened read-only refuses every write.
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	for range 100 {
		out := NewOutput(file)
		var wg sync.WaitGroup
		wg.Go(func() { _, _ = out.Write([]byte("lost")) })
		wg.Go(func() { _ = out.Err() })
		wg.Wait()
		if out.Err() == nil {
			t.Fatal("Err reported no error after a write that failed")
		}
	}
}
