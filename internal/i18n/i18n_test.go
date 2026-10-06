package i18n

import (
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

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
		if got := Match(setting).String(); got != want {
			t.Errorf("Match(%q) = %s, want %s", setting, got, want)
		}
	}
}

func TestPseudoMarksTextAndKeepsTemplateActions(t *testing.T) {
	got := New(Pseudo).T("usage.unknown_command", map[string]any{"Name": "nope"})
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
}

func TestFormats(t *testing.T) {
	en, fr := New(language.English), newPrinter(bundle, language.MustParse("fr-CA"))
	day := time.Date(2026, 10, 6, 14, 30, 0, 0, time.UTC)
	cases := []struct{ got, want string }{
		{en.Number(1234.5, 1), "1,234.5"},
		{en.Number(2, 1), "2"},
		{strings.ReplaceAll(fr.Number(1234.5, 1), " ", " "), "1 234,5"},
		{en.Size(512), "512 B"},
		{en.Size(1500), "1.5 kB"},
		{en.Size(1_500_000), "1.5 MB"},
		{en.Size(2_000_000_000_000_000), "2,000 TB"},
		{en.Date(day), "2026-10-06"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

type catalog map[string]map[string]string

func load(t *testing.T, path string) catalog {
	t.Helper()
	var c catalog
	if _, err := toml.DecodeFile(path, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEveryEntryHasADescription(t *testing.T) {
	for id, entry := range load(t, "locales/en.toml") {
		if strings.TrimSpace(entry["description"]) == "" {
			t.Errorf("%s has no description", id)
		}
	}
}

func TestEveryCatalogMatchesEnglish(t *testing.T) {
	en := slices.Sorted(maps.Keys(load(t, "locales/en.toml")))
	paths, _ := filepath.Glob("locales/*.toml")
	for _, path := range paths {
		if got := slices.Sorted(maps.Keys(load(t, path))); !slices.Equal(got, en) {
			t.Errorf("%s holds the entries %v, en holds %v", path, got, en)
		}
	}
}

// TestEveryEntryIsUsed finds each ID as a string literal in the Go code, so
// a dead entry never waits for a translation.
func TestEveryEntryIsUsed(t *testing.T) {
	var source strings.Builder
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			data, err := os.ReadFile(path)
			source.Write(data)
			return err
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for id := range load(t, "locales/en.toml") {
		if !regexp.MustCompile(`"` + regexp.QuoteMeta(id) + `"`).MatchString(source.String()) {
			t.Errorf("no Go code uses %s", id)
		}
	}
}
