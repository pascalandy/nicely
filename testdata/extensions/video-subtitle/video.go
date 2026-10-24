// Package video is the named fixture that supplies sibling completion values.
package video

import (
	"context"
	"embed"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

//go:embed skill
var skill embed.FS

// Extension groups a list command beneath a named root.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{
		"video subtitle list": func(_ context.Context, _ extension.Call) contract.Outcome {
			return contract.Outcome{Data: map[string]any{"files": []map[string]string{{"name": "clip.mov"}, {"name": "other.mov"}}}}
		},
		"video subtitle fetch": func(_ context.Context, call extension.Call) contract.Outcome {
			if _, problem := call.Key("deepgram"); problem != nil {
				return contract.Outcome{Errors: []contract.Problem{*problem}}
			}
			return contract.Outcome{Data: map[string]bool{"received_key": true}}
		},
		"video subtitle status": func(_ context.Context, call extension.Call) contract.Outcome {
			return contract.Outcome{Data: map[string]any{"checks": []map[string]string{{
				"id": "video-subtitle.ready", "component": "video-subtitle", "status": "pass",
				"message": call.Printer.T("video-subtitle.ready"),
			}}}}
		},
	},
}
