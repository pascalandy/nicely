package standalone

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/runner"
)

// tally runs tally-tool and answers the whole number that it prints. The
// scenarios give it the fake child of the test kit as tally-tool.
func tally(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	path := runner.Lookup("tally-tool", os.Getenv("PATH"))
	if path == "" {
		return failure(contract.PrereqMissing, p.T("standalone.tally.missing"), p.T("standalone.tally.missing_hint"))
	}
	var stdout bytes.Buffer
	result, err := runner.Run(ctx, runner.Program{Path: path, Env: runner.Environ(os.Environ(), nil), Stdout: &stdout, Stderr: io.Discard})
	if err != nil || result.Exit != 0 {
		return failure(contract.Runtime, p.T("standalone.tally.failed"), "ncly standalone tally")
	}
	count, err := strconv.Atoi(strings.TrimSpace(stdout.String()))
	if err != nil {
		return failure(contract.Runtime, p.T("standalone.tally.failed"), "ncly standalone tally")
	}
	_, _ = fmt.Fprintln(call.Stdout, p.T("standalone.tally.value", map[string]any{"Value": count}))
	return contract.Outcome{Data: map[string]int{"count": count}}
}

// fetch needs the key of the service example, which it reads through
// Call.Key. It answers the SHA-256 of the key, never the key, and returns
// the problem that Call.Key gives as it is.
func fetch(_ context.Context, call extension.Call) contract.Outcome {
	key, problem := call.Key("example")
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	sum := sha256.Sum256([]byte(key))
	digest := hex.EncodeToString(sum[:])
	_, _ = fmt.Fprintln(call.Stdout, call.Printer.T("standalone.fetch.digest", map[string]any{"Digest": digest}))
	return contract.Outcome{Data: map[string]string{"sha256": digest}}
}

func failure(code contract.Code, message, hint string) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: code, Message: message, Hint: hint}}}
}
