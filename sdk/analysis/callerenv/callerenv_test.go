package callerenv

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// root is the module of the SDK, so the cases use the real testscript.Env.
var root = filepath.Join("..", "..", "..")

// TestCallerEnvironment reads the cases as one more owned package, whose
// constructor is NewOwnedRun. history.go fails first on the parent of
// 09809608, and notices.go on the go test of e35bd0bb.
func TestCallerEnvironment(t *testing.T) {
	cases := "github.com/pascalandy/nicely/sdk/analysis/callerenv/testdata/cases"
	owners[cases] = "NewOwnedRun"
	t.Cleanup(func() { delete(owners, cases) })
	analysistest.Run(t, root, Analyzer, "./sdk/analysis/callerenv/testdata/cases")
}

// TestOwnedPackages requires that the kit and the repository gates read the
// environment of the caller only in NewOwnedRun.
func TestOwnedPackages(t *testing.T) {
	analysistest.Run(t, root, Analyzer, "./sdk/extensionrepo", "./sdk/extensiontest")
}
