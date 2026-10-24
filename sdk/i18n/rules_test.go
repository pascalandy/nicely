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

	"github.com/pascalandy/nicely/sdk/contract"
)

// TestEveryCatalogFollowsTheRules walks every catalog folder of the
// repository, core's, the SDK's, and each extension's, as SDK in
// contract.md asks, and checks that core and the SDK share no ID. It skips
// template/, whose catalog is a Copier source: extension.Check checks the
// catalog that it renders, in scripts/template-test.sh.
func TestEveryCatalogFollowsTheRules(t *testing.T) {
	root := filepath.Join("..", "..")
	owners := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || path == filepath.Join(root, "template")):
			return filepath.SkipDir
		case !d.IsDir() || d.Name() != "locales":
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		domain := ""
		if parent := filepath.Dir(filepath.Dir(rel)); filepath.Base(parent) == "extensions" {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(path), "manifest.toml"))
			if err != nil {
				return err
			}
			manifest, err := contract.ParseManifest(data, "")
			if err != nil {
				return err
			}
			domain = manifest.Slug()
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
		{"$.Count", `other = "{{$.Count}} items"`, `one = "{{$.Count}} item"` + "\n" + `other = "{{$.Count}} items"`},
		{"(.).Count", `other = "{{(.).Count}} items"`, `one = "{{(.).Count}} item"` + "\n" + `other = "{{(.).Count}} items"`},
		{"Count passed to a template", `other = '{{define "n"}}{{.}} items{{end}}{{template "n" .Count}}'`, `one = "{{.Count}} item"` + "\n" + `other = '{{define "n"}}{{.}} items{{end}}{{template "n" .Count}}'`},
		{"N in a template", `other = '{{define "n"}}{{.N}} B{{end}}{{template "n" .}}'`, `one = "{{.N}} B"` + "\n" + `other = '{{define "n"}}{{.N}} B{{end}}{{template "n" .}}'`},
		{"($).Count", `other = "{{($).Count}} items"`, `one = "{{.Count}} item"` + "\n" + `other = "{{($).Count}} items"`},
		{"((.)).Count", `other = "{{((.)).Count}} items"`, `one = "{{.Count}} item"` + "\n" + `other = "{{((.)).Count}} items"`},
		{"($).N", `other = "{{($).N}} B"`, `one = "{{.N}} B"` + "\n" + `other = "{{($).N}} B"`},
		{"a variable", `other = "{{$root := .}}{{$root.Count}} items"`, `one = "{{.Count}} item"` + "\n" + `other = "{{$root := .}}{{$root.Count}} items"`},
		{"Count behind another field", `other = "{{if .Show}}{{.Count}} items{{end}}"`, `one = "{{.Count}} item"` + "\n" + `other = "{{if .Show}}{{.Count}} items{{end}}"`},
		{"N after a failing call", `other = "{{len .Items}} of {{.N}} B"`, `one = "{{.N}} B"` + "\n" + `other = "{{len .Items}} of {{.N}} B"`},
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
	// The printer parses a form under the name "", so a form may define a
	// template named other.
	for _, other := range []string{"No items.", "{{.Name}} is ready.", "{{.Number}} of {{.Counter}}", `{{define "other"}}Ready.{{end}}{{template "other" .}}`, "{{range .}}{{.}} is ready.{{end}}"} {
		fsys := fstest.MapFS{"en.toml": {Data: []byte(entry(fmt.Sprintf("other = %q\n", other)))}}
		if err := CheckCatalogs(fsys, ""); err != nil {
			t.Errorf("%q, which reads neither .Count nor .N, needs only other: %v", other, err)
		}
	}
}
