package transcript

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// videoURL holds the three forms of a video URL that the source program
// accepts. They are a watch URL, whose first parameter is v, a short link,
// and a short. A watch URL with more parameters, such as a playlist, still selects
// its one video, and a playlist page is no video URL.
var videoURL = []*regexp.Regexp{
	regexp.MustCompile(`^https?://(www\.)?youtube\.com/watch\?v=[\w-]+`),
	regexp.MustCompile(`^https?://youtu\.be/[\w-]+`),
	regexp.MustCompile(`^https?://(www\.)?youtube\.com/shorts/[\w-]+`),
}

// probeLimit bounds each version probe of preparation.
const probeLimit = 10 * time.Second

// job is a call that preparation checked: the same value serves the dry run
// and execution.
type job struct {
	// url is the URL as supplied, which keys the item.
	url string
	// output is the absolute output parent, and given is true when --output
	// selected it rather than the config or the current folder.
	output string
	given  bool
	// browser is arc when the config selects Arc's cookies, and empty
	// otherwise.
	browser string
	// recorded is true once the record of a pending run gave output and
	// browser, which a new attempt never picks again.
	recorded bool
	// uv is where PATH found uv.
	uv string
	// cache is Nicely's cache folder.
	cache   string
	timeout time.Duration
}

// prepare checks what the call alone states, before any effect: the URL,
// and the absolute path of an explicit output parent. It reads no config,
// checks no folder, and starts no program, so a saved run decides first.
func prepare(call extension.Call) (job, *contract.Problem) {
	p := call.Printer
	j := job{url: call.Flags["url"]}
	if !isVideoURL(j.url) {
		return j, usage(p, "transcript.run.url_invalid", "transcript.run.url_invalid_hint", map[string]any{"URL": j.url})
	}
	home, _ := os.UserHomeDir()
	j.cache = config.Locate(home, os.LookupEnv).Cache
	dir, given := call.Flags["output"]
	j.given = given
	if given {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return j, usage(p, "transcript.run.output_invalid", "transcript.run.output_invalid_hint", map[string]any{"Dir": dir})
		}
		j.output = abs
	}
	return j, nil
}

// resolve completes the job for a run of the engine, in the order of
// Preparation in spec.md. Unless the record of a pending run gave them, it
// picks the browser and the output parent. Then it checks uv and Python.
func resolve(ctx context.Context, call extension.Call, j *job) *contract.Problem {
	p := call.Printer
	if !j.recorded {
		if problem := pick(p, j, call.Flags["output"]); problem != nil {
			return problem
		}
	}
	for _, name := range []string{"uv", "python3"} {
		found, problem := checkProgram(ctx, p, name)
		if problem != nil {
			return problem
		}
		if name == "uv" {
			j.uv = found
		}
	}
	return nil
}

// pick checks that an explicit output parent exists, as flag gives it. Then
// it reads the config, and the output parent that the config selects when
// --output did not, or else the current folder.
func pick(p *i18n.Printer, j *job, flag string) *contract.Problem {
	if j.given {
		if _, err := folder(j.output); err != nil {
			return usage(p, "transcript.run.output_invalid", "transcript.run.output_invalid_hint", map[string]any{"Dir": flag})
		}
	}
	home, _ := os.UserHomeDir()
	configured, problem := readSettings(p, home, config.Locate(home, os.LookupEnv), j)
	if problem != nil || j.given {
		return problem
	}
	dir := configured
	if dir == "" {
		dir = "."
	}
	abs, err := folder(dir)
	if err != nil {
		return &contract.Problem{Code: contract.ConfigInvalid, Message: p.T("transcript.config_output_missing", map[string]any{"Dir": dir}), Hint: p.T("transcript.run.output_invalid_hint", map[string]any{"Dir": dir})}
	}
	j.output = abs
	return nil
}

// folder is the absolute path of dir, which must be an existing folder.
func folder(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err == nil && !info.IsDir() {
		err = errors.New("not a folder")
	}
	return abs, err
}

// isVideoURL reports whether url names one YouTube video in a form that the
// engine accepts.
func isVideoURL(url string) bool {
	for _, pattern := range videoURL {
		if pattern.MatchString(url) {
			return true
		}
	}
	return false
}

// readSettings reads [transcript] browser and [transcript.youtube] output,
// and returns the configured output, resolved from home, or empty when no
// file sets it. CONFIG_INVALID, whose hint names the file, answers a file
// that cannot be read and a value of another type. It also answers a
// browser other than arc and a relative path.
func readSettings(p *i18n.Printer, home string, paths config.Paths, j *job) (string, *contract.Problem) {
	invalid := func(id string, data map[string]any) *contract.Problem {
		return &contract.Problem{Code: contract.ConfigInvalid, Message: p.T(id, data), Hint: p.T("transcript.config_hint", data)}
	}
	values, err := config.LoadValues(paths)
	var bad *config.InvalidError
	switch {
	case errors.As(err, &bad) && bad.Line > 0:
		return "", invalid("transcript.config_invalid_line", map[string]any{"File": bad.File, "Line": bad.Line})
	case errors.As(err, &bad):
		return "", invalid("transcript.config_invalid", map[string]any{"File": bad.File})
	case err != nil:
		return "", invalid("transcript.config_invalid", map[string]any{"File": paths.Shared})
	}
	set, err := values.Decode("transcript.browser", &j.browser)
	if err != nil {
		return "", invalid("transcript.config_value", map[string]any{"File": set.Source, "Key": "browser"})
	}
	if j.browser != "" && j.browser != "arc" {
		return "", invalid("transcript.config_browser", map[string]any{"File": set.Source, "Browser": j.browser})
	}
	var output string
	set, err = values.Decode("transcript.youtube.output", &output)
	if err != nil {
		return "", invalid("transcript.config_value", map[string]any{"File": set.Source, "Key": "output"})
	}
	if output == "" {
		return "", nil
	}
	resolved, err := config.Path(home, output)
	if err != nil {
		return "", invalid("transcript.config_path", map[string]any{"File": set.Source, "Path": output})
	}
	return resolved, nil
}

// checkProgram checks a program as doctor checks it, with the program
// check of sdk/runner and the entry of the manifest. It returns where PATH
// found it. A program that is missing, older, or whose probe fails answers
// PREREQ_MISSING, whose hint and next hold the install command of this
// machine.
func checkProgram(ctx context.Context, p *i18n.Printer, name string) (string, *contract.Problem) {
	m, err := contract.ParseManifest(manifest, "")
	if err != nil {
		return "", &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.manifest"), Hint: p.T("outcome.check_hint")}
	}
	r := m.Requirements[name]
	found, err := runner.CheckProgram(ctx, name, r, runner.Environ(os.Environ(), nil), probeLimit)
	data := map[string]any{"Program": name, "Path": found.Path, "Found": found.Version, "Version": r.Version}
	var message string
	switch {
	case err != nil:
		message = p.T("transcript.program_probe_failed", data)
	case found.Signal != nil:
		// A signal ended the probe: run answers the signal instead.
		return "", &contract.Problem{Code: contract.Runtime}
	case found.Verdict == runner.ProgramFound:
		return found.Path, nil
	case found.Verdict == runner.ProgramMissing:
		message = p.T("transcript.program_missing", data)
	case found.Verdict == runner.ProgramTooOld:
		message = p.T("transcript.program_old", data)
	default:
		message = p.T("transcript.program_probe_failed", data)
	}
	problem := &contract.Problem{Code: contract.PrereqMissing, Message: message, Hint: p.T("transcript.program_hint", data)}
	if found.Install != nil {
		problem.Hint, problem.Next = strings.Join(found.Install, " "), found.Install
	}
	return "", problem
}
