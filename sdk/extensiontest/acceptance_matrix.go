package extensiontest

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
	"github.com/rogpeppe/go-internal/testscript"
)

// outputCases are the cases of the Output group of Generic cases in the
// contract, by their names there.
var outputCases = []string{
	"Exists, without `--force`",
	"Is a link to the input",
	"Appears during the run, without `--force`",
	"Temporary file",
	"Keeps private data of the input",
	"Keeps data that a reader needs",
	"Loses data that changes the meaning",
}

// appearsCase needs no case of reads, because the obligation appears of
// ## Publication proves it. privateCases need a private text.
var (
	appearsCase  = caseName("Appears during the run, without `--force`")
	privateCases = []string{caseName("Keeps private data of the input"), caseName("Keeps data that a reader needs")}
)

// caseName is the name of a case without its backticks, so a scenario and
// a cell of spec.md can quote a flag or not.
func caseName(text string) string { return strings.ReplaceAll(text, "`", "") }

// An outputCase is an Output case that a cell Source of ## Case matrix in
// spec.md names, at its line.
type outputCase struct {
	name string
	line int
}

// acceptance is what reads records across parallel scenarios, as
// Acceptance cases in the contract defines. It holds each scenario that
// called the extension, and each Output case that a case proved.
type acceptance struct {
	mu     sync.Mutex
	called map[*testscript.TestScript]bool
	proved map[string]bool
}

// markCalled records that a call of the extension ran in ts.
func (a *acceptance) markCalled(ts *testscript.TestScript) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.called[ts] {
		a.called[ts] = true
		ts.Defer(func() {
			a.mu.Lock()
			defer a.mu.Unlock()
			delete(a.called, ts)
		})
	}
}

func (a *acceptance) wasCalled(ts *testscript.TestScript) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.called[ts]
}

func (a *acceptance) prove(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.proved[name] = true
}

func (a *acceptance) wasProved(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.proved[name]
}

// matrixOutputs returns each Output case that a cell Source of
// ## Case matrix in spec names. A cell names one as Output: <case>, and a
// cell can name more than one.
func matrixOutputs(spec string) []outputCase {
	var found []outputCase
	header := false
	for _, l := range extensionrepo.SpecLines(spec) {
		switch {
		case l.Part != "Case matrix":
		case l.Kind == extensionrepo.SpecHeader:
			header = slices.Equal(l.Cells, []string{"Family", "Source", "Accepted", "Refused"})
		case l.Kind == extensionrepo.SpecRow && header && len(l.Cells) == 4 && strings.Contains(l.Cells[1], "Output"):
			for _, name := range outputCases {
				if strings.Contains(caseName(l.Cells[1]), caseName(name)) && !slices.ContainsFunc(found, func(c outputCase) bool { return c.name == caseName(name) }) {
					found = append(found, outputCase{name: caseName(name), line: l.Number})
				}
			}
		}
	}
	return found
}

// acceptanceMissing reports each Output case of ## Case matrix that no
// passing reads -case proved, except the case that appears proves. Run
// reports them after all scenarios, with coverage.
func (k *kit) acceptanceMissing() []string {
	var problems []string
	for _, o := range k.outputs {
		if o.name != appearsCase && !k.acceptance.wasProved(o.name) {
			problems = append(problems, fmt.Sprintf("spec.md:%d: ## Case matrix names the Output case %q, and no scenario proves it with a passing reads -case. "+
				"Fix: make its test file in a scenario, run the command, and check the output with reads -case '%s'", o.line, o.name, o.name))
		}
	}
	return problems
}
