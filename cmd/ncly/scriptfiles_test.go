package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/txtar"
)

// scriptDirs returns the folders of the scenarios under root: core's, then
// those of each extension and of each fixture extension. TestScripts runs
// them, and TestFailedReportsUseAnswer reads them.
func scriptDirs(root string) (core string, others []string) {
	extensions, _ := filepath.Glob(filepath.Join(root, "extensions", "*", "testdata", "script"))
	fixtureDirs, _ := filepath.Glob(filepath.Join(root, "testdata", "extensions", "*", "testdata", "script"))
	return filepath.Join(root, "testdata", "script"), append(extensions, fixtureDirs...)
}

// failedReport starts the envelope of a failed report, as a scenario's regex
// writes it.
const failedReport = `\{"ok":false,"contract_version":1`

// stdoutLine matches a script line that runs stdout or ! stdout, after any
// [condition] prefix, in the order that testscript reads them.
var stdoutLine = regexp.MustCompile(`^(\[\S+\]\s+)*(!\s+)?stdout\s`)

// TestFailedReportsUseAnswer fails on a scenario that checks a failed report
// with a stdout regex. testscript compiles that regex with (?m), so ^ and $
// match one line. A second line or an errors key then passes unseen, as #199
// found. answer stdout report reads the whole stream instead. No scenario can
// check the other scenarios, so this test reads the folders of TestScripts.
// A folder without a scenario fails it, so a wrong list cannot pass empty.
func TestFailedReportsUseAnswer(t *testing.T) {
	root := filepath.Join("..", "..")
	core, others := scriptDirs(root)
	if len(others) == 0 {
		t.Fatalf("no scenario folder of an extension under %s", root)
	}
	for _, dir := range append([]string{core}, others...) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		read := 0
		for _, entry := range entries {
			// testscript runs each .txt and .txtar file of the folder.
			if !strings.HasSuffix(entry.Name(), ".txtar") && !strings.HasSuffix(entry.Name(), ".txt") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			name, err := filepath.Rel(root, filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, finding := range failedReportRegexes(filepath.ToSlash(name), data) {
				t.Error(finding)
			}
			read++
		}
		if read == 0 {
			t.Fatalf("no scenario in %s", dir)
		}
	}
}

// failedReportRegexes lists each line of a scenario's script that checks a
// failed report with a stdout regex. It reads only the script, before the
// first file of the archive, so a fixture may still print such a report.
// It skips a comment, which may name such a regex.
func failedReportRegexes(name string, data []byte) []string {
	var found []string
	for i, line := range strings.Split(string(txtar.Parse(data).Comment), "\n") {
		line = strings.TrimSpace(withoutComment(line))
		if stdoutLine.MatchString(line) && strings.Contains(line, failedReport) {
			found = append(found, fmt.Sprintf("%s:%d: a stdout regex checks a failed report: use answer stdout report", name, i+1))
		}
	}
	return found
}

// withoutComment cuts a script line at its first # outside single quotes,
// where testscript ends the line.
func withoutComment(line string) string {
	quoted := false
	for i, c := range line {
		switch {
		case c == '\'':
			quoted = !quoted
		case c == '#' && !quoted:
			return line[:i]
		}
	}
	return line
}

// TestFailedReportRegexes proves what the check flags and what it allows.
// No scenario of main holds such a line, so TestFailedReportsUseAnswer alone
// proves nothing about a regex that it should flag.
func TestFailedReportRegexes(t *testing.T) {
	cases := []struct {
		name, script string
		// lines lists the flagged lines, by number.
		lines []int
	}{
		{"an anchored regex", "exec ncly doctor\n" + `stdout '^\{"ok":false,"contract_version":1,"checks":\[\]\}$'` + "\n", []int{2}},
		{"a regex without an anchor", `stdout '\{"ok":false,"contract_version":1,'` + "\n", []int{1}},
		{"a negated regex", `! stdout '\A\{"ok":false,"contract_version":1,'` + "\n", []int{1}},
		{"a regex under conditions", `[linux] [!short] ! stdout '\{"ok":false,"contract_version":1'` + "\n", []int{1}},
		{"a failure on stderr", `stderr '^\{"ok":false,"contract_version":1,"errors"'` + "\n", nil},
		{"a success on stdout", `stdout '^\{"ok":true,"contract_version":1,'` + "\n", nil},
		{"a comment", `# stdout '\{"ok":false,"contract_version":1'` + "\n", nil},
		{"a success with a comment after it", `stdout '^\{"ok":true,"contract_version":1\}$' # not \{"ok":false,"contract_version":1` + "\n", nil},
		{"a quoted hash", `stdout '\{"ok":false,"contract_version":1,"hint":"#'` + "\n", []int{1}},
		{"another command", `exec echo stdout '\{"ok":false,"contract_version":1'` + "\n", nil},
		{"a fixture", "exec sh report.sh\n-- report.sh --\n" + `stdout '\{"ok":false,"contract_version":1'` + "\n", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var want []string
			for _, line := range c.lines {
				want = append(want, fmt.Sprintf("x.txtar:%d: a stdout regex checks a failed report: use answer stdout report", line))
			}
			if got := failedReportRegexes("x.txtar", []byte(c.script)); !slices.Equal(got, want) {
				t.Errorf("failedReportRegexes(%q) = %q, want %q", c.script, got, want)
			}
		})
	}
}
