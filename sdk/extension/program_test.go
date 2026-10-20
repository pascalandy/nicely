package extension

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// TestWorkerExtractsIntoAPrivateFolderUntilTheRunEnds proves that the first
// Worker of a run writes every worker into one folder of TMPDIR. Only the
// user can read that folder. No worker comes back once the run removed it.
func TestWorkerExtractsIntoAPrivateFolderUntilTheRunEnds(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	catalogs, err := i18n.Load(i18n.Locales())
	if err != nil {
		t.Fatal(err)
	}
	resources := fstest.MapFS{
		"worker/count.sh":  {Data: []byte("echo count\n")},
		"worker/lib/io.sh": {Data: []byte("echo io\n")},
		"other.txt":        {Data: []byte("not a worker\n")},
	}
	w := &workers{resources: resources, domain: "demo"}
	call := Call{Printer: catalogs.Printer("en"), workers: w}
	count, problem := call.Worker("count.sh")
	if problem != nil {
		t.Fatal(problem.Message)
	}
	dir := filepath.Dir(count)
	if filepath.Dir(dir) != tmp || !strings.HasPrefix(filepath.Base(dir), "ncly-demo-") {
		t.Fatalf("the worker is at %s, want a folder ncly-demo-* of %s", count, tmp)
	}
	if info, err := os.Stat(dir); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("the folder of the workers has the mode %v and %v, want 0700", info.Mode(), err)
	}
	if data, err := os.ReadFile(count); err != nil || string(data) != "echo count\n" {
		t.Fatalf("the worker holds %q and %v", data, err)
	}
	if io, problem := call.Worker("lib/io.sh"); problem != nil || io != filepath.Join(dir, "lib", "io.sh") {
		t.Fatalf("the second worker is at %s, want it beside the first", io)
	}
	for _, name := range []string{"missing.sh", "../other.txt", "/etc/passwd", "lib"} {
		if _, problem := call.Worker(name); problem == nil || problem.Code != contract.Runtime || !strings.Contains(problem.Message, "no worker named") {
			t.Errorf("Worker(%q) returned %+v, want worker.missing", name, problem)
		}
	}
	w.remove()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("the folder of the workers remains: %v", err)
	}
	if _, problem := call.Worker("count.sh"); problem == nil || problem.Code != contract.Runtime {
		t.Fatalf("Worker after the run returned %+v, want RUNTIME", problem)
	}
	if entries, _ := os.ReadDir(tmp); len(entries) != 0 {
		t.Fatalf("TMPDIR holds %v after the run", entries)
	}
	none := Call{Printer: catalogs.Printer("en"), workers: &workers{domain: "demo"}}
	if _, problem := none.Worker("count.sh"); problem == nil {
		t.Fatal("an extension without Resources returned a worker")
	}
}
