// Package verification runs the TLA+ models of a repository with a pinned
// TLC, as Formal models in the contract defines (COR-104-T8). The repository
// owns its models, their mapping to its code, and their expected outcomes.
// This package owns the tool and the record.
//
// Its rules:
//
//   - The jar holds the SHA-256 that the suite pins before Java starts, from
//     the URL or from a local jar. On 2026-11-02, the pinned URL of TLC v1.8.0
//     served other bytes. Changed bytes never prove a model
//   - A run writes the record only when each check gives its expected
//     outcome, and no input changed during the run
//   - CheckRecord reads only the suite, the inputs, and the record. So just
//     check needs no Java and no network
//   - The caller gives the whole environment of Java, which
//     extensionrepo.NewOwnedRun builds. The analyzer callerenv refuses
//     os.Environ in this package
package verification

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// SuiteFile and RecordFile are the suite and the record of a repository,
// from its root.
const (
	SuiteFile  = "verification/tlc.json"
	RecordFile = "verification/tlc-record.json"
)

// Suite pins Java and TLC, and lists the checks of a repository. Java is the
// tool of mise, such as java@temurin-21.0.13+9.0.LTS.
type Suite struct {
	Java   string  `json:"java"`
	URL    string  `json:"tlc_url"`
	SHA256 string  `json:"tlc_sha256"`
	Checks []Check `json:"checks"`
}

// Check is one model and one configuration, from the root. Violation names
// the invariant that TLC must find violated, or is empty when TLC must
// explore every state with no error.
type Check struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Config    string `json:"config"`
	Violation string `json:"violation,omitempty"`
}

// Result is the outcome of one check. A complete exploration gives the
// distinct states and the depth of the graph, which no rerun changes. A
// violation gives only its invariant, because TLC can stop at another state.
type Result struct {
	Name      string `json:"name"`
	Exit      int    `json:"exit"`
	Violation string `json:"violation,omitempty"`
	Distinct  int    `json:"distinct,omitempty"`
	Depth     int    `json:"depth,omitempty"`
}

// Record is the outcome of each check of the suite, with the pins and the
// SHA-256 of each input. just check reads it.
type Record struct {
	Version int               `json:"version"`
	Java    string            `json:"java"`
	SHA256  string            `json:"tlc_sha256"`
	Inputs  map[string]string `json:"inputs"`
	Results []Result          `json:"results"`
}

// Options are the folders and the environment of one run.
type Options struct {
	// Root is the absolute root of the repository.
	Root string
	// Dir is the absolute folder of the run. It keeps the jar, the log of
	// each program, and the states of TLC.
	Dir string
	// Jar is a local jar to use in place of the URL, or empty.
	Jar string
	// Env is the whole environment of each program of the run.
	Env []string
}

// Run checks each model of the suite of the repository, then writes the
// record. It returns the results, or an error and no record.
func Run(o Options) ([]Result, error) {
	if !filepath.IsAbs(o.Root) || !filepath.IsAbs(o.Dir) {
		return nil, errors.New("the root and the folder of a model run must be absolute paths")
	}
	if o.Env == nil {
		return nil, errors.New("a model run needs the environment of its programs. Build it with extensionrepo.NewOwnedRun")
	}
	suite, err := readSuite(o.Root)
	if err != nil {
		return nil, err
	}
	before, err := inputs(o.Root, suite)
	if err != nil {
		return nil, err
	}
	jar, err := acquire(o, suite)
	if err != nil {
		return nil, err
	}
	if err := startJava(o, suite); err != nil {
		return nil, err
	}
	var results []Result
	for _, check := range suite.Checks {
		result, err := runCheck(o, suite, jar, check)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	after, err := inputs(o.Root, suite)
	if err != nil {
		return nil, err
	}
	if !maps.Equal(before, after) {
		return nil, fmt.Errorf("%s: an input changed during the run: %s. Fix: run just verify-models again", SuiteFile, changedInputs(before, after)[0])
	}
	record := Record{Version: 1, Java: suite.Java, SHA256: suite.SHA256, Inputs: after, Results: results}
	if err := writeRecord(filepath.Join(o.Root, RecordFile), record); err != nil {
		return nil, err
	}
	return results, nil
}

// changedInputs names each input that differs between two sets of digests.
func changedInputs(old, current map[string]string) []string {
	var changes []string
	for _, name := range slices.Sorted(maps.Keys(old)) {
		switch digest, ok := current[name]; {
		case !ok:
			changes = append(changes, name+" is gone")
		case digest != old[name]:
			changes = append(changes, name+" changed")
		}
	}
	for _, name := range slices.Sorted(maps.Keys(current)) {
		if _, ok := old[name]; !ok {
			changes = append(changes, name+" is new")
		}
	}
	return changes
}

// isFile reports whether path names a file.
func isFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil && info.Mode().IsRegular(), err
}
