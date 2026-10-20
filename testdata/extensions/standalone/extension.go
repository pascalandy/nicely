// Package standalone is the fixture extension repository of the test kit.
// It is one Go module outside Nicely, laid out as Extension repository in
// contract.md says. Its go test runs sdk/extensiontest.
package standalone

import (
	"embed"

	"github.com/pascalandy/nicely/sdk/extension"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

//go:embed skill
var skill embed.FS

// Extension is the standalone extension, which cmd/ncly-standalone runs.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Skill:    skill,
	Handlers: map[string]extension.Handler{
		"standalone tally": tally,
		"standalone fetch": fetch,
	},
}
