package extensionrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// signoffReplay records the run of the first scenario that signoff repeats.
type signoffReplay struct {
	Commit  string   `json:"commit"`
	Command []string `json:"command"`
	Exit    int      `json:"exit"`
	// Failed is the line of the scenario that failed in the repeat.
	Failed string `json:"failed"`
}

// replayFirstRun runs the scenario of [first_run] of the review.toml of
// worktree again, at its Scenario commit, in a new worktree under run. It
// builds the go test arguments from the name of the scenario, so the text of
// the record never runs (#422). The run must fail on a line of the scenario
// that the recorded assertion holds. So an invented scenario, or a compile
// error, signs nothing.
func replayFirstRun(ctx context.Context, env []string, root, worktree, run string, log io.Writer) (*signoffReplay, error) {
	const fix = "Fix: run the first scenario at its Scenario commit with just scenario <name>, and record that run in [first_run] of review.toml"
	data, err := os.ReadFile(filepath.Join(worktree, reviewFile))
	if err != nil {
		return nil, err
	}
	var record reviewRecord
	if problems := decodeRecord(reviewFile, string(data), &record); len(problems) != 0 {
		return nil, errors.New(strings.Join(problems, "\n"))
	}
	first := record.FirstRun
	if first == nil || !scenarioName.MatchString(first.Scenario) || record.Order.Scenario == "" {
		return nil, fmt.Errorf("review.toml records no scenario name and no Scenario commit to run again. %s", fix)
	}
	commit, err := signoffOutput(ctx, env, worktree, "git", "rev-parse", "--verify", record.Order.Scenario+"^{commit}")
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(run, "first-run")
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if output, err := signoffCommand(cleanupCtx, env, root, "git", "worktree", "remove", "--force", dir).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "signoff: worktree cleanup: %v: %s\n", err, output)
		}
	}()
	if _, err := signoffOutput(ctx, env, root, "git", "worktree", "add", "--quiet", "--detach", dir, commit); err != nil {
		return nil, err
	}
	args := append([]string{"go", "test", "-count=1"}, scenarioArgs(first.Scenario)...)
	if _, err := fmt.Fprintf(log, "$ %s  # at the Scenario commit %s\n", strings.Join(args, " "), commit); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	cmd := signoffCommand(ctx, append(env, ScenarioEnv+"="+first.Scenario), dir, args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = io.MultiWriter(&output, log), io.MultiWriter(&output, log)
	replay := &signoffReplay{Commit: commit, Command: args, Exit: signoffExit(cmd.Run())}
	if err := ctx.Err(); err != nil {
		return replay, err
	}
	file := "testdata/script/" + first.Scenario + ".txtar"
	if replay.Exit == 0 {
		return replay, fmt.Errorf("the first scenario %s passes at its Scenario commit %s, so it never failed before the handler. %s", file, commit, fix)
	}
	failed := regexp.MustCompile(`(?m)^\s*FAIL: .*testdata/script/` + regexp.QuoteMeta(first.Scenario) + `\.txtar:([0-9]+): `).FindStringSubmatch(output.String())
	if failed == nil {
		return replay, fmt.Errorf("the first scenario %s fails at %s before any line of the scenario, such as on a compile error, which proves no missing behavior. "+
			"The log is %s. %s", file, commit, filepath.Join(run, "check.log"), fix)
	}
	script, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return replay, err
	}
	number, _ := strconv.Atoi(failed[1])
	lines := strings.Split(string(script), "\n")
	if number < 1 || number > len(lines) {
		return replay, fmt.Errorf("the first scenario %s fails at %s on line %d, which it does not hold. %s", file, commit, number, fix)
	}
	replay.Failed = strings.TrimSpace(lines[number-1])
	if replay.Failed == "" || !strings.Contains(first.Assertion, replay.Failed) {
		return replay, fmt.Errorf("the first scenario %s fails at %s on line %d, %q, and first_run of review.toml names the assertion %q. "+
			"Fix: record the assertion that failed, as the run names it", file, commit, number, replay.Failed, first.Assertion)
	}
	return replay, nil
}
