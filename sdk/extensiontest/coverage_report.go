package extensiontest

import (
	"flag"
	"os"
	"slices"
	"testing"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// reportCoverage prints each deferral of a declared code, and fails t for
// each command, code, or obligation of ## Publication that no scenario
// observed. Run calls it after all
// scenarios finish. A run of one scenario, such as just scenario, skips it,
// because coverage needs every scenario (#422).
func (k *kit) reportCoverage(t *testing.T) {
	if name := os.Getenv(extensionrepo.ScenarioEnv); name != "" {
		if run := flag.Lookup("test.run"); run != nil && run.Value.String() == extensionrepo.ScenarioRun(name) {
			return
		}
	}
	for _, reservation := range k.Reservations() {
		if err := extensionrepo.Notice(reservation); err != nil {
			t.Errorf("extensionrepo.Notice: %v", err)
		}
	}
	for _, problem := range slices.Concat(k.Missing(), k.publicationMissing(), k.acceptanceMissing()) {
		t.Error(problem)
	}
}
