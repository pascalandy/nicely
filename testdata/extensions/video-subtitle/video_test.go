package video

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
)

func TestExtension(t *testing.T) {
	if err := extension.Check(Extension); err != nil {
		t.Fatal(err)
	}
}
