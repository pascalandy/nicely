// Package i18n looks up the text that ncly shows a human in the message
// catalogs under locales/. Every catalog entry carries a description that
// tells a translator where the text appears.
package i18n

import (
	"embed"
	"errors"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var locales embed.FS

var bundle = func() *goi18n.Bundle {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	if _, err := b.LoadMessageFileFS(locales, "locales/en.toml"); err != nil {
		panic(err)
	}
	return b
}()

// Printer returns catalog text in one language.
type Printer struct {
	localizer *goi18n.Localizer
}

// New returns a printer for a language tag. English is the only catalog
// until language matching arrives, so every tag selects it.
func New(string) *Printer {
	return &Printer{localizer: goi18n.NewLocalizer(bundle, "en")}
}

// T returns the message id, with data filling its template. A missing
// message returns the id itself, so a scenario shows it.
func (p *Printer) T(id string, data ...map[string]any) string {
	cfg := &goi18n.LocalizeConfig{MessageID: id}
	if len(data) > 0 {
		cfg.TemplateData = data[0]
	}
	text, err := p.localizer.Localize(cfg)
	var missing *goi18n.MessageNotFoundErr
	if errors.As(err, &missing) || text == "" {
		return id
	}
	return text
}
