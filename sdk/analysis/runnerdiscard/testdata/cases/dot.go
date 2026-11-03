package cases

import (
	. "io"

	"github.com/pascalandy/nicely/sdk/runner"
)

var _ = runner.Program{Stderr: Discard} // want `runner.Program.Stderr: use nil or omit the field instead of io.Discard`
