package extension

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/pascalandy/nicely/sdk/contract"
)

// Requirement returns the entry in programs of a program that the command
// declares in requires.bins. It returns false for any other program.
// sdk/wrap reads it to check a program as doctor does.
func (c Call) Requirement(program string) (contract.Requirement, bool) {
	r, declared := c.programs[program]
	return r, declared
}

// Worker returns the absolute path of a worker that Resources holds, such
// as "count.sh" for worker/count.sh, or the problem that stops the call.
// The first call of a run extracts every worker into a private temporary
// folder. The SDK removes that folder once the handler returns.
func (c Call) Worker(name string) (string, *contract.Problem) {
	p := c.Printer
	data := map[string]any{"Name": name}
	if c.workers == nil || !c.workers.holds(name) {
		return "", &contract.Problem{Code: contract.Runtime, Message: p.T("worker.missing", data), Hint: p.T("outcome.extension_broken_hint")}
	}
	dir, err := c.workers.extract()
	if err != nil {
		return "", &contract.Problem{Code: contract.Runtime, Message: p.T("worker.extract_failed", data), Hint: p.T("worker.extract_failed_hint")}
	}
	return filepath.Join(dir, filepath.FromSlash(name)), nil
}

// workers holds the workers of one run: the folder where they were
// extracted, once a handler asks for one, until the run removes it.
type workers struct {
	resources fs.FS
	domain    string
	mu        sync.Mutex
	dir       string
	removed   bool
}

// holds reports whether Resources holds the worker name as a file.
func (w *workers) holds(name string) bool {
	if w.resources == nil || !fs.ValidPath(name) {
		return false
	}
	info, err := fs.Stat(w.resources, path.Join("worker", name))
	return err == nil && info.Mode().IsRegular()
}

// errRemoved is the error of a worker that a handler asks for once its run
// removed the folder. A handler that outlives its cleanup can do that.
var errRemoved = errors.New("the run already removed its workers")

// extract writes the workers into a new folder that only the user can
// read, MkdirTemp's 0700, on the first call, and returns that folder.
func (w *workers) extract() (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	switch {
	case w.removed:
		return "", errRemoved
	case w.dir != "":
		return w.dir, nil
	}
	tree, err := fs.Sub(w.resources, "worker")
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "ncly-"+w.domain+"-")
	if err != nil {
		return "", err
	}
	if err := os.CopyFS(dir, tree); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	w.dir = dir
	return dir, nil
}

// remove deletes the folder of the workers, and keeps any later call from
// extracting them again.
func (w *workers) remove() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.removed = true
	if w.dir != "" {
		_ = os.RemoveAll(w.dir)
	}
}
