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

// TestCheckCatalogsRequiresTheOneFormOfANumber rejects, in each language,
// an entry whose other form reads .Count or .N without a one form, which
// English hides when it writes both forms the same, and accepts the entry
// once it has that form.
func TestCheckCatalogsRequiresTheOneFormOfANumber(t *testing.T) {
	entry := func(forms string) string { return "[\"item.count\"]\ndescription = \"d\"\n" + forms }
	cases := []struct{ name, without, with string }{
		{"Count", `other = "{{.Count}} items"`, `one = "{{.Count}} item"` + "\n" + `other = "{{.Count}} items"`},
		{"N", `other = "{{.N}} B"`, `one = "{{.N}} B"` + "\n" + `other = "{{.N}} B"`},
		{"Count with spaces", `other = "{{ .Count }} items"`, `one = "{{ .Count }} item"` + "\n" + `other = "{{ .Count }} items"`},
		{"empty one", `one = ""` + "\n" + `other = "{{.N}} B"`, `one = "{{.N}} B"` + "\n" + `other = "{{.N}} B"`},
	}
	for _, c := range cases {
		valid := entry(c.with + "\n")
		for _, file := range []string{"en.toml", "fr-CA.toml"} {
			fsys := fstest.MapFS{"en.toml": {Data: []byte(valid)}, "fr-CA.toml": {Data: []byte(valid)}}
			fsys[file] = &fstest.MapFile{Data: []byte(entry(c.without + "\n"))}
			err := CheckCatalogs(fsys, "")
			if err == nil || !strings.HasPrefix(err.Error(), file+": item.count ") || !strings.Contains(err.Error(), "one form") {
				t.Errorf("%s in %s: got %v, want an error that names %s, item.count, and the one form", c.name, file, err, file)
			}
			fsys[file] = &fstest.MapFile{Data: []byte(valid)}
			if err := CheckCatalogs(fsys, ""); err != nil {
				t.Errorf("%s in %s with its one form: %v", c.name, file, err)
			}
		}
	}
	for _, other := range []string{"No items.", "{{.Name}} is ready.", "{{.Number}} of {{.Counter}}"} {
		fsys := fstest.MapFS{"en.toml": {Data: []byte(entry(fmt.Sprintf("other = %q\n", other)))}}
		if err := CheckCatalogs(fsys, ""); err != nil {
			t.Errorf("%q, which reads neither .Count nor .N, needs only other: %v", other, err)
		}
	}
}
