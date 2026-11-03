// Package analysis holds the Go analyzers of the SDK and runs them through
// go vet -vettool, as Analyzers in contract.md defines. Core's just lint and
// the check of extensionrepo run the same rules, at the Nicely version that
// the module selects.
//
// Add a rule as a package of its own beside runnerdiscard, then list it in
// Analyzers. Its cases repeat the past mistake that the rule prevents, and
// fail first without the rule.
package analysis

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pascalandy/nicely/sdk/analysis/callerenv"
	"github.com/pascalandy/nicely/sdk/analysis/runnerdiscard"
	goanalysis "golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

// Analyzers are the rules that nicelyvet runs.
var Analyzers = []*goanalysis.Analyzer{
	callerenv.Analyzer,
	runnerdiscard.Analyzer,
}

// Main serves go vet -vettool, which gives one package to each call. It
// does not load packages itself, so `go run` of nicelyvet on a package path
// fails.
func Main() {
	unitchecker.Main(Analyzers...)
}

// Check builds nicelyvet from the Nicely module that the current module
// selects, then runs it on patterns through go vet -vettool. It builds into
// a scratch folder, which it removes. The extensionrepo package imports this
// package, so go mod tidy keeps the dependencies of the build.
func Check(patterns ...string) error {
	dir, err := os.MkdirTemp("", "ncly-nicelyvet-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	tool := filepath.Join(dir, "nicelyvet")
	if err := run("build", "-o", tool, "github.com/pascalandy/nicely/sdk/analysis/cmd/nicelyvet"); err != nil {
		return err
	}
	return run(append([]string{"vet", "-vettool=" + tool}, patterns...)...)
}

// run prints the go command, then runs it with the streams of the caller.
func run(args ...string) error {
	fmt.Fprintln(os.Stderr, "go "+strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
