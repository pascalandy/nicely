// Package callerenv refuses whole caller environment reads in the test kit,
// repository gates, TLC runner and their child packages.
// The exact allowed readers are listed below and in Analyzers of contract.md.
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
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports os.Environ and syscall.Environ outside the exact allowed
// readers in the owned package trees.
var Analyzer = &analysis.Analyzer{
	Name: "callerenv",
	Doc:  "refuse whole caller environment reads in owned SDK package trees outside their exact allowed readers",
	Run:  run,
}

// Each key owns its package and descendants. Its allowed function names apply
// only to that exact package, never to descendants.
var owners = map[string][]string{
	"github.com/pascalandy/nicely/sdk/extensionrepo": {"NewOwnedRun", "updateEnviron"},
	"github.com/pascalandy/nicely/sdk/extensiontest": nil,
	"github.com/pascalandy/nicely/sdk/verification":  nil,
	// stage preserves the isolated scenario environment across syscall.Exec.
	"github.com/pascalandy/nicely/sdk/extensiontest/guard": {"stage"},
	// run implements the fixture's env action, which prints names, never values.
	"github.com/pascalandy/nicely/sdk/extensiontest/fake": {"run"},
}

func run(pass *analysis.Pass) (any, error) {
	path := pass.Pkg.Path()
	owned := false
	for parent := range owners {
		if path == parent || strings.HasPrefix(path, parent+"/") {
			owned = true
			break
		}
	}
	if !owned {
		return nil, nil
	}
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && slices.Contains(owners[path], fn.Name.Name) {
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
