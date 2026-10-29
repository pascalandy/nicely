package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// efforts are the reasoning efforts that preparation knows, as Codex
// invocation in spec.md lists them. The model may still refuse one.
var efforts = []string{"low", "medium", "high", "xhigh", "max", "ultra"}

// defaultEffort applies when neither --effort nor [agent] effort sets one.
const defaultEffort = "medium"

// probeLimit bounds codex --version during preparation.
const probeLimit = 10 * time.Second

// task is a task that preparation checked: the same value serves the dry
// run and execution. prepare fills what the call states, and pick the
// model, the effort, Codex, and its environment.
type task struct {
	text   []byte
	sha256 string
	model  string
	effort string
	cwd    string
	// codex is where PATH found Codex, and version the version that it
	// printed.
	codex   string
	version string
	// env is the environment of Codex: the one of ncly, without the key of
	// any service, with depth, the depth of Codex, in NCLY_AGENT_DEPTH.
	env   []string
	depth int
	// settings are the values of [agent], which pick reads.
	settings settings
	// timeout is the limit of the whole call, which its errors name.
	timeout time.Duration
}

// args returns the words after codex, as Codex invocation in spec.md lists
// them. The task goes on stdin, so it never appears in them.
func (t task) args(answer string) []string {
	return []string{
		"exec", "--model", t.model, "-c", `model_reasoning_effort="` + t.effort + `"`,
		"--sandbox", "workspace-write", "--skip-git-repo-check", "--cd", t.cwd,
		"--json", "--output-last-message", answer, "-",
	}
}

// settings are the values of [agent] in the config files.
type settings struct {
	model    string
	effort   string
	maxDepth int64
	// file is the shared config file, which a hint names, and depthFile
	// the file that set max_depth.
	file      string
	depthFile string
}

// prepare checks what the call states, in the order of ncly agent in
// spec.md, and answers the first problem: the settings, the depth, the
// task, the folder, and the tools mode. It reads no stdin and starts no
// task. pick then checks what a new task needs.
func prepare(call extension.Call) (task, *contract.Problem) {
	p := call.Printer
	s, problem := readSettings(p)
	if problem != nil {
		return task{}, problem
	}
	t := task{settings: s}
	if t.depth, problem = childDepth(p, s); problem != nil {
		return task{}, problem
	}
	if t.text, problem = readTask(p, call); problem != nil {
		return task{}, problem
	}
	sum := sha256.Sum256(t.text)
	t.sha256 = hex.EncodeToString(sum[:])
	if t.cwd, problem = workingFolder(p, call); problem != nil {
		return task{}, problem
	}
	if call.Flags["tools"] != "true" {
		return task{}, &contract.Problem{Code: contract.CapabilityUnsupported, Message: p.T("agent.run.tools_required"), Hint: p.T("agent.run.tools_required_hint")}
	}
	return t, nil
}

// pick checks what a new task needs, once no record forbids it: the model
// and the effort, from the flags over the settings, or from the resolved
// values of an earlier attempt, then Codex, through the environment that
// Codex receives.
func pick(ctx context.Context, call extension.Call, t *task, resolved map[string]any) *contract.Problem {
	p := call.Printer
	t.model, t.effort = t.settings.model, t.settings.effort
	if model, given := call.Flags["model"]; given {
		t.model = model
	}
	if effort, given := call.Flags["effort"]; given {
		t.effort = effort
	}
	if resolved != nil {
		t.model, _ = resolved["model"].(string)
		t.effort, _ = resolved["effort"].(string)
	}
	if strings.TrimSpace(t.model) == "" {
		return usage(p, "agent.run.model_missing", "agent.run.model_missing_hint", map[string]any{"File": t.settings.file})
	}
	t.env = runner.Environ(os.Environ(), map[string]string{depthVariable: strconv.Itoa(t.depth)})
	var problem *contract.Problem
	t.codex, t.version, problem = checkCodex(ctx, p, t.env)
	return problem
}

// usage is a USAGE_INVALID problem whose message and hint come from the
// catalog.
func usage(p *i18n.Printer, message, hint string, data map[string]any) *contract.Problem {
	return &contract.Problem{Code: contract.UsageInvalid, Message: p.T(message, data), Hint: p.T(hint, data)}
}

// readSettings reads model, effort, and max_depth from [agent], each
// defaulting when no file sets it. A file that cannot be read, a value of
// another type, an unknown effort, and a max_depth below 1 answer
// CONFIG_INVALID, whose hint names the file.
func readSettings(p *i18n.Printer) (settings, *contract.Problem) {
	home, _ := os.UserHomeDir()
	paths := config.Locate(home, os.LookupEnv)
	s := settings{effort: defaultEffort, maxDepth: defaultMaxDepth, file: paths.Shared}
	invalid := func(id string, data map[string]any) *contract.Problem {
		return &contract.Problem{Code: contract.ConfigInvalid, Message: p.T(id, data), Hint: p.T("agent.config_hint", data)}
	}
	values, err := config.LoadValues(paths)
	var bad *config.InvalidError
	switch {
	case errors.As(err, &bad) && bad.Line > 0:
		return s, invalid("agent.config_invalid_line", map[string]any{"File": bad.File, "Line": bad.Line})
	case errors.As(err, &bad):
		return s, invalid("agent.config_invalid", map[string]any{"File": bad.File})
	case err != nil:
		return s, invalid("agent.config_invalid", map[string]any{"File": paths.Shared})
	}
	if set, err := values.Decode("agent.model", &s.model); err != nil {
		return s, invalid("agent.config_value", map[string]any{"File": set.Source, "Key": "model"})
	}
	set, err := values.Decode("agent.effort", &s.effort)
	if err != nil {
		return s, invalid("agent.config_value", map[string]any{"File": set.Source, "Key": "effort"})
	}
	if !slices.Contains(efforts, s.effort) {
		return s, invalid("agent.config_effort", map[string]any{"File": set.Source, "Effort": s.effort, "Efforts": strings.Join(efforts, ", ")})
	}
	set, err = values.Decode("agent.max_depth", &s.maxDepth)
	if err != nil || s.maxDepth < 1 {
		return s, invalid("agent.config_max_depth", map[string]any{"File": set.Source})
	}
	s.depthFile = set.Source
	return s, nil
}

// isInput reports whether info is the input of ncly, such as /dev/stdin
// when a regular file feeds it, which this command never reads.
func isInput(info os.FileInfo) bool {
	in, err := os.Stdin.Stat()
	return err == nil && os.SameFile(info, in)
}

// readTask returns the task of the call: its argument, or the regular file
// that --task-file names, read from the current folder. Exactly one of them
// gives a task that holds more than whitespace.
func readTask(p *i18n.Printer, call extension.Call) ([]byte, *contract.Problem) {
	file, fromFile := call.Flags["task-file"]
	switch {
	case fromFile && len(call.Args) > 0:
		return nil, usage(p, "agent.run.task_twice", "agent.run.task_hint", nil)
	case !fromFile && len(call.Args) == 0:
		return nil, usage(p, "agent.run.task_missing", "agent.run.task_hint", nil)
	}
	text := []byte(strings.Join(call.Args, " "))
	if fromFile {
		// Only a regular file holds a task: stdin, a pipe, or a device would
		// make preparation read the input of ncly, or wait for a writer.
		info, err := os.Stat(file)
		if err == nil && !info.Mode().IsRegular() {
			return nil, usage(p, "agent.run.task_not_file", "agent.run.task_unreadable_hint", map[string]any{"File": file})
		}
		f, err := os.Open(file)
		if err != nil {
			return nil, usage(p, "agent.run.task_unreadable", "agent.run.task_unreadable_hint", map[string]any{"File": file})
		}
		defer func() { _ = f.Close() }()
		// The opened file, not its path, shows the input of ncly: macOS gives
		// /dev/stdin the stat of another device.
		if opened, err := f.Stat(); err == nil && isInput(opened) {
			return nil, usage(p, "agent.run.task_is_input", "agent.run.task_is_input_hint", map[string]any{"File": file})
		}
		if text, err = io.ReadAll(f); err != nil {
			return nil, usage(p, "agent.run.task_unreadable", "agent.run.task_unreadable_hint", map[string]any{"File": file})
		}
	}
	if strings.TrimSpace(string(text)) == "" {
		return nil, usage(p, "agent.run.task_empty", "agent.run.task_hint", nil)
	}
	return text, nil
}

// workingFolder returns the absolute folder where Codex works: --cwd, from
// the current folder, or the current folder itself.
func workingFolder(p *i18n.Printer, call extension.Call) (string, *contract.Problem) {
	dir := call.Flags["cwd"]
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err == nil {
		var info os.FileInfo
		if info, err = os.Stat(abs); err == nil && !info.IsDir() {
			err = errors.New("not a folder")
		}
	}
	if err != nil {
		return "", usage(p, "agent.run.cwd_invalid", "agent.run.cwd_invalid_hint", map[string]any{"Dir": dir})
	}
	return abs, nil
}

// checkCodex checks Codex as doctor checks its agent.codex, with the
// program check of sdk/runner and the entry of the manifest, and returns
// where it found Codex and the version that Codex printed. A Codex that
// is missing, older, or whose probe fails answers PREREQ_MISSING, whose
// hint and next hold the install command of this machine.
func checkCodex(ctx context.Context, p *i18n.Printer, env []string) (string, string, *contract.Problem) {
	m, err := contract.ParseManifest(manifest, "")
	if err != nil {
		return "", "", &contract.Problem{Code: contract.Runtime, Message: p.T("agent.run.manifest"), Hint: p.T("outcome.check_hint")}
	}
	r := m.Requirements["codex"]
	found, err := runner.CheckProgram(ctx, "codex", r, env, probeLimit)
	data := map[string]any{"Path": found.Path, "Found": found.Version, "Version": r.Version, "Exit": found.Exit, "Probe": strings.Join(append([]string{"codex"}, r.ProbeArgs()...), " ")}
	var message string
	switch {
	case err != nil:
		data["Exit"] = -1
		message = p.T("agent.codex_probe_failed", data)
	case found.Signal != nil:
		// A signal ended the probe: run answers the signal instead.
		return "", "", &contract.Problem{Code: contract.Runtime}
	case found.Verdict == runner.ProgramFound:
		return found.Path, found.Version, nil
	case found.Verdict == runner.ProgramMissing:
		message = p.T("agent.codex_missing", data)
	case found.Verdict == runner.ProgramTooOld:
		message = p.T("agent.codex_old", data)
	case found.Verdict == runner.ProgramProbeFailed:
		message = p.T("agent.codex_probe_failed", data)
	case found.Verdict == runner.ProgramNoVersion:
		message = p.T("agent.codex_no_version", data)
	default:
		return "", "", &contract.Problem{Code: contract.PrereqMissing, Message: p.T("agent.codex_timed_out", data), Hint: p.T("agent.codex_timed_out_hint", data)}
	}
	problem := &contract.Problem{Code: contract.PrereqMissing, Message: message, Hint: p.T("agent.codex_hint")}
	if found.Install != nil {
		problem.Hint, problem.Next = strings.Join(found.Install, " "), found.Install
	}
	return "", "", problem
}
