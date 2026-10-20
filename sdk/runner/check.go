package runner

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"golang.org/x/sys/unix"
)

// Verdict is the outcome of a program check, as Required programs in
// core-spec.md defines it.
type Verdict string

const (
	// ProgramFound is a program found in PATH, recent enough when it
	// declares a version.
	ProgramFound Verdict = "found"
	// ProgramMissing is a program that no folder of PATH holds.
	ProgramMissing Verdict = "missing"
	// ProgramTooOld is a program whose version is older than the one it
	// declares.
	ProgramTooOld Verdict = "too_old"
	// ProgramProbeFailed is a probe that exited with another code than 0.
	ProgramProbeFailed Verdict = "probe_failed"
	// ProgramNoVersion is a probe that printed no version.
	ProgramNoVersion Verdict = "no_version"
	// ProgramTimedOut is a probe that still ran at its limit, or a signal
	// that ended it.
	ProgramTimedOut Verdict = "timed_out"
)

// ProgramCheck is what a program check found.
type ProgramCheck struct {
	Verdict Verdict
	// Path is where the lookup found the program, or empty when missing.
	Path string
	// Version is the version that the probe printed, or empty.
	Version string
	// Exit is the probe's exit code, for ProgramProbeFailed.
	Exit int
	// Install is the install command of the first package manager of this
	// system that the requirement names and PATH holds, or nil.
	Install []string
	// Signal is the signal that ended the probe, or nil.
	Signal os.Signal
}

// Failed reports whether the check fails with PREREQ_MISSING.
func (c ProgramCheck) Failed() bool {
	return c.Verdict != ProgramFound && c.Verdict != ProgramTimedOut
}

// probeOutput bounds what a version probe keeps of each stream. A version
// comes first, so a probe that writes without end never fills memory.
const probeOutput = 64 << 10

// CheckProgram checks one program that a manifest requires. It looks name
// up in the PATH of env. When r declares a version, it runs the probe for at
// most limit, with stdin closed, in a new session. It then compares the
// version that the probe prints with r's. A failed check names the install
// command of this system. Doctor, sdk/wrap, and a Go extension's status
// command share it, so none copies the rule.
func CheckProgram(ctx context.Context, name string, r contract.Requirement, env []string, limit time.Duration) (ProgramCheck, error) {
	path := pathOf(env)
	c := ProgramCheck{Verdict: ProgramFound, Path: Lookup(name, path)}
	if c.Path == "" {
		c.Verdict = ProgramMissing
	} else if r.Version != "" {
		stdout, stderr := Capture{Limit: probeOutput}, Capture{Limit: probeOutput}
		result, err := Run(ctx, Program{Path: c.Path, Args: r.ProbeArgs(), Env: env, Stdout: &stdout, Stderr: &stderr, Limit: limit})
		if err != nil {
			return ProgramCheck{}, err
		}
		c.Version = versionIn(string(stdout.Bytes()))
		if c.Version == "" {
			c.Version = versionIn(string(stderr.Bytes()))
		}
		switch {
		case result.TimedOut || result.Signal != nil:
			c.Verdict, c.Version, c.Signal = ProgramTimedOut, "", result.Signal
		case result.Exit != 0:
			c.Verdict, c.Exit, c.Version = ProgramProbeFailed, result.Exit, ""
		case c.Version == "":
			c.Verdict = ProgramNoVersion
		case olderThan(c.Version, r.Version):
			c.Verdict = ProgramTooOld
		}
	}
	if c.Failed() {
		c.Install = installCommand(r.Install, path)
	}
	return c, nil
}

// pathOf returns the value of PATH in env.
func pathOf(env []string) string {
	for _, v := range slices.Backward(env) {
		if value, ok := strings.CutPrefix(v, "PATH="); ok {
			return value
		}
	}
	return ""
}

// Lookup returns the first executable file named name in the folders of
// path, skipping an empty or relative one, as discovery does, or empty.
func Lookup(name, path string) string {
	for _, dir := range filepath.SplitList(path) {
		if !filepath.IsAbs(dir) {
			continue
		}
		file := filepath.Join(dir, name)
		if isExecutable(file) {
			return file
		}
	}
	return ""
}

// isExecutable reports whether path is a regular file, or a link to one,
// that the user may execute.
func isExecutable(path string) bool {
	var st unix.Stat_t
	return unix.Stat(path, &st) == nil && st.Mode&unix.S_IFMT == unix.S_IFREG && unix.Access(path, unix.X_OK) == nil
}

// versionPattern is the first run of two or more numbers joined by dots,
// such as 6.1.1 in "ffmpeg version 6.1.1-3ubuntu5".
var versionPattern = regexp.MustCompile(`[0-9]+(\.[0-9]+)+`)

func versionIn(text string) string { return versionPattern.FindString(text) }

// olderThan compares two versions number by number, where a missing number
// counts as 0, so 3.12 meets 3.12.0. A number of any length compares by its
// value.
func olderThan(found, want string) bool {
	a, b := strings.Split(found, "."), strings.Split(want, ".")
	for i := range max(len(a), len(b)) {
		var x, y string
		if i < len(a) {
			x = strings.TrimLeft(a[i], "0")
		}
		if i < len(b) {
			y = strings.TrimLeft(b[i], "0")
		}
		if len(x) != len(y) {
			return len(x) < len(y)
		}
		if x != y {
			return x < y
		}
	}
	return false
}

// managers lists the package managers that a check knows, in the order of
// this system.
func managers() []string {
	if runtime.GOOS == "darwin" {
		return []string{"brew"}
	}
	return []string{"pacman", "apt", "dnf", "brew"}
}

// installCommand returns the install command of the first package manager
// of this system that install names and path holds.
func installCommand(install map[string][]string, path string) []string {
	for _, manager := range managers() {
		if command, named := install[manager]; named && Lookup(manager, path) != "" {
			return command
		}
	}
	return nil
}
