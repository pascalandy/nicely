package extensionrepo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/config"
)

type releaseMeasurement struct {
	Name        string   `json:"name"`
	Command     []string `json:"command"`
	Exit        int      `json:"exit"`
	ElapsedMS   int64    `json:"elapsed_ms"`
	StdoutBytes int      `json:"stdout_bytes"`
	StderrBytes int      `json:"stderr_bytes"`
	Stdout      string   `json:"stdout"`
	Stderr      string   `json:"stderr"`
}

type releaseReceipt struct {
	dir          string
	Tag          string               `json:"tag"`
	SHA          string               `json:"sha"`
	OS           string               `json:"os"`
	Signoff      signoffReceipt       `json:"signoff"`
	Pushed       bool                 `json:"pushed"`
	ClonedSHA    string               `json:"cloned_sha"`
	Steps        []releaseMeasurement `json:"steps"`
	Economy      []releaseMeasurement `json:"economy"`
	HostedChecks json.RawMessage      `json:"hosted_checks,omitempty"`
	Notes        []string             `json:"notes"`
	Reservations []string             `json:"reservations"`
	Failure      string               `json:"failure,omitempty"`
}

func releaseSigned(sha string) (signoffReceipt, error) {
	base := filepath.Join(config.Locate(os.Getenv("HOME"), os.LookupEnv).Cache, "signoff")
	paths, err := filepath.Glob(filepath.Join(base, "*", "receipt.json"))
	if err != nil {
		return signoffReceipt{}, err
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var r signoffReceipt
		if json.Unmarshal(data, &r) != nil || r.SHA != sha || len(r.Checks) == 0 || r.Signoff == nil || r.Signoff.Exit != 0 || len(r.Signoff.Command) == 0 {
			continue
		}
		var raw struct {
			Checks  []json.RawMessage `json:"checks"`
			Signoff json.RawMessage   `json:"signoff"`
		}
		if json.Unmarshal(data, &raw) != nil {
			continue
		}
		valid := releaseZeroExit(raw.Signoff)
		for _, check := range raw.Checks {
			valid = valid && releaseZeroExit(check)
		}
		for _, check := range r.Checks {
			valid = valid && check.Exit == 0 && len(check.Command) != 0
		}
		if valid {
			return r, nil
		}
	}
	return signoffReceipt{}, fmt.Errorf("no successful receipt for %s; run just signoff for this pushed SHA", sha)
}

// Missing and null exit fields cannot stand for successful checks.
func releaseZeroExit(data json.RawMessage) bool {
	var check map[string]json.RawMessage
	if json.Unmarshal(data, &check) != nil {
		return false
	}
	return bytes.Equal(bytes.TrimSpace(check["exit"]), []byte("0"))
}

func newReleaseRun(root, tag string, signed signoffReceipt) (*releaseReceipt, error) {
	cache := config.Locate(os.Getenv("HOME"), os.LookupEnv).Cache
	if !filepath.IsAbs(cache) {
		return nil, errors.New("release needs an absolute cache folder outside the repository")
	}
	base := filepath.Join(cache, "release")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, err
	}
	base, err := filepath.EvalSymlinks(base)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, base)
	if err != nil {
		return nil, err
	}
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("cache folder is inside the repository; set XDG_CACHE_HOME outside it")
	}
	dir, err := os.MkdirTemp(base, time.Now().UTC().Format("20060102T150405Z")+"-"+signed.SHA+"-")
	if err != nil {
		return nil, err
	}
	return &releaseReceipt{
		dir: dir, Tag: tag, SHA: signed.SHA, OS: runtime.GOOS, Signoff: signed,
		Steps: []releaseMeasurement{}, Economy: []releaseMeasurement{}, Notes: []string{}, Reservations: []string{},
	}, nil
}

func (r *releaseReceipt) write() error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.dir, "receipt.json"), append(data, '\n'), 0o600)
}

// measure keeps streams in separate files, without copying command output into the receipt.
func (r *releaseReceipt) measure(ctx context.Context, env []string, dir, name string, args ...string) (releaseMeasurement, error) {
	row := releaseMeasurement{Name: name, Command: args, Stdout: name + ".stdout", Stderr: name + ".stderr"}
	stdout, err := os.OpenFile(filepath.Join(r.dir, row.Stdout), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return row, err
	}
	defer func() { _ = stdout.Close() }()
	stderr, err := os.OpenFile(filepath.Join(r.dir, row.Stderr), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return row, err
	}
	defer func() { _ = stderr.Close() }()
	program, err := releaseProgram(env, args[0])
	if err != nil {
		row.Exit = 1
		return row, err
	}
	cmd := signoffCommand(ctx, env, dir, program, args[1:]...)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	start := time.Now()
	runErr := cmd.Run()
	row.ElapsedMS, row.Exit = time.Since(start).Milliseconds(), signoffExit(runErr)
	out, outErr := stdout.Stat()
	errout, errErr := stderr.Stat()
	if err := errors.Join(outErr, errErr); err != nil {
		return row, err
	}
	row.StdoutBytes, row.StderrBytes = int(out.Size()), int(errout.Size())
	var exited *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exited) {
		return row, runErr
	}
	return row, nil
}

func (r *releaseReceipt) require(ctx context.Context, env []string, dir, name string, args ...string) error {
	row, err := r.measure(ctx, env, dir, name, args...)
	r.Steps = append(r.Steps, row)
	if err != nil {
		return err
	}
	if row.Exit != 0 {
		return fmt.Errorf("%s exited %d; read %s", name, row.Exit, filepath.Join(r.dir, row.Stderr))
	}
	return r.write()
}

func (r *releaseReceipt) hosted(ctx context.Context, env []string, root string) {
	if _, err := exec.LookPath("gh"); err != nil {
		r.Notes = append(r.Notes, "gh is unavailable; hosted check runs were not read")
		return
	}
	row, err := r.measure(ctx, env, root, "hosted-checks", "gh", "api", "--paginate", "--slurp", "repos/{owner}/{repo}/commits/"+r.SHA+"/check-runs")
	r.Steps = append(r.Steps, row)
	data, readErr := os.ReadFile(filepath.Join(r.dir, row.Stdout))
	if err != nil || readErr != nil || row.Exit != 0 || !json.Valid(bytes.TrimSpace(data)) {
		r.Notes = append(r.Notes, "hosted check runs could not be read; see hosted-checks.stderr")
		return
	}
	r.HostedChecks = json.RawMessage(data)
}
