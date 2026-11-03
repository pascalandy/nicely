package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/platform"
)

var (
	sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	namePattern   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	invariant     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

const recordFix = "Fix: run just verify-models, then commit " + RecordFile

// decodeStrict decodes one JSON value, and refuses a key that v does not
// define. So no pin or outcome hides in a key that nothing reads.
func decodeStrict(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("it holds more than one JSON value")
	}
	return nil
}

// readSuite reads and checks the suite of root.
func readSuite(root string) (Suite, error) {
	var suite Suite
	data, err := os.ReadFile(filepath.Join(root, SuiteFile))
	if err != nil {
		return suite, err
	}
	if err := decodeStrict(data, &suite); err != nil {
		return suite, fmt.Errorf("%s: %w", SuiteFile, err)
	}
	var problems []string
	if !strings.HasPrefix(suite.Java, "java@") || strings.ContainsAny(suite.Java, " \t\n") {
		problems = append(problems, fmt.Sprintf("java %q is no tool of mise, such as java@temurin-21.0.13+9.0.LTS", suite.Java))
	}
	if !strings.HasPrefix(suite.URL, "https://") && !strings.HasPrefix(suite.URL, "http://") {
		problems = append(problems, fmt.Sprintf("tlc_url %q is no HTTP URL", suite.URL))
	}
	if !sha256Pattern.MatchString(suite.SHA256) {
		problems = append(problems, "tlc_sha256 holds no SHA-256 of 64 lowercase hexadecimal digits")
	}
	if len(suite.Checks) == 0 {
		problems = append(problems, "it lists no check")
	}
	seen := map[string]bool{}
	for _, c := range suite.Checks {
		if !namePattern.MatchString(c.Name) {
			problems = append(problems, fmt.Sprintf("the check %q has no name of letters, digits, - and _", c.Name))
		}
		if seen[c.Name] {
			problems = append(problems, "it names the check "+c.Name+" twice")
		}
		seen[c.Name] = true
		for _, p := range [][2]string{{c.Model, ".tla"}, {c.Config, ".cfg"}} {
			if !filepath.IsLocal(p[0]) || path.Clean(p[0]) != p[0] || path.Ext(p[0]) != p[1] {
				problems = append(problems, fmt.Sprintf("the check %s names %q, which is no clean %s file of the repository", c.Name, p[0], p[1]))
			} else if ok, err := isFile(filepath.Join(root, p[0])); !ok {
				problems = append(problems, fmt.Sprintf("the check %s names %s, which is no file: %v", c.Name, p[0], err))
			}
		}
		if c.Violation != "" && !invariant.MatchString(c.Violation) {
			problems = append(problems, fmt.Sprintf("the check %s expects the violation of %q, which is no TLA+ name", c.Name, c.Violation))
		}
	}
	if len(problems) != 0 {
		return suite, fmt.Errorf("%s: %s", SuiteFile, strings.Join(problems, "; "))
	}
	return suite, nil
}

// inputs returns the SHA-256 of each input of the suite. The inputs are the
// suite, and each .tla and .cfg file of the folder of each model and each
// configuration. So a module that a model extends is an input too.
func inputs(root string, suite Suite) (map[string]string, error) {
	files := []string{SuiteFile}
	var dirs []string
	for _, c := range suite.Checks {
		dirs = append(dirs, path.Dir(c.Model), path.Dir(c.Config))
	}
	slices.Sort(dirs)
	for _, dir := range slices.Compact(dirs) {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if ext := path.Ext(e.Name()); e.Type().IsRegular() && (ext == ".tla" || ext == ".cfg") {
				files = append(files, path.Join(dir, e.Name()))
			}
		}
	}
	digests := map[string]string{}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		digests[file] = hex.EncodeToString(sum[:])
	}
	return digests, nil
}

// writeRecord replaces the record durably, so a failed write keeps the
// record before it.
func writeRecord(file string, record Record) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return platform.WriteFile(file, append(data, '\n'), 0o644)
}

// CheckRecord reports each problem of the record of root, which just check
// runs. It reads only the suite, its inputs, and the record, and starts no
// program. A repository without a suite has no problem.
func CheckRecord(root string) []string {
	if ok, err := isFile(filepath.Join(root, SuiteFile)); err != nil {
		return []string{SuiteFile + ": " + err.Error()}
	} else if !ok {
		return nil
	}
	suite, err := readSuite(root)
	if err != nil {
		return []string{err.Error()}
	}
	current, err := inputs(root, suite)
	if err != nil {
		return []string{SuiteFile + ": " + err.Error()}
	}
	data, err := os.ReadFile(filepath.Join(root, RecordFile))
	if errors.Is(err, os.ErrNotExist) {
		return []string{fmt.Sprintf("%s: %s lists models, and no record holds their outcomes. %s", RecordFile, SuiteFile, recordFix)}
	} else if err != nil {
		return []string{RecordFile + ": " + err.Error()}
	}
	var record Record
	if err := decodeStrict(data, &record); err != nil {
		return []string{fmt.Sprintf("%s: %v. %s", RecordFile, err, recordFix)}
	}
	at := RecordFile + ": "
	var problems []string
	if record.Version != 1 {
		problems = append(problems, fmt.Sprintf("%sthe version %d is not 1. %s", at, record.Version, recordFix))
	}
	if record.SHA256 != suite.SHA256 || record.Java != suite.Java {
		problems = append(problems, fmt.Sprintf("%sthe record is of TLC %s with %s, and %s pins TLC %s with %s. A new pin needs a review, then a run. %s",
			at, record.SHA256, record.Java, SuiteFile, suite.SHA256, suite.Java, recordFix))
	}
	if changes := changedInputs(record.Inputs, current); len(changes) != 0 {
		problems = append(problems, fmt.Sprintf("%sthe record is of other inputs: %s. %s", at, strings.Join(changes, ", "), recordFix))
	}
	return append(problems, resultProblems(suite, record.Results)...)
}

// resultProblems reports each result that is not the expected outcome of
// its check, in the order of the suite.
func resultProblems(suite Suite, results []Result) []string {
	at := RecordFile + ": "
	same := len(results) == len(suite.Checks)
	for i := 0; same && i < len(results); i++ {
		same = results[i].Name == suite.Checks[i].Name
	}
	if !same {
		return []string{fmt.Sprintf("%sthe results are not those of the checks of %s, in its order. %s", at, SuiteFile, recordFix)}
	}
	var problems []string
	for i, c := range suite.Checks {
		r, want := results[i], expected(c)
		switch {
		case r.Exit != want || r.Violation != c.Violation:
			problems = append(problems, fmt.Sprintf("%sthe check %s gives exit %d and the violation %q, and %s expects exit %d and %q. %s", at, c.Name, r.Exit, r.Violation, SuiteFile, want, c.Violation, recordFix))
		case c.Violation == "" && (r.Distinct < 1 || r.Depth < 1):
			problems = append(problems, fmt.Sprintf("%sthe check %s gives no statistics of a complete graph. %s", at, c.Name, recordFix))
		case c.Violation != "" && (r.Distinct != 0 || r.Depth != 0):
			problems = append(problems, fmt.Sprintf("%sthe check %s keeps statistics of a violation, which a rerun can change. %s", at, c.Name, recordFix))
		}
	}
	return problems
}

// expected is the exit code of TLC for the outcome that c expects. It is 0
// for a complete exploration, and 12 for the violation of a safety property.
func expected(c Check) int {
	if c.Violation != "" {
		return 12
	}
	return 0
}
