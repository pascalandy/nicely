package extensionrepo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/verification"
)

// verifyModels runs just verify-models: each check of verification/tlc.json
// with the pinned TLC, in a host run, then the record that just check reads
// (COR-104-T8). A local jar must hold the pinned SHA-256 too. The folder of
// the run stays, with the log of each program.
func verifyModels(args []string) result {
	const action = "verify-models"
	root, err := os.Getwd()
	if err != nil {
		return failed(action, err)
	}
	var jar string
	if len(args) == 1 {
		if jar, err = filepath.Abs(args[0]); err != nil {
			return failed(action, err)
		}
	}
	dir, err := os.MkdirTemp("", "ncly-models-")
	if err != nil {
		return failed(action, err)
	}
	if dir, err = filepath.Abs(dir); err != nil {
		return failed(action, err)
	}
	run, err := NewOwnedRun("", Profile{})
	if err != nil {
		return failed(action, err)
	}
	results, err := verification.Run(verification.Options{Root: root, Dir: dir, Jar: jar, Env: run.Env})
	if err != nil {
		return failed(action, fmt.Errorf("%w. The logs of the run are in %s", err, dir))
	}
	r := result{action: action}
	for _, check := range results {
		line := fmt.Sprintf("%s: %s: no error in %d distinct states, depth %d", action, check.Name, check.Distinct, check.Depth)
		if check.Violation != "" {
			line = fmt.Sprintf("%s: %s: the expected violation of %s", action, check.Name, check.Violation)
		}
		r.answer = append(r.answer, line)
	}
	r.answer = append(r.answer, fmt.Sprintf("%s: wrote %s. The logs of the run are in %s", action, verification.RecordFile, dir))
	return r
}

// modelProblems prints each problem of the record of the models, and reports
// whether there was one. just check reads the record with no Java and no
// network.
func modelProblems() bool {
	problems := verification.CheckRecord(".")
	for _, problem := range problems {
		fmt.Fprintln(os.Stderr, problem)
	}
	return len(problems) != 0
}
