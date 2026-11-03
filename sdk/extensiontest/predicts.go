package extensiontest

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
)

// predicts proves that a successful dry run predicts its run, for a mode
// that ## Publication of spec.md declares: predicts <command> <arg>...
// --force selects the mode replace. The dry run must change no file of the
// scenario outside the cache and temporary folders, and give the destination
// and the overwrite choice. The run must then publish a regular file at that
// destination and keep each protected path. A run that fails after a
// successful dry run fails, as on the destination with no hard link of #425.
func (k *kit) predicts(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 1 {
		ts.Fatalf("usage: predicts <command> <arg>...")
	}
	defer machineMode(ts)()
	program, words, call := k.pairedCall(ts, args[0], args[1:], filepath.Join(k.bin, "ncly"))
	if _, given := call.line.Switch("dry-run"); given || !call.command.Modes.DryRun {
		ts.Fatalf("predicts: give a command with a dry run, without a dry-run flag")
	}
	mode := "create"
	if on, _ := call.line.Switch("force"); on {
		mode = "replace"
	}
	row, found := k.publicationOf(call.path, mode)
	if !found {
		ts.Fatalf("predicts: ## Publication of spec.md declares no %s mode of `ncly %s`. Fix: give that mode a row", mode, call.path)
	}
	protected := protectedPaths(ts, call, row.protected)
	work, skip := ts.Getenv("WORK"), []string{ts.Getenv("XDG_CACHE_HOME"), ts.Getenv("TMPDIR")}
	before, err := snapshotScenario(work, skip, protected)
	ts.Check(err)
	input, done := saveStdin(ts, "predicts")
	defer done()

	preview := append([]string{"--dry-run=true"}, args[1:]...)
	observed := slices.Clone(preview)
	if args[0] == "ncly-"+k.manifest.Slug() {
		observed = append(k.manifest.Root(), preview...)
	}
	exit := execSaved(ts, input, program, preview)
	code, problem := k.checkedCode(observed, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"))
	switch {
	case problem != "":
		ts.Fatalf("predicts: dry run: %s", problem)
	case exit != 0:
		ts.Fatalf("predicts: the dry run refused with %s and exit %d. Fix: give a call whose dry run succeeds, or prove the refusal with refuses", code, exit)
	}
	plan := answerObject(ts, "dry run")
	destination, isPath := plan[row.destination].(string)
	overwrite, isChoice := plan[row.overwrite].(bool)
	if !isPath || destination == "" || !isChoice {
		ts.Fatalf("predicts: the dry run gives no path at `%s` or no boolean at `%s`, which ## Publication declares", row.destination, row.overwrite)
	}
	if row.dryRun == "pending" {
		for _, op := range row.operations {
			if !pending(plan["pending_checks"], "publication."+op) {
				ts.Fatalf("predicts: the dry run lists no publication.%s in pending_checks, and ## Publication declares the treatment pending. Fix: list it, because the dry run cannot check it", op)
			}
		}
	}
	after, err := snapshotScenario(work, skip, protected)
	ts.Check(err)
	if changed := changedPaths(before, after); len(changed) > 0 {
		ts.Fatalf("predicts: the dry run changed %s. Fix: change no file in a dry run, which may fill only the cache of Nicely", strings.Join(changed, ", "))
	}

	target := filepath.Clean(ts.MkAbs(destination))
	if slices.Contains(protected, target) {
		ts.Fatalf("predicts: the destination %s is a protected path of the call", destination)
	}
	_, err = os.Lstat(target)
	existed := err == nil
	switch {
	case mode == "create" && existed:
		ts.Fatalf("predicts: the dry run predicts a new file at %s, which exists. Fix: refuse an existing destination without --force, as Output: Exists, without --force in Generic cases says", destination)
	case overwrite != existed:
		ts.Fatalf("predicts: the dry run gives %s %t, and the destination %s %s", row.overwrite, overwrite, destination, map[bool]string{true: "exists", false: "does not exist"}[existed])
	}
	kept, err := snapshotFiles(protected, false)
	ts.Check(err)
	exit = execSaved(ts, input, program, args[1:])
	code, problem = k.checkedCode(words, exit, ts.ReadFile("stdout"), ts.ReadFile("stderr"))
	switch {
	case problem != "":
		ts.Fatalf("predicts: run: %s", problem)
	case exit != 0:
		ts.Fatalf("predicts: the dry run predicted a file at %s, and the run failed with %s and exit %d. "+
			"Fix: make the dry run check what the run needs, such as the operations %s, or declare pending and list them in pending_checks", destination, code, exit, strings.Join(row.operations, ", "))
	}
	if published, _ := answerObject(ts, "run")[row.destination].(string); published == "" || filepath.Clean(ts.MkAbs(published)) != target {
		ts.Fatalf("predicts: the run published at %q, and the dry run predicted %s", published, destination)
	}
	if info, err := os.Lstat(target); err != nil || !info.Mode().IsRegular() {
		ts.Fatalf("predicts: the run left no regular file at %s", destination)
	}
	same, err := unchanged(protected, kept)
	ts.Check(err)
	if !same {
		ts.Fatalf("predicts: the run changed protected paths: %s", strings.Join(protected, ", "))
	}
	k.prove(call.path, mode, "predicts")
}

// protectedPaths returns the absolute path of each value of a call that
// names an argument or a flag of the row.
func protectedPaths(ts *testscript.TestScript, call parsed, names []string) []string {
	var values []string
	if len(call.line.Words) >= len(call.command.Path) {
		values = slices.Clone(call.line.Words[len(call.command.Path):])
	}
	values = append(values, call.line.Operands...)
	var paths []string
	for _, name := range names {
		if flag, isFlag := strings.CutPrefix(name, "--"); isFlag {
			if value, given := call.line.Value(flag); given {
				paths = append(paths, filepath.Clean(ts.MkAbs(value)))
			}
			continue
		}
		for j, arg := range call.command.Args {
			switch {
			case arg.Name != name || j >= len(values):
			case arg.Variadic:
				for _, value := range values[j:] {
					paths = append(paths, filepath.Clean(ts.MkAbs(value)))
				}
			default:
				paths = append(paths, filepath.Clean(ts.MkAbs(values[j])))
			}
		}
	}
	return paths
}

// snapshotScenario records the scenario folder, without the folders of
// skip, and the protected paths. A file of a type that protect refuses,
// such as a socket, counts by its mode alone.
func snapshotScenario(work string, skip, protected []string) (map[string]fileState, error) {
	states, err := snapshotFiles(protected, false)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(work, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && slices.Contains(skip, path) {
			return filepath.SkipDir
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		state := fileState{mode: info.Mode()}
		switch {
		case info.Mode().IsRegular():
			state.sum, err = fileDigest(path)
		case info.Mode()&os.ModeSymlink != 0:
			state.link, err = os.Readlink(path)
		}
		states[path] = state
		return err
	})
	return states, err
}

// changedPaths returns each path that two snapshots record differently,
// sorted.
func changedPaths(before, after map[string]fileState) []string {
	var changed []string
	for path, state := range before {
		if other, found := after[path]; !found || other != state {
			changed = append(changed, path)
		}
	}
	for path := range after {
		if _, found := before[path]; !found {
			changed = append(changed, path)
		}
	}
	slices.Sort(changed)
	return slices.Compact(changed)
}

// answerObject decodes the last line of stdout, the answer that
// checkedCode validated.
func answerObject(ts *testscript.TestScript, mode string) map[string]any {
	lines := strings.Split(strings.TrimRight(ts.ReadFile("stdout"), "\n"), "\n")
	var answer map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &answer); err != nil {
		ts.Fatalf("predicts: the %s gives no JSON object on stdout: %v", mode, err)
	}
	return answer
}

// pending reports whether pending_checks lists id, as a string or as an
// object with that id.
func pending(checks any, id string) bool {
	items, _ := checks.([]any)
	return slices.ContainsFunc(items, func(item any) bool {
		object, _ := item.(map[string]any)
		return item == id || object["id"] == id
	})
}
