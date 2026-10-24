// Package video is the named fixture whose root converts one file.
package video

import (
	"context"
	"embed"
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

// Extension has one executable root, without descendants or prerequisites.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{"video convert": convert},
}

func convert(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	if call.Args[0] == "fail.mov" {
		return contract.Outcome{Errors: []contract.Problem{{Code: "VIDEO_CONVERT_FAILED", Message: p.T("video-convert.failed"), Hint: p.T("video-convert.hint")}}}
	}
	out := contract.Outcome{Data: map[string]any{"path": call.Path, "file": call.Args[0]}}
	if call.Args[0] == "warn.mov" {
		out.Warnings = []contract.Problem{{Code: "VIDEO_CONVERT_WARNING", Message: p.T("video-convert.failed"), Hint: p.T("video-convert.hint")}}
	}
	if !call.JSON {
		_, _ = fmt.Fprintln(call.Stdout, p.T("video-convert.converted", map[string]any{"File": call.Args[0]}))
	}
	return out
}
