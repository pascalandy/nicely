// Package agent is the external extension that runs one trusted task
// through Codex, as extensions/agent/spec.md defines. It imports only sdk/,
// and installs as ncly-agent with its manifest, catalogs, and skill beside it.
package agent

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

// Extension is the agent extension, which cmd/ncly-agent runs.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{
		"agent run": run,
	},
}

// run answers that this build prepares no task yet.
func run(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	return contract.Outcome{Errors: []contract.Problem{{
		Code:    contract.CapabilityUnsupported,
		Message: p.T("agent.run.unavailable"),
		Hint:    p.T("agent.run.unavailable_hint"),
	}}}
}
