// Package i18n looks up the text that ncly shows a human in the message
// catalogs under locales/, and formats numbers, sizes, and dates for the
// active language. Every catalog entry carries a description that tells a
// translator where the text appears.
package i18n

import (
	"embed"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

//go:embed locales/*.toml
var locales embed.FS

// Pseudo is the pseudo-locale: every catalog string, marked between ⟦ and ⟧,
// so a scenario finds the text that bypasses the catalog.
var Pseudo = language.MustParse("en-XA")

// catalogs lists the languages that have a catalog, English first.
var catalogs = []language.Tag{language.English}

var bundle, matcher = func() (*goi18n.Bundle, language.Matcher) {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	en, err := b.LoadMessageFileFS(locales, "locales/en.toml")
	if err != nil {
		panic(err)
	}
	for _, m := range en.Messages {
		if err := b.AddMessages(Pseudo, pseudoMessage(m)); err != nil {
			panic(err)
		}
	}
	return b, language.NewMatcher(catalogs)
}()

// Match picks the catalog for a language setting, such as fr-CA, fr, or the
// POSIX form fr_CA.UTF-8. A setting with no close catalog, including C and
// POSIX, gets English. The pseudo-locale needs an exact request, so that
// en-GB never selects it.
func Match(setting string) language.Tag {
	setting, _, _ = strings.Cut(setting, ".")
	setting, _, _ = strings.Cut(setting, "@")
	tag, err := language.Parse(strings.ReplaceAll(setting, "_", "-"))
	if err != nil {
		return language.English
	}
	if tag == Pseudo {
		return Pseudo
	}
	_, index, confidence := matcher.Match(tag)
	if confidence == language.No {
		return language.English
	}
	return catalogs[index]
}

// Printer returns catalog text and formats values in one language.
type Printer struct {
	localizer *goi18n.Localizer
	numbers   *message.Printer
}

// New returns a printer for a catalog that Match picked.
func New(tag language.Tag) *Printer { return newPrinter(bundle, tag) }

func newPrinter(b *goi18n.Bundle, tag language.Tag) *Printer {
	return &Printer{
		localizer: goi18n.NewLocalizer(b, tag.String()),
		numbers:   message.NewPrinter(tag),
	}
}

// T returns the message id, with data filling its template. A missing
// message returns the id itself, so a scenario shows it.
func (p *Printer) T(id string, data ...map[string]any) string {
	cfg := &goi18n.LocalizeConfig{MessageID: id}
	if len(data) > 0 {
		cfg.TemplateData = data[0]
	}
	return p.localize(cfg)
}

// N returns the plural form of the message id that fits count, with data
// filling its template. Count in the template is always count, formatted.
func (p *Printer) N(id string, count int, data ...map[string]any) string {
	fill := map[string]any{}
	if len(data) > 0 {
		for k, v := range data[0] {
			fill[k] = v
		}
	}
	fill["Count"] = p.Number(float64(count), 0)
	return p.localize(&goi18n.LocalizeConfig{MessageID: id, PluralCount: count, TemplateData: fill})
}

func (p *Printer) localize(cfg *goi18n.LocalizeConfig) string {
	text, err := p.localizer.Localize(cfg)
	var missing *goi18n.MessageNotFoundErr
	if errors.As(err, &missing) || text == "" {
		return cfg.MessageID
	}
	return text
}

// Number formats v with at most the given decimals, such as 1,234.5 in
// English and 1 234,5 in Canadian French.
func (p *Printer) Number(v float64, decimals int) string {
	return p.numbers.Sprint(number.Decimal(v, number.MaxFractionDigits(decimals)))
}

// sizeUnits holds the catalog IDs of decimal size units, as macOS and most
// download tools count them.
var sizeUnits = []string{"size.bytes", "size.kilobytes", "size.megabytes", "size.gigabytes", "size.terabytes"}

// Size formats a byte count with one decimal in the largest unit that keeps
// the number at 1 or more, such as 1.5 MB.
func (p *Printer) Size(bytes int64) string {
	value, unit := float64(bytes), 0
	for value >= 1000 && unit < len(sizeUnits)-1 {
		value /= 1000
		unit++
	}
	decimals := 1
	if unit == 0 {
		decimals = 0
	}
	return p.T(sizeUnits[unit], map[string]any{"N": p.Number(value, decimals)})
}

// Date formats a day as ISO 8601, such as 2026-10-06, in every language:
// it reads the same in English and is the Canadian French standard.
func (p *Printer) Date(t time.Time) string { return t.Format(time.DateOnly) }

// pseudoMessage marks every plural form of an English message.
func pseudoMessage(m *goi18n.Message) *goi18n.Message {
	out := *m
	for _, form := range []*string{&out.Zero, &out.One, &out.Two, &out.Few, &out.Many, &out.Other} {
		if *form != "" {
			*form = pseudo(*form)
		}
	}
	return &out
}

var accents = strings.NewReplacer(
	"a", "á", "b", "ƀ", "c", "ç", "d", "ð", "e", "é", "f", "ƒ", "g", "ĝ", "h", "ĥ", "i", "î",
	"j", "ĵ", "k", "ķ", "l", "ļ", "m", "ɱ", "n", "ñ", "o", "ö", "p", "þ", "q", "ǫ", "r", "ŕ",
	"s", "š", "t", "ţ", "u", "û", "v", "ṽ", "w", "ŵ", "x", "ẋ", "y", "ý", "z", "ž",
	"A", "Á", "C", "Ç", "E", "É", "I", "Î", "N", "Ñ", "O", "Ö", "S", "Š", "U", "Û", "Y", "Ý",
)

// pseudo accents the text outside template actions, lengthens it by about a
// third as translations often do, and marks it between ⟦ and ⟧.
func pseudo(text string) string {
	var b strings.Builder
	b.WriteString("⟦")
	for rest := text; rest != ""; {
		before, action, found := strings.Cut(rest, "{{")
		b.WriteString(accents.Replace(before))
		if !found {
			break
		}
		action, rest, _ = strings.Cut(action, "}}")
		b.WriteString("{{" + action + "}}")
	}
	b.WriteString(strings.Repeat("·", (utf8.RuneCountInString(text)+2)/3))
	b.WriteString("⟧")
	return b.String()
}
