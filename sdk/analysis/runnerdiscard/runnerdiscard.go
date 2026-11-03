// Package runnerdiscard refuses io.Discard as the Stdout or Stderr of a
// runner.Program, as Analyzers in contract.md defines. The runner reads such
// a writer through a pipe. A descendant that keeps the pipe open holds the
// call until the drain limit, as in 3e6a4eff. A nil writer gives the program
// the null device, with no pipe.
//
// The rule reads struct literals and field assignments in production
// files. It skips test files, because a test can exercise a discarded pipe
// on purpose. It follows no variable, so `w := io.Discard` then `Stdout: w`
// passes.
package runnerdiscard

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports io.Discard in the Stdout or Stderr of a runner.Program.
var Analyzer = &analysis.Analyzer{
	Name: "runnerdiscard",
	Doc:  "refuse io.Discard as the Stdout or Stderr of a runner.Program",
	Run:  run,
}

const runnerPath = "github.com/pascalandy/nicely/sdk/runner"

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.KeyValueExpr:
				// A key of a struct literal resolves to its field.
				if key, ok := node.Key.(*ast.Ident); ok {
					check(pass, pass.TypesInfo.ObjectOf(key), node.Value)
				}
			case *ast.AssignStmt:
				if len(node.Lhs) != len(node.Rhs) {
					return true
				}
				for i, target := range node.Lhs {
					if selector, ok := ast.Unparen(target).(*ast.SelectorExpr); ok {
						check(pass, pass.TypesInfo.ObjectOf(selector.Sel), node.Rhs[i])
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// check reports value when it is io.Discard and target is a stream field of
// runner.Program.
func check(pass *analysis.Pass, target types.Object, value ast.Expr) {
	if isStream(target) && isDiscard(pass, value) {
		pass.Reportf(value.Pos(), "runner.Program.%s: use nil or omit the field instead of io.Discard; the runner reads a discard writer through a pipe that a descendant can keep open", target.Name())
	}
}

// isStream reports whether object is the Stdout or Stderr field of
// runner.Program, which an alias, a pointer, or an embedding reaches too.
func isStream(object types.Object) bool {
	field, ok := object.(*types.Var)
	if !ok || !field.IsField() || field.Pkg() == nil || field.Pkg().Path() != runnerPath {
		return false
	}
	if field.Name() != "Stdout" && field.Name() != "Stderr" {
		return false
	}
	program, ok := field.Pkg().Scope().Lookup("Program").(*types.TypeName)
	if !ok {
		return false
	}
	fields, ok := program.Type().Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for i := range fields.NumFields() {
		if fields.Field(i) == field {
			return true
		}
	}
	return false
}

func isDiscard(pass *analysis.Pass, value ast.Expr) bool {
	var object types.Object
	switch value := ast.Unparen(value).(type) {
	case *ast.SelectorExpr:
		object = pass.TypesInfo.ObjectOf(value.Sel)
	case *ast.Ident:
		object = pass.TypesInfo.ObjectOf(value)
	}
	_, isVar := object.(*types.Var)
	return isVar && object.Pkg() != nil && object.Pkg().Path() == "io" && object.Name() == "Discard"
}
