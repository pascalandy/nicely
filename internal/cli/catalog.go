package cli

import (
	"embed"
	"io/fs"

	"github.com/pascalandy/nicely/sdk/i18n"
)

//go:embed locales/*.toml
var locales embed.FS

// catalogs holds core's catalogs beside the SDK's, whose text, such as help
// pages and usage errors, core's commands share.
var catalogs = func() *i18n.Catalogs {
	own, err := fs.Sub(locales, "locales")
	if err == nil {
		var c *i18n.Catalogs
		if c, err = i18n.Load(i18n.Locales(), own); err == nil {
			return c
		}
	}
	panic(err)
}()
