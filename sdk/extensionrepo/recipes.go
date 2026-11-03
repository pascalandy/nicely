package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pascalandy/nicely/sdk/analysis"
	"github.com/pascalandy/nicely/sdk/verification"
)

// checkRepository runs just check. Its go test runs every scenario, so the
// coverage checks run too, because the owned run drops NCLY_TEST_SCENARIO.
func checkRepository() result {
	if err := printRuleIndex(); err != nil {
		return failed("check", err)
	}
	if modelProblems() {
		return failed("check", errors.New("the record of the models is not that of "+verification.SuiteFile))
	}
	if err := updateRun("go", "mod", "tidy", "-diff"); err != nil {
		return failed("check", err)
	}
	files, err := exec.Command("gofmt", "-l", ".").Output()
	if err != nil {
		return failed("check", err)
	}
	if len(files) != 0 {
		fmt.Fprint(os.Stderr, string(files))
		return failed("check", errors.New("run gofmt -w . to format these files"))
	}
	if err := updateRun("go", "vet", "./..."); err != nil {
		return failed("check", err)
	}
	if err := analysis.Check("./..."); err != nil {
		return failed("check", err)
	}
	fmt.Fprintln(os.Stderr, "go test -count=1 ./...")
	return test("check", nil, nil)
}

// contractDirectory answers just contract with the folder of the contract in
// the pinned Nicely module.
func contractDirectory() result {
	if err := updateRun("go", "mod", "download", nicelyModule); err != nil {
		return failed("contract", err)
	}
	dir, err := updateOutput("go", "list", "-m", "-f", "{{.Dir}}", nicelyModule)
	if err != nil {
		return failed("contract", err)
	}
	path := filepath.Join(strings.TrimSpace(string(dir)), "docs", "north-star", "contract")
	info, err := os.Stat(path)
	if err != nil {
		return failed("contract", err)
	}
	if !info.IsDir() {
		return failed("contract", fmt.Errorf("%s is not a directory", path))
	}
	return result{action: "contract", answer: []string{path}}
}
