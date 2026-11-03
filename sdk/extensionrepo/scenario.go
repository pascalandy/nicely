package extensionrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// ScenarioEnv names the scenario that a run of go test runs alone, for just
// scenario and for the repeat of signoff. When -run is ScenarioRun of that
// name, the test kit skips its coverage checks, which need every scenario.
const ScenarioEnv = "NCLY_TEST_SCENARIO"

// ScenarioRun is the -run pattern of go test that selects only the scenario
// testdata/script/<name>.txtar.
func ScenarioRun(name string) string {
	return "^TestExtension$/^" + name + "$"
}

// scenarioRecipe is the one command that ## First failing run can record:
// just scenario <name>. Finish and signoff read the name of the scenario from
// it, and never run the text of the record (#422).
var scenarioRecipe = regexp.MustCompile(`^just scenario ([A-Za-z0-9][A-Za-z0-9_-]*)$`)

// scenarioArgs gives the go test arguments that run the scenario
// testdata/script/<name>.txtar. The test kit runs each scenario as a subtest
// of TestExtension, as test-kit.md in the contract says.
func scenarioArgs(name string) []string {
	return []string{"-run", ScenarioRun(name), "./..."}
}

// runScenario runs just scenario <name>, through the go test of Test.
func runScenario(name string) result {
	if !scenarioRecipe.MatchString("just scenario " + name) {
		return failed("scenario", fmt.Errorf("%q is no name of a scenario. Fix: give the name of a file of testdata/script/ without .txtar, such as inspect", name))
	}
	if _, err := os.Stat(filepath.Join("testdata", "script", name+".txtar")); err != nil {
		return failed("scenario", fmt.Errorf("testdata/script/%s.txtar: %w. Fix: give the name of a scenario of this commit", name, err))
	}
	return test("scenario", scenarioArgs(name), map[string]string{ScenarioEnv: name})
}
