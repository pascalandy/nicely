// Package i18n looks up the text that ncly shows a human in the message
// catalogs under locales/, and formats numbers, sizes, and dates for the
// active language. Every catalog entry carries a description that tells a
// translator where the text appears.
package i18n

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n/template"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

//go:embed locales/*.toml
var locales embed.FS

// Pseudo is the pseudo-locale: every catalog string, marked between ⟦ and ⟧,
// so a scenario finds the text that bypasses the catalog.
var Pseudo = language.MustParse("en-XA")

var core = func() *Catalogs {
	sub, err := fs.Sub(locales, "locales")
	if err == nil {
		var c *Catalogs
		if c, err = Load(sub); err == nil {
			return c
		}
	}
	panic(err)
}()

// Match picks core's catalog for a language setting, such as fr-CA, fr, or
// the POSIX form fr_CA.UTF-8, as Catalogs.Match does.
func Match(setting string) language.Tag { return core.Match(setting) }

// New returns a printer of core's catalogs for a tag that Match picked.
func New(tag language.Tag) *Printer { return newPrinter(core.bundle, tag) }

// Catalogs holds the catalogs of one owner, such as core or an extension:
// one file per language, named after its tag, such as en.toml or
// fr-CA.toml. The en catalog is required, and the pseudo-locale marks it.
type Catalogs struct {
	bundle  *goi18n.Bundle
	tags    []language.Tag // the languages with a catalog, en first
	matcher language.Matcher
	en      map[string]bool
}

// CatalogError reports an en catalog that is missing, unreadable, or not a
// catalog. Line is 0 when the parser does not say where.
type CatalogError struct {
	File string
	Line int
	Err  error
}

func (e *CatalogError) Error() string { return fmt.Sprintf("%s: %v", e.File, e.Err) }

func (e *CatalogError) Unwrap() error { return e.Err }

// Load reads the catalogs at the top of fsys. Only the en catalog must be
// valid: a translation that cannot be read is skipped, so its entries fall
// back to en.
func Load(fsys fs.FS) (*Catalogs, error) {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	en, err := b.LoadMessageFileFS(fsys, "en.toml")
	if err != nil {
		e := &CatalogError{File: "en.toml", Err: err}
		var parse toml.ParseError
		if errors.As(err, &parse) {
			e.Line = parse.Position.Line
		}
		return nil, e
	}
	c := &Catalogs{bundle: b, tags: []language.Tag{language.English}, en: map[string]bool{}}
	for _, m := range en.Messages {
		if err := checkMessage(m); err != nil {
			return nil, &CatalogError{File: "en.toml", Err: err}
		}
		c.en[m.ID] = true
		if err := b.AddMessages(Pseudo, pseudoMessage(m)); err != nil {
			return nil, &CatalogError{File: "en.toml", Err: err}
		}
	}
	files, _ := fs.Glob(fsys, "*.toml")
	for _, file := range files {
		if file == "en.toml" {
			continue
		}
		// A translation that cannot be read is skipped, and so is the
		// pseudo-locale, which ncly writes itself. An entry that
		// checkMessage refuses is left out, so it falls back to en.
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			continue
		}
		f, err := goi18n.ParseMessageFileBytes(data, file, map[string]goi18n.UnmarshalFunc{"toml": toml.Unmarshal})
		if err != nil || f.Tag == Pseudo {
			continue
		}
		valid := slices.DeleteFunc(f.Messages, func(m *goi18n.Message) bool { return checkMessage(m) != nil })
		if b.AddMessages(f.Tag, valid...) == nil {
			c.tags = append(c.tags, f.Tag)
		}
	}
	c.matcher = language.NewMatcher(c.tags)
	return c, nil
}

// Has reports whether the en catalog holds an ID.
func (c *Catalogs) Has(id string) bool { return c.en[id] }

// Match picks the catalog for a language setting, such as fr-CA, fr, or the
// POSIX form fr_CA.UTF-8. A setting with no close catalog, including C and
// POSIX, gets English. The pseudo-locale needs an exact request, so that
// en-GB never selects it.
func (c *Catalogs) Match(setting string) language.Tag {
	tag, err := parseSetting(setting)
	if err != nil {
		return language.English
	}
	if tag == Pseudo {
		return Pseudo
	}
	_, index, confidence := c.matcher.Match(tag)
	if confidence == language.No {
		return language.English
	}
	return c.tags[index]
}

// Printer returns a printer of these catalogs for a language setting.
func (c *Catalogs) Printer(setting string) *Printer { return newPrinter(c.bundle, c.Match(setting)) }

// Requested returns the language that a setting requests, as a tag, such as
// fr-CA for fr_CA.UTF-8, whether or not a catalog speaks it. A setting that
// names no known language, including C and POSIX, requests en.
func Requested(setting string) string {
	tag, err := parseSetting(setting)
	if err != nil || tag == language.Und {
		return language.English.String()
	}
	return tag.String()
}

// parseSetting reads a language setting as a tag, without the encoding and
// the modifier of the POSIX form.
func parseSetting(setting string) (language.Tag, error) {
	setting, _, _ = strings.Cut(setting, ".")
	setting, _, _ = strings.Cut(setting, "@")
	return language.Parse(strings.ReplaceAll(setting, "_", "-"))
}

// Printer returns catalog text and formats values in one language.
type Printer struct {
	localizer *goi18n.Localizer
	numbers   *message.Printer
}

func newPrinter(b *goi18n.Bundle, tag language.Tag) *Printer {
	return &Printer{
		localizer: goi18n.NewLocalizer(b, tag.String()),
		numbers:   message.NewPrinter(tag),
	}
}

// T returns the catalog message for id, with data filling its template. A
// message that the active catalog lacks falls back to en, and a message that
// en lacks too returns the id itself, so a scenario shows it.
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
		maps.Copy(fill, data[0])
	}
	fill["Count"] = p.Number(float64(count), 0)
	return p.localize(&goi18n.LocalizeConfig{MessageID: id, PluralCount: count, TemplateData: fill})
}

func (p *Printer) localize(cfg *goi18n.LocalizeConfig) string {
	// Localize reports a message missing from the active catalog even when it
	// returns the en text, so only an empty text means that both lack it.
	text, _ := p.localizer.Localize(cfg)
	if text == "" {
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

// Size formats a byte count with at most one decimal in the largest unit
// that keeps the number at 1 or more, such as 1.5 MB.
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

// checkMessage requires the other form, which the printer uses for every
// message without a count, and parses each form as the printer will, so
// that a broken entry fails when the catalog loads, not as an ID in help.
func checkMessage(m *goi18n.Message) error {
	if m.Other == "" {
		return fmt.Errorf("%s has no other form", m.ID)
	}
	for _, form := range []string{m.Zero, m.One, m.Two, m.Few, m.Many, m.Other} {
		if _, err := (&template.TextParser{}).Parse(form, m.LeftDelim, m.RightDelim); err != nil {
			return fmt.Errorf("%s: %w", m.ID, err)
		}
	}
	return nil
}

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
