package extensiontest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// goTest copies the fixture, points its replace at this checkout,
// applies the edits, and runs its go test with path as PATH.
func goTest(t *testing.T, fixture, path string, edits ...edit) (string, error) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join(root, "testdata", "extensions", fixture))); err != nil {
		t.Fatal(err)
	}
	if out, err := command(dir, os.Getenv("PATH"), "go", "mod", "edit", "-replace", "github.com/pascalandy/nicely="+root); err != nil {
		t.Fatalf("go mod edit: %v\n%s", err, out)
	}
	for _, e := range edits {
		file := filepath.Join(dir, e.file)
		var err error
		switch data, readErr := os.ReadFile(file); {
		case e.old == "" && e.new == "":
			err = os.Remove(file)
		case e.old == "":
			err = os.WriteFile(file, []byte(e.new), 0o644)
		case readErr != nil || !strings.Contains(string(data), e.old):
			t.Fatalf("%s lacks %q: %v", e.file, e.old, readErr)
		default:
			err = os.WriteFile(file, []byte(strings.Replace(string(data), e.old, e.new, 1)), 0o644)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return command(dir, path, "go", "test", "-count=1", "./...")
}

// command runs a program in dir with path as PATH, and returns its output.
func command(dir, path, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
