// Package keyed is the Go fixture extension whose commands need a key.
// keyed fetch reads it through Call.Key, and keyed relay passes it to a
// program through sdk/wrap. It stays apart from hello, which verify-nicely
// installs in every run, so that its key never fails the doctor of a run
// that holds none.
package keyed

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/wrap"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

// Extension is the keyed fixture, which the scenarios install as an
// external extension.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Handlers: map[string]extension.Handler{"keyed fetch": fetch, "keyed relay": relay},
}

// fetch needs the key of deepgram, which it reads through Call.Key, except
// under --cached, which answers without it, as a summary-only resume of
// transcript needs no Deepgram key. It answers the SHA-256 of the key,
// never the key, and returns the problem that Call.Key gives as it is.
func fetch(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	if call.Flags["cached"] == "true" {
		_, _ = fmt.Fprintln(call.Stdout, p.T("keyed.fetch.from_cache"))
		return contract.Outcome{Data: map[string]string{"source": "cache"}}
	}
	key, problem := call.Key("deepgram")
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	sum := sha256.Sum256([]byte(key))
	digest := hex.EncodeToString(sum[:])
	_, _ = fmt.Fprintln(call.Stdout, p.T("keyed.fetch.digest", map[string]any{"Digest": digest}))
	return contract.Outcome{Data: map[string]string{"source": "key", "sha256": digest}}
}

// relay passes the key of deepgram to keytool, a program that the scenarios
// put in PATH, through sdk/wrap, which reads the key through Call.Key. It
// answers once keytool exits 0.
func relay(ctx context.Context, call extension.Call) contract.Outcome {
	program := wrap.Program{Name: "keytool", Keys: []string{"deepgram"}}
	return wrap.Run(ctx, call, program, func(end wrap.Ended) contract.Outcome {
		if end.Exit != 0 {
			return end.Failed()
		}
		_, _ = fmt.Fprintln(call.Stdout, call.Printer.T("keyed.relay.done"))
		return contract.Outcome{}
	})
}
