package i18n

import (
	"strings"
	"testing"
	"time"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// TestDurationPluralForms proves that a duration takes the plural form of
// the number that it shows, decimals included.
func TestDurationPluralForms(t *testing.T) {
	b := goi18n.NewBundle(language.English)
	seconds := &goi18n.Message{ID: "duration.seconds", One: "{{.N}} second", Other: "{{.N}} seconds"}
	secondes := &goi18n.Message{ID: "duration.seconds", One: "{{.N}} seconde", Other: "{{.N}} secondes"}
	if err := b.AddMessages(language.English, seconds); err != nil {
		t.Fatal(err)
	}
	if err := b.AddMessages(language.MustParse("fr-CA"), secondes); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		lang string
		d    time.Duration
		want string
	}{
		{"en", time.Second, "1 second"},
		{"en", 1500 * time.Millisecond, "1.5 seconds"},
		{"en", 2 * time.Second, "2 seconds"},
		{"fr-CA", 1500 * time.Millisecond, "1,5 seconde"},
		// 1.95 shows as 1,9, which takes one, though rounding half up gives 2.
		{"fr-CA", 1950 * time.Millisecond, "1,9 seconde"},
		{"fr-CA", 2 * time.Second, "2 secondes"},
	}
	for _, c := range cases {
		if got := newPrinter(b, language.MustParse(c.lang)).Duration(c.d); got != c.want {
			t.Errorf("%s, %v: got %q, want %q", c.lang, c.d, got, c.want)
		}
	}
}

// TestSizePluralForms proves that a size takes the plural form of the
// number that it shows, after rounding, not of the raw value: 1.04 kB shows
// as 1 and 1.96 kB as 2.
func TestSizePluralForms(t *testing.T) {
	b := goi18n.NewBundle(language.English)
	en := []*goi18n.Message{
		{ID: "size.bytes", One: "{{.N}} byte", Other: "{{.N}} bytes"},
		{ID: "size.kilobytes", One: "{{.N}} kilobyte", Other: "{{.N}} kilobytes"},
	}
	fr := []*goi18n.Message{
		{ID: "size.bytes", One: "{{.N}} octet", Other: "{{.N}} octets"},
		{ID: "size.kilobytes", One: "{{.N}} kilooctet", Other: "{{.N}} kilooctets"},
	}
	if err := b.AddMessages(language.English, en...); err != nil {
		t.Fatal(err)
	}
	if err := b.AddMessages(language.MustParse("fr-CA"), fr...); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		lang  string
		bytes int64
		want  string
	}{
		{"en", 0, "0 bytes"},
		{"en", 1, "1 byte"},
		{"en", 2, "2 bytes"},
		{"en", 1_000, "1 kilobyte"},
		{"en", 1_040, "1 kilobyte"},
		{"en", 1_500, "1.5 kilobytes"},
		{"fr-CA", 0, "0 octet"},
		{"fr-CA", 1, "1 octet"},
		{"fr-CA", 2, "2 octets"},
		{"fr-CA", 1_500, "1,5 kilooctet"},
		{"fr-CA", 1_950, "1,9 kilooctet"},
		{"fr-CA", 1_960, "2 kilooctets"},
	}
	for _, c := range cases {
		if got := newPrinter(b, language.MustParse(c.lang)).Size(c.bytes); got != c.want {
			t.Errorf("%s, %d: got %q, want %q", c.lang, c.bytes, got, c.want)
		}
	}
}

func TestFormats(t *testing.T) {
	sdk := sdkCatalogs(t)
	en, fr := sdk.Printer("en"), newPrinter(sdk.bundle, language.MustParse("fr-CA"))
	day := time.Date(2026, 10, 6, 14, 30, 0, 0, time.UTC)
	cases := []struct{ got, want string }{
		{en.Number(1234.5, 1), "1,234.5"},
		{en.Number(2, 1), "2"},
		{strings.ReplaceAll(fr.Number(1234.5, 1), " ", " "), "1 234,5"},
		{en.Size(512), "512 B"},
		{en.Size(1500), "1.5 kB"},
		{en.Size(1_500_000), "1.5 MB"},
		{en.Size(2_000_000_000_000_000), "2,000 TB"},
		{en.Duration(250 * time.Millisecond), "250 ms"},
		{en.Duration(1500 * time.Millisecond), "1.5 s"},
		{en.Duration(90 * time.Second), "1.5 min"},
		{en.Duration(2 * time.Hour), "2 h"},
		{en.Duration(200 * time.Hour), "200 h"},
		{strings.ReplaceAll(fr.Duration(1500*time.Millisecond), " ", " "), "1,5 s"},
		{en.Date(day), "2026-10-06"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}
