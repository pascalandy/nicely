// Package callerenv refuses a read of the whole environment of the caller in
// the test kit and the repository gates. Only extensionrepo.NewOwnedRun reads
// it, as Analyzers in contract.md defines.
package callerenv

import "golang.org/x/tools/go/analysis"

// Analyzer reports os.Environ and syscall.Environ outside the constructor of
// owned runs. This stub reports nothing.
var Analyzer = &analysis.Analyzer{
	Name: "callerenv",
	Doc:  "refuse os.Environ in the test kit and the repository gates outside extensionrepo.NewOwnedRun",
	Run:  func(*analysis.Pass) (any, error) { return nil, nil },
}

// owners maps each package that the rule reads to the one function there that
// may read the environment of the caller as a whole. An empty name allows
// none.
var owners = map[string]string{
	"github.com/pascalandy/nicely/sdk/extensionrepo": "NewOwnedRun",
	"github.com/pascalandy/nicely/sdk/extensiontest": "",
}
