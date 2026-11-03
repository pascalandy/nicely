package extensiontest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// build makes the programs that the scenarios run. ncly is the one that go
// tool ncly builds from the Nicely version of go.mod. ncly-<slug> comes
// from cmd/ncly-<slug>/, the fake child from fake/, and the guard from
// guard/.
func (k *kit) build(t *testing.T) {
	dir := t.TempDir()
	k.bin = filepath.Join(dir, "bin")
	if err := os.Mkdir(k.bin, 0o755); err != nil {
		t.Fatal(err)
	}
	host, err := goCommand(k.root, "tool", "-n", "ncly")
	if err == nil {
		err = copyProgram(host, filepath.Join(k.bin, "ncly"))
	}
	if err != nil {
		t.Fatalf("go.mod: go tool ncly fails: %v. Fix: run go get -tool github.com/pascalandy/nicely/cmd/ncly", err)
	}
	main := "./cmd/ncly-" + k.manifest.Slug()
	k.built = filepath.Join(dir, "ncly-"+k.manifest.Slug())
	if _, err := goCommand(k.root, "build", "-o", k.built, main); err != nil {
		t.Fatalf("%s: go build fails: %v. Fix: give %s a main.go whose main calls extension.Main", main, err, main)
	}
	k.fake = filepath.Join(dir, "fake")
	if _, err := goCommand(k.root, "build", "-o", k.fake, "github.com/pascalandy/nicely/sdk/extensiontest/fake"); err != nil {
		t.Fatalf("go build of the fake child fails: %v", err)
	}
	// Only Linux runs the guard, which refuses an operation for one call.
	if runtime.GOOS == "linux" {
		k.guard = filepath.Join(dir, "guard")
		if _, err := goCommand(k.root, "build", "-o", k.guard, "github.com/pascalandy/nicely/sdk/extensiontest/guard"); err != nil {
			t.Fatalf("go build of the guard fails: %v", err)
		}
	}
}

// goCommand runs the go command in dir and returns its stdout without the
// final newline, or an error that holds its stderr.
func goCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(string(out)), nil
}

// copyProgram copies the program from to the new file to.
func copyProgram(from, to string) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	target, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(target, source)
	if closeErr := target.Close(); err == nil {
		err = closeErr
	}
	return err
}
