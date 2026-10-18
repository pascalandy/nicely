package i18n

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func sdkCatalogs(t *testing.T) *Catalogs {
	t.Helper()
	c, err := Load(Locales())
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMatch(t *testing.T) {
	cases := map[string]string{
		"en":          "en",
		"en-US":       "en",
		"en_US.UTF-8": "en",
		"en-GB":       "en",
		"en-XA":       "en-XA",
		"en_XA.UTF-8": "en-XA",
		"fr_CA.UTF-8": "en",
		"xx":          "en",
		"C":           "en",
		"C.UTF-8":     "en",
		"POSIX":       "en",
		"":            "en",
	}
	for setting, want := range cases {
		if got := sdkCatalogs(t).Match(setting).String(); got != want {
			t.Errorf("Match(%q) = %s, want %s", setting, got, want)
		}
	}
}

func TestPluralForms(t *testing.T) {
	b := goi18n.NewBundle(language.English)
	files := &goi18n.Message{ID: "files", One: "{{.Count}} file", Other: "{{.Count}} files"}
	fichiers := &goi18n.Message{ID: "files", One: "{{.Count}} fichier", Other: "{{.Count}} fichiers"}
	if err := b.AddMessages(language.English, files); err != nil {
		t.Fatal(err)
	}
	if err := b.AddMessages(language.MustParse("fr-CA"), fichiers); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		lang  string
		count int
		want  string
	}{
		{"en", 0, "0 files"},
		{"en", 1, "1 file"},
		{"en", 2, "2 files"},
		{"en", 1500, "1,500 files"},
		{"fr-CA", 0, "0 fichier"},
		{"fr-CA", 1, "1 fichier"},
		{"fr-CA", 2, "2 fichiers"},
		{"fr-CA", 1500, "1 500 fichiers"},
	}
	for _, c := range cases {
		got := newPrinter(b, language.MustParse(c.lang)).N("files", c.count)
		if got := strings.ReplaceAll(got, " ", " "); got != c.want {
			t.Errorf("%s, %d: got %q, want %q", c.lang, c.count, got, c.want)
		}
	}
	if got := newPrinter(b, language.English).N("files", 1, map[string]any{"Count": "2"}); got != "1 file" {
		t.Errorf("a Count in data changed the count: got %q, want %q", got, "1 file")
	}
}

// TestPluralGuard proves the guard that the test binary of cmd/ncly turns
// on. T refuses a message with a one form in any loaded language and names
// its ID, and N still renders it. Off, as in a release, T renders the other
// form. go-i18n rejects tlh, which has no plural rule, so its forms count for
// nothing.
func TestPluralGuard(t *testing.T) {
	c, err := Load(fstest.MapFS{
		"en.toml": {Data: []byte(`
[files]
description = "d"
one = "{{.Count}} file"
other = "{{.Count}} files"

[ready]
description = "d"
other = "Download ready"

[hello]
description = "d"
other = "Hello"
`)},
		"fr-CA.toml": {Data: []byte(`
[ready]
description = "d"
one = "{{.Count}} fichier prêt"
other = "{{.Count}} fichiers prêts"
`)},
		"tlh.toml": {Data: []byte(`
[hello]
description = "d"
one = "nuqneH"
other = "nuqneH"
`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	en, fr := c.Printer("en"), c.Printer("fr-CA")
	if got := en.T("files", map[string]any{"Count": "1"}); got != "1 files" {
		t.Errorf("guard off: got %q, want %q", got, "1 files")
	}
	PluralGuard = true
	t.Cleanup(func() { PluralGuard = false })
	if got := en.T("hello"); got != "Hello" {
		t.Errorf("a message without a one form: got %q, want %q", got, "Hello")
	}
	if got := en.N("files", 1); got != "1 file" {
		t.Errorf("N: got %q, want %q", got, "1 file")
	}
	// ready has a one form in fr-CA only, so T refuses it in en too.
	for _, call := range []struct {
		lang string
		p    *Printer
		id   string
	}{{"en", en, "files"}, {"en", en, "ready"}, {"fr-CA", fr, "ready"}} {
		func() {
			defer func() {
				if r := recover(); !strings.Contains(fmt.Sprint(r), call.id+" has plural forms") {
					t.Errorf("%s, %s: got panic %v, want one that names %s", call.lang, call.id, r, call.id)
				}
			}()
			call.p.T(call.id, map[string]any{"Count": "1"})
		}()
	}
}
