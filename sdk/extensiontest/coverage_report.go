package extensiontest

import (
	"os"
	"testing"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// reportCoverage prints each deferral of a declared code, and fails t for
// each command or code that no scenario observed. Run calls it after all
// scenarios finish. A run of one scenario, such as just scenario, skips it,
// because coverage needs every scenario (#422).
func (k *kit) reportCoverage(t *testing.T) {
	if os.Getenv(extensionrepo.ScenarioEnv) != "" {
		return
	}
	for _, reservation := range k.Reservations() {
		if err := extensionrepo.Notice(reservation); err != nil {
			t.Errorf("extensionrepo.Notice: %v", err)
		}
	}
	for _, problem := range k.Missing() {
		t.Error(problem)
	}
}
