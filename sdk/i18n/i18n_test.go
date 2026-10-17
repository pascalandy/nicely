package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

func TestPseudoMarksTextAndKeepsTemplateActions(t *testing.T) {
	got := sdkCatalogs(t).Printer("en-XA").T("usage.unknown_command", map[string]any{"Name": "nope"})
	// The English text holds 28 characters, so a third more is 10 dots.
	want := `⟦Ûñķñöŵñ çöɱɱáñð "nope".··········⟧`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
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
		{"fr-CA", 2 * time.Second, "2 secondes"},
	}
	for _, c := range cases {
		if got := newPrinter(b, language.MustParse(c.lang)).Duration(c.d); got != c.want {
			t.Errorf("%s, %v: got %q, want %q", c.lang, c.d, got, c.want)
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

// TestEveryCatalogFollowsTheRules walks every catalog folder of the
// repository, core's, the SDK's, and each extension's, as SDK in
// contract.md asks, and checks that core and the SDK share no ID.
func TestEveryCatalogFollowsTheRules(t *testing.T) {
	root := filepath.Join("..", "..")
	owners := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules"):
			return filepath.SkipDir
		case !d.IsDir() || d.Name() != "locales":
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		domain := ""
		if parent := filepath.Dir(filepath.Dir(rel)); filepath.Base(parent) == "extensions" {
			domain = filepath.Base(filepath.Dir(rel))
		}
		if err := CheckCatalogs(os.DirFS(path), domain); err != nil {
			t.Errorf("%s: %v", rel, err)
		}
		en, err := readCatalog(os.DirFS(path), "en.toml")
		owners[filepath.ToSlash(rel)] = slices.Sorted(maps.Keys(en))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	core, sdk := owners["internal/cli/locales"], owners["sdk/i18n/locales"]
	if len(core) == 0 || len(sdk) == 0 {
		t.Fatalf("found the catalog folders %v, want core's and the SDK's among them", slices.Sorted(maps.Keys(owners)))
	}
	for _, id := range core {
		if slices.Contains(sdk, id) {
			t.Errorf("core's catalog and the SDK's both hold %s", id)
		}
	}
}

// TestCheckCatalogsHoldsEachTranslationToTheIDsOfEnglish rejects a
// translation that lacks an ID of en, and one that adds an ID, such as one
// of the SDK's, whose text it would replace.
func TestCheckCatalogsHoldsEachTranslationToTheIDsOfEnglish(t *testing.T) {
	entry := func(id string) string { return fmt.Sprintf("[%q]\ndescription = \"d\"\nother = \"x\"\n", id) }
	for name, translation := range map[string]string{
		"lacks an ID": "",
		"adds an ID":  entry("root.summary") + entry("usage.invalid"),
	} {
		fsys := fstest.MapFS{
			"en.toml":    {Data: []byte(entry("root.summary"))},
			"fr-CA.toml": {Data: []byte(translation)},
		}
		if err := CheckCatalogs(fsys, ""); err == nil {
			t.Errorf("%s: CheckCatalogs accepted it", name)
		}
	}
}
