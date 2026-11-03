package runnerdiscard_test

import (
	"path/filepath"
	"testing"

	"github.com/pascalandy/nicely/sdk/analysis/runnerdiscard"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestRunnerStreams loads the cases in the module of the SDK, so they use the
// real runner.Program instead of a copy. history.go fails first on the parent
// of 3e6a4eff.
func TestRunnerStreams(t *testing.T) {
	root := filepath.Join(analysistest.TestData(), "..", "..", "..", "..")
	analysistest.Run(t, root, runnerdiscard.Analyzer, "./sdk/analysis/runnerdiscard/testdata/cases")
}
