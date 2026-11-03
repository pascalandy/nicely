package extensiontest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
)

// A guardRun runs the calls of one check under the guard program, which
// the kit builds from guard/ on Linux. The guard adopts each orphan of
// a call, and records each link and rename in the scenario folder. With
// refuse, it refuses that operation there for the call only, as a file
// system without that ability does. HOME holds the scenario. The folders of
// Nicely stay outside the scope: TMPDIR and the XDG folders hold the
// temporary files, the config, the records, and the cache.
type guardRun struct {
	program, refuse, dir string
	scope                string
	skip                 []string
	calls                int
	refused              bool
}

// guardReport and guardOperation read the report of the guard.
type guardReport struct {
	Exit       int              `json:"exit"`
	Error      string           `json:"error"`
	Operations []guardOperation `json:"operations"`
	Children   []string         `json:"children"`
}

type guardOperation struct {
	Name    string `json:"operation"`
	Target  string `json:"target"`
	Refused bool   `json:"refused"`
	Done    bool   `json:"done"`
}

// guarded returns the guard of the next check. It refuses the operation of
// a without call. observe asks for a guard on Linux without one, so keeps
// and holds see the children and the commit. It returns nil when no guard
// runs.
func (c *Coverage) guarded(ts *testscript.TestScript, name string, observe bool) *guardRun {
	files := c.script(ts)
	refuse := files.without
	files.without = ""
	switch {
	case refuse == "" && (!observe || c.guard == ""):
		return nil
	case refuse != "" && c.guard == "":
		ts.Fatalf("%s: without needs the guard, which only extensiontest.Run builds on Linux", name)
	}
	dir, err := os.MkdirTemp("", "ncly-guard-")
	ts.Check(err)
	ts.Defer(func() { _ = os.RemoveAll(dir) })
	g := &guardRun{program: c.guard, refuse: refuse, dir: dir, scope: ts.Getenv("WORK")}
	for _, name := range []string{"TMPDIR", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME"} {
		if value := ts.Getenv(name); value != "" {
			g.skip = append(g.skip, value)
		}
	}
	return g
}

// wrap returns the guard call that runs program with args. A nil guard
// returns the call unchanged.
func (g *guardRun) wrap(program string, args []string) (string, []string) {
	if g == nil {
		return program, args
	}
	g.calls++
	call := []string{"-report", g.report(), "-scope", g.scope}
	for _, dir := range g.skip {
		call = append(call, "-skip", dir)
	}
	if g.refuse != "" {
		call = append(call, "-refuse", g.refuse)
	}
	return g.program, append(append(call, "--", program), args...)
}

func (g *guardRun) report() string {
	return filepath.Join(g.dir, fmt.Sprintf("report-%d.json", g.calls))
}

// read returns the report of the last call. It fails the scenario when the
// guard could not supervise the call, or when a descendant still ran after
// the call exited.
func (g *guardRun) read(ts *testscript.TestScript, name string) guardReport {
	var r guardReport
	if g == nil {
		return r
	}
	data, err := os.ReadFile(g.report())
	if err == nil {
		err = json.Unmarshal(data, &r)
	}
	switch {
	case err != nil:
		ts.Fatalf("%s: the guard of the kit wrote no report: %v", name, err)
	case r.Error != "":
		ts.Fatalf("%s: this host cannot run the guard of the kit: %s. Fix: run the scenario on Linux 5.5 or later, which allows a seccomp filter with a listener", name, r.Error)
	case len(r.Children) > 0:
		ts.Fatalf("%s: these processes of the call still ran after it exited: %s. Fix: stop each child before the command exits, as runner does", name, strings.Join(r.Children, "; "))
	}
	g.refused = g.refused || slices.ContainsFunc(r.Operations, func(op guardOperation) bool { return op.Refused })
	return r
}

// proves requires that a call of the check met the refusal, and returns
// the operation that the check proves. It returns "" without a refusal.
func (g *guardRun) proves(ts *testscript.TestScript, name string) string {
	if g == nil || g.refuse == "" {
		return ""
	}
	if !g.refused {
		ts.Fatalf("%s: no call tried `%s` in the scenario folder, so the check proves nothing about a destination without it. Fix: give a call that publishes a file, or remove without", name, g.refuse)
	}
	return g.refuse
}

// checkCommit requires that the run published destination with an
// operation of its row, as the guard observed it. Generic cases forbid a
// rename that replaces in the mode create: a file that appears after the
// check is lost.
func checkCommit(ts *testscript.TestScript, name string, r guardReport, row publication, destination string) {
	target := realPath(ts.MkAbs(destination))
	var used []string
	for _, op := range r.Operations {
		if op.Target == target && op.Done {
			used = append(used, op.Name)
		}
	}
	slices.Sort(used)
	used = slices.Compact(used)
	if len(used) == 0 {
		ts.Fatalf("%s: the run published %s with no link and no rename, so a reader could see a partial file. Fix: publish with an operation of its row: %s", name, destination, strings.Join(row.operations, ", "))
	}
	for _, op := range used {
		if slices.Contains(row.operations, op) {
			continue
		}
		why := ""
		if row.mode == "create" && op == "rename" {
			why = " A create never renames over a file, because a file that appears after a check is lost, as Generic cases say."
		}
		ts.Fatalf("%s: the run published %s with `%s`, which the %s mode of `ncly %s` does not declare.%s Fix: publish with %s, or fail", name, destination, op, row.mode, row.path, why, strings.Join(row.operations, " or "))
	}
}

// realPath resolves the symbolic links of the folder of path, as the guard
// reads the folder of a process in /proc.
func realPath(path string) string {
	path = filepath.Clean(path)
	if dir, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
		return filepath.Join(dir, filepath.Base(path))
	}
	return path
}

// without runs one check with an operation refused in the scenario folder:
// without link|rename-noreplace|rename <check> <arg>... The check is
// refuses, predicts, appears, or keeps. Only Linux can refuse an operation
// for the call alone, so another platform fails with its name.
func (k *kit) without(commands map[string]func(*testscript.TestScript, bool, []string)) func(*testscript.TestScript, bool, []string) {
	checks := []string{"refuses", "predicts", "appears", "keeps"}
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg || len(args) < 2 || !slices.Contains(operations, args[0]) || !slices.Contains(checks, args[1]) {
			ts.Fatalf("usage: without %s %s <arg>...", strings.Join(operations, "|"), strings.Join(checks, "|"))
		}
		if runtime.GOOS != "linux" {
			ts.Fatalf("without: %s cannot refuse `%s` for one call alone. Only Linux can, with a seccomp filter. Fix: run this line on Linux only, as [linux] without %s %s", runtime.GOOS, args[0], args[0], strings.Join(args[1:], " "))
		}
		k.script(ts).without = args[0]
		defer func() { k.script(ts).without = "" }()
		commands[args[1]](ts, false, args[2:])
	}
}
