package extensiontest

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/rogpeppe/go-internal/testscript"
)

// holds proves the timeline of a publication: holds <marker> <path>
// <command> <arg>... The worker writes part of its output in a stage,
// writes the marker, then waits until the marker is gone. While the worker
// waits, path must keep its state before the call: absent, or the same
// file with the same bytes. Then the kit removes the marker, and the run
// must succeed and publish path. On Linux, the guard observes the commit:
// an operation of the row of the mode must publish path.
func (k *kit) holds(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 3 {
		ts.Fatalf("usage: holds <marker> <path> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, call := k.pairedCall(ts, args[2], args[3:], filepath.Join(k.bin, "ncly"))
	mode := modeOf(call)
	row, found := k.publicationOf(call.path, mode)
	if !found {
		ts.Fatalf("holds: ## Publication of spec.md declares no %s mode of `ncly %s`. Fix: give that mode a row", mode, call.path)
	}
	marker, path := ts.MkAbs(args[0]), ts.MkAbs(args[1])
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		ts.Fatalf("holds: marker %s must not exist before the call", args[0])
	}
	protected := slices.Concat(k.takeNext(ts), protectedPaths(ts, call, row.protected))
	before, err := snapshotFiles(protected, false)
	ts.Check(err)
	public, err := publicStateOf(path)
	ts.Check(err)
	g := k.guarded(ts, "holds", true)
	guard, guarded := g.wrap(program, args[3:])
	exit, stdout, stderr := foreground(ts, "holds", guard, guarded, func(done <-chan struct{}, host *os.Process) error {
		if err := waitMarker("holds", done, host, marker); err != nil {
			return err
		}
		held, err := publicStateOf(path)
		if err != nil || !held.same(public) {
			kill(host)
			return fmt.Errorf("holds: while the worker waits, %s holds work of the run, so a reader can see a partial file. Fix: write the work in a private stage, and publish it with an operation of its row: %s", args[1], strings.Join(row.operations, ", "))
		}
		if err := os.Remove(marker); err != nil {
			kill(host)
			return fmt.Errorf("holds: release the worker: %w", err)
		}
		if !waitExit(done, host, 10*time.Second+runner.DrainLimit) {
			return fmt.Errorf("holds: host did not exit within 10s after the kit removed %s", args[0])
		}
		return nil
	})
	report := g.read(ts, "holds")
	if _, problem := k.checkedCode(words, exit, stdout, stderr); problem != "" {
		ts.Fatalf("holds: %s", problem)
	}
	if exit != 0 {
		ts.Fatalf("holds: %s exited with %d after the kit released the worker, want 0", args[2], exit)
	}
	if published, _ := decodeAnswer(stdout)[row.destination].(string); published == "" || filepath.Clean(ts.MkAbs(published)) != filepath.Clean(path) {
		ts.Fatalf("holds: the run published at %q, want %s", published, args[1])
	}
	if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
		ts.Fatalf("holds: the run left no regular file at %s", args[1])
	}
	if g != nil {
		checkCommit(ts, "holds", report, row, args[1])
	}
	same, err := unchanged(protected, before, marker, path)
	ts.Check(err)
	if !same {
		ts.Fatalf("holds: the run changed protected paths: %s", strings.Join(protected, ", "))
	}
}

// A publicState is what a reader of the public path sees: no file, or one
// file with its identity, mode, and bytes.
type publicState struct {
	info  os.FileInfo
	files map[string]fileState
}

func publicStateOf(path string) (publicState, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return publicState{}, nil
	}
	if err != nil {
		return publicState{}, err
	}
	files, err := snapshotFiles([]string{path}, false)
	return publicState{info: info, files: files}, err
}

// same reports whether a reader sees the same state.
func (s publicState) same(other publicState) bool {
	if s.info == nil || other.info == nil {
		return s.info == nil && other.info == nil
	}
	return os.SameFile(s.info, other.info) && maps.Equal(s.files, other.files)
}
