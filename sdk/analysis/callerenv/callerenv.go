// Package callerenv refuses a read of the whole environment of the caller in
// the test kit and the repository gates. Only extensionrepo.NewOwnedRun reads
// it, as Analyzers in contract.md defines.
//
// At the parent of 09809608, the lifecycle check of the kit passed
// os.Environ() to a scratch install, which replaced HOME and kept an invalid
// NCLY_CONFIG. Each other caller filtered its own copy. The rule reads each
// use of os.Environ or syscall.Environ, a call or a function value, in
// production files. It skips test files, because a test can read its caller
// on purpose. A program that starts with no Env inherits the caller too, and
// the rule cannot see that.
package callerenv

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports os.Environ and syscall.Environ outside the constructor of
// owned runs.
var Analyzer = &analysis.Analyzer{
	Name: "callerenv",
	Doc:  "refuse os.Environ in the test kit and the repository gates outside extensionrepo.NewOwnedRun",
	Run:  run,
}

// owners maps each package that the rule reads to the one function there that
// may read the environment of the caller as a whole. An empty name allows
// none.
var owners = map[string]string{
	"github.com/pascalandy/nicely/sdk/extensionrepo": "NewOwnedRun",
	"github.com/pascalandy/nicely/sdk/extensiontest": "",
}

func run(pass *analysis.Pass) (any, error) {
	owner, ok := owners[pass.Pkg.Path()]
	if !ok {
		return nil, nil
	}
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && owner != "" && fn.Name.Name == owner {
				continue
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				if ident, ok := node.(*ast.Ident); ok {
					check(pass, ident)
				}
				return true
			})
		}
	}
	return nil, nil
}

// check reports ident when it names os.Environ or syscall.Environ.
func check(pass *analysis.Pass, ident *ast.Ident) {
	fn, ok := pass.TypesInfo.Uses[ident].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Name() != "Environ" {
		return
	}
	if path := fn.Pkg().Path(); path == "os" || path == "syscall" {
		pass.Reportf(ident.Pos(), "%s.Environ: build the environment with extensionrepo.NewOwnedRun, which drops the caller state; a run that only replaces HOME keeps it, as at the parent of 09809608", fn.Pkg().Name())
	}
}
