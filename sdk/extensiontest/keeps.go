package extensiontest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

// keeps runs any registered call with its exact exit, and keeps the
// protected paths around it: keeps <exit> <command> <arg>... The paths are
// those of the last protect call, and the Protected paths of the row of
// its mode. It checks the children of the call before the cleanup of the
// kit: on Linux, the guard reports each descendant that still runs. On
// another platform, the process group of the call must be empty.
func (k *kit) keeps(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 2 {
		ts.Fatalf("usage: keeps <exit> <command> <arg>...")
	}
	want, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("keeps: invalid exit %q", args[0])
	}
	defer machineMode(ts)()
	program, words, call := k.pairedCall(ts, args[1], args[2:], filepath.Join(k.bin, "ncly"))
	mode := modeOf(call)
	row, hasRow := k.publicationOf(call.path, mode)
	protected := k.takeNext(ts)
	if hasRow {
		protected = slices.Concat(protected, protectedPaths(ts, call, row.protected))
	}
	g := k.guarded(ts, "keeps", true)
	before, err := snapshotFiles(protected, false)
	ts.Check(err)
	guard, guarded := g.wrap(program, args[2:])
	exit, stdout, stderr := foreground(ts, "keeps", guard, guarded, func(done <-chan struct{}, host *os.Process) error {
		<-done
		if g != nil {
			return nil
		}
		return emptyGroup(host.Pid)
	})
	report := g.read(ts, "keeps")
	if _, problem := k.checkedCode(words, exit, stdout, stderr); problem != "" {
		ts.Fatalf("keeps: %s", problem)
	}
	if exit != want {
		ts.Fatalf("keeps: %s exited with %d, want %d", args[1], exit, want)
	}
	same, err := unchanged(protected, before)
	ts.Check(err)
	if !same {
		ts.Fatalf("keeps: the call changed protected paths: %s", strings.Join(protected, ", "))
	}
	dryRun, _ := call.line.Switch("dry-run")
	if dryRun {
		// A dry run publishes nothing, and proves no operation of a run.
		return
	}
	if exit == 0 && hasRow && g != nil {
		destination, _ := decodeAnswer(stdout)[row.destination].(string)
		checkCommit(ts, "keeps", report, row, destination)
	}
	if op := g.proves(ts, "keeps"); op != "" {
		k.prove(call.path, mode, op)
	}
}

// emptyGroup waits up to two seconds until the group of pid is empty. The
// guard waits as long for the descendants of a call.
func emptyGroup(pid int) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := syscall.Kill(-pid, 0); errors.Is(err, syscall.ESRCH) {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("keeps: a process of the group of the call still ran after it exited. Fix: stop each child before the command exits, as runner does")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// decodeAnswer decodes the last line of stdout, an answer that checkedCode
// validated, as a JSON object.
func decodeAnswer(stdout string) map[string]any {
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	var answer map[string]any
	_ = json.Unmarshal([]byte(lines[len(lines)-1]), &answer)
	return answer
}
