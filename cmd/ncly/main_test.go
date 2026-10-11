package main

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
)

// TestEveryBundledExtensionMatchesItsManifest catches a broken bundled
// extension before a release ships it.
func TestEveryBundledExtensionMatchesItsManifest(t *testing.T) {
	for _, e := range bundled {
		if err := extension.Check(e); err != nil {
			t.Error(err)
		}
	}
}
