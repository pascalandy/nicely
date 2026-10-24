// Package video is an independent named extension fixture for the test kit.
package video

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

//go:embed skill
var skill embed.FS

// Extension exposes one file-taking root through ncly-video-convert.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{"video convert": convert},
}

func convert(_ context.Context, call extension.Call) contract.Outcome {
	key, problem := call.Key("example")
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	sum := sha256.Sum256([]byte(key))
	_, _ = fmt.Fprintln(call.Stdout, call.Printer.T("video-convert.checked", map[string]any{"File": call.Args[0]}))
	return contract.Outcome{Data: map[string]string{"file": call.Args[0], "sha256": hex.EncodeToString(sum[:])}}
}
