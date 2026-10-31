package extensiontest

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/rogpeppe/go-internal/testscript"
)

// appearsText is what the kit writes at the path. A run that replaces the file cannot keep its inode.
const appearsText = "Another process wrote this file during the run. The run must keep it.\n"

// appears proves that a run without --force never replaces a file that appears during the run.
// At the marker, the kit creates the path with an exclusive create. The run must then fail and keep that file.
func (c *Coverage) appears(ts *testscript.TestScript, neg bool, args, protected []string, registeredHost string) {
	if neg || len(args) < 3 {
		ts.Fatalf("usage: appears <marker> <path> <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, call := c.pairedCall(ts, args[2], args[3:], registeredHost)
	if _, given := call.line.Switch("force"); given || slices.ContainsFunc(args[3:], func(arg string) bool {
		return arg == "--force" || strings.HasPrefix(arg, "--force=")
	}) {
		ts.Fatalf("appears: remove --force, because appears proves a run without it")
	}
	marker, path := ts.MkAbs(args[0]), ts.MkAbs(args[1])
	for i, file := range []string{marker, path} {
		if _, err := os.Lstat(file); !os.IsNotExist(err) {
			ts.Fatalf("appears: %s must not exist before the call", args[i])
		}
	}
	before, err := snapshotFiles(protected, false)
	ts.Check(err)
	var made os.FileInfo
	exit, stdout, stderr := foreground(ts, "appears", program, args[3:], func(done <-chan struct{}, host *os.Process) error {
		if err := waitMarker("appears", done, host, marker); err != nil {
			return err
		}
		info, err := createNew(path)
		if err != nil {
			kill(host)
			return fmt.Errorf("appears: %s exists at the marker, or the kit cannot create it: %w", args[1], err)
		}
		made = info
		if !waitExit(done, host, 10*time.Second+runner.DrainLimit) {
			return fmt.Errorf("appears: host did not exit within 10s after %s appeared", args[1])
		}
		return nil
	})
	if _, problem := c.checkedCode(words, exit, stdout, stderr); problem != "" {
		ts.Fatalf("appears: %s", problem)
	}
	if exit == 0 {
		ts.Fatalf("appears: %s exited with 0, want a failure that keeps %s", args[2], args[1])
	}
	ts.Check(kept(args[1], path, made))
	same, err := unchanged(protected, before, marker, path)
	ts.Check(err)
	if !same {
		ts.Fatalf("appears: the run changed protected paths: %s", fmt.Sprint(protected))
	}
}

// createNew creates path with an exclusive create and writes appearsText.
// It records the file before the write, because the host can act as soon as the path exists.
func createNew(path string) (os.FileInfo, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err == nil {
		_, err = file.WriteString(appearsText)
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return info, err
}

// kept requires the same regular file at path, with the same mode and bytes.
func kept(name, path string, made os.FileInfo) error {
	info, err := os.Lstat(path)
	switch {
	case err != nil:
		return fmt.Errorf("appears: the run removed %s: %w", name, err)
	case !info.Mode().IsRegular():
		return fmt.Errorf("appears: %s is no longer the regular file that the kit created", name)
	case !os.SameFile(made, info):
		return fmt.Errorf("appears: the run replaced %s", name)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(data) != appearsText || info.Mode() != made.Mode() {
		return fmt.Errorf("appears: the run changed %s", name)
	}
	return nil
}
