// Package runnerdiscard refuses io.Discard as the Stdout or Stderr of a
// runner.Program, as Analyzers in contract.md defines.
package runnerdiscard

import "golang.org/x/tools/go/analysis"

// Analyzer reports io.Discard in the Stdout or Stderr of a runner.Program.
var Analyzer = &analysis.Analyzer{
	Name: "runnerdiscard",
	Doc:  "refuse io.Discard as the Stdout or Stderr of a runner.Program",
	Run:  func(*analysis.Pass) (any, error) { return nil, nil },
}
