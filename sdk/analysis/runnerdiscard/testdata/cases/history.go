package cases

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/pascalandy/nicely/sdk/runner"
)

// launch repeats the call of extensions/agent/execute.go at b695c8ba, the
// parent of 3e6a4eff. A descendant of Codex kept the discarded stdout open,
// so the call waited for the drain limit after Codex ended. The rule does not
// follow variables, so the stderr variable passes.
func launch(ctx context.Context, codex string, args, env []string, text []byte, verbose io.Writer) error {
	stderr := io.Discard
	if os.Getenv("NCLY_VERBOSE") == "1" {
		stderr = verbose
	}
	_, err := runner.Run(ctx, runner.Program{Path: codex, Args: args, Env: env, Stdin: bytes.NewReader(text), Stdout: io.Discard, Stderr: stderr}) // want `runner.Program.Stdout: use nil or omit the field instead of io.Discard`
	return err
}
