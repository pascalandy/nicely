package standalone

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extensiontest"
)

func TestExtension(t *testing.T) {
	extensiontest.Run(t, Extension)
}
