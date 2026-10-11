package hello

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
)

func TestTheExtensionMatchesItsManifest(t *testing.T) {
	if err := extension.Check(Extension); err != nil {
		t.Fatal(err)
	}
}
