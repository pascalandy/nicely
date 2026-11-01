package extensiontest

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// reportCoverage prints each deferral of a declared code, and fails t for
// each command or code that no scenario observed. Run calls it after all
// scenarios finish.
func (k *kit) reportCoverage(t *testing.T) {
	for _, reservation := range k.Reservations() {
		if err := extensionrepo.Notice(reservation); err != nil {
			t.Errorf("extensionrepo.Notice: %v", err)
		}
	}
	for _, problem := range k.Missing() {
		t.Error(problem)
	}
}
