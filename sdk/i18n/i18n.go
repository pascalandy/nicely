// Package i18n looks up the text that ncly and its extensions show a human
// in their message catalogs, and formats numbers, sizes, and dates for the
// active language. Every catalog entry carries a description that tells a
// translator where the text appears.
//
// Send every string that a human reads through a catalog. Name each ID
// <domain>.<thing>, and never reuse an ID for a new meaning. Give every
// message that holds a number its plural forms. Send such a message through
// N, Size, or Duration: T always shows its other form, and the test binary
// of cmd/ncly refuses it. Format numbers, sizes, durations, and dates through
// this package, and let translated text set its own width. Pass a value from
// the command line or a variable through Clip before a message that refuses
// it quotes it.
package i18n

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n/template"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed locales/*.toml
var locales embed.FS

// Locales returns the SDK's own catalogs: the text that the SDK renders,
// such as help pages, usage errors, and outcome summaries.
func Locales() fs.FS {
	sub, err := fs.Sub(locales, "locales")
	if err != nil {
		panic(err)
	}
	return sub
}

// PluralGuard, when true, makes Printer.T panic on a message with a one form
// in any loaded language, and name its ID. T passes no count, so it shows the
// other form. Only the test binary of cmd/ncly sets it, so a scenario fails
// where such a call runs, even with the ID in a variable. See #234.
var PluralGuard bool

// Catalogs holds the catalogs of one or more owners, such as the SDK and
// core, or the SDK and an extension: for each owner, one file per language,
// named after its tag, such as en.toml or fr-CA.toml. Each owner's en
// catalog is required, and the pseudo-locale marks it.
type Catalogs struct {
	bundle  *goi18n.Bundle
	tags    []language.Tag // the languages with a catalog, en first
	matcher language.Matcher
	en      map[string]bool
	plural  map[string]bool // the IDs with a one form in any loaded language
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

// Load reads the catalogs at the top of each folder. Only each en catalog
// must be valid: a translation that cannot be read is skipped, so its
// entries fall back to en. The owners' IDs never overlap, as the catalog
// rules require, so the order of the folders does not matter.
func Load(folders ...fs.FS) (*Catalogs, error) {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c := &Catalogs{bundle: b, tags: []language.Tag{language.English}, en: map[string]bool{}, plural: map[string]bool{}}
	for _, fsys := range folders {
		en, err := b.LoadMessageFileFS(fsys, "en.toml")
		if err != nil {
			e := &CatalogError{File: "en.toml", Err: err}
			var parse toml.ParseError
			if errors.As(err, &parse) {
				e.Line = parse.Position.Line
			}
			return nil, e
		}
		for _, m := range en.Messages {
			if err := checkMessage(m); err != nil {
				return nil, &CatalogError{File: "en.toml", Err: err}
			}
			c.en[m.ID] = true
			c.markPlural(m)
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
			if b.AddMessages(f.Tag, valid...) != nil {
				continue
			}
			for _, m := range valid {
				c.markPlural(m)
			}
			if !slices.Contains(c.tags, f.Tag) {
				c.tags = append(c.tags, f.Tag)
			}
		}
	}
	c.matcher = language.NewMatcher(c.tags)
	return c, nil
}

// markPlural records the ID of a message with a one form, for PluralGuard.
func (c *Catalogs) markPlural(m *goi18n.Message) {
	if m.One != "" {
		c.plural[m.ID] = true
	}
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
func (c *Catalogs) Printer(setting string) *Printer {
	p := newPrinter(c.bundle, c.Match(setting))
	p.plural = c.plural
	return p
}

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

// IsLanguage reports whether ncly accepts a language setting, with or
// without a catalog. It accepts a known tag such as de, its POSIX form such
// as fr_CA.UTF-8, C, and POSIX. It refuses an unknown language, such as xx-YY.
func IsLanguage(setting string) bool {
	if posix := withoutEncoding(setting); posix == "C" || posix == "POSIX" {
		return true
	}
	_, err := parseSetting(setting)
	return err == nil
}

// parseSetting reads a language setting as a tag, without the encoding and
// the modifier of the POSIX form.
func parseSetting(setting string) (language.Tag, error) {
	return language.Parse(strings.ReplaceAll(withoutEncoding(setting), "_", "-"))
}

func withoutEncoding(setting string) string {
	setting, _, _ = strings.Cut(setting, ".")
	setting, _, _ = strings.Cut(setting, "@")
	return setting
}

// Printer returns catalog text and formats values in one language.
type Printer struct {
	localizer *goi18n.Localizer
	numbers   *message.Printer
	// plain writes a number in the digits that numbers shows, in English
	// and without grouping, for the plural rules.
	plain *message.Printer
	// plural holds the IDs that PluralGuard refuses.
	plural map[string]bool
}

func newPrinter(b *goi18n.Bundle, tag language.Tag) *Printer {
	return &Printer{
		localizer: goi18n.NewLocalizer(b, tag.String()),
		numbers:   message.NewPrinter(tag),
		plain:     message.NewPrinter(language.English),
	}
}

// T returns the catalog message for id, with data filling its template. A
// message that the active catalog lacks falls back to en, and a message that
// en lacks too returns the id itself, so a scenario shows it.
func (p *Printer) T(id string, data ...map[string]any) string {
	if PluralGuard && p.plural[id] {
		panic(fmt.Sprintf("i18n: %s has plural forms, which Printer.T cannot choose: use Printer.N, Size, or Duration", id))
	}
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
