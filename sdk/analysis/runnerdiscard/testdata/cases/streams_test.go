package cases

import (
	"io"

	"github.com/pascalandy/nicely/sdk/runner"
)

// A test can discard the output of a program through a pipe on purpose.
var _ = runner.Program{Stdout: io.Discard}
