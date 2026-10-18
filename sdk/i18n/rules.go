package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// CheckCatalogs applies the catalog rules of SDK in contract.md to the
// catalogs at the top of fsys: every entry has a description, an entry
// whose other form reads a number field has a one form, and every language
// holds exactly the IDs of en. A non-empty domain, an extension's, also
// requires every ID to start with the domain and a dot.
func CheckCatalogs(fsys fs.FS, domain string) error {
	en, err := readCatalog(fsys, "en.toml")
	if err != nil {
		return err
	}
	files, err := fs.Glob(fsys, "*.toml")
	if err != nil {
		return err
	}
	for _, file := range files {
		catalog, err := readCatalog(fsys, file)
		if err != nil {
			return err
		}
		for _, id := range slices.Sorted(maps.Keys(catalog)) {
			if description, _ := catalog[id]["description"].(string); strings.TrimSpace(description) == "" {
				return fmt.Errorf("%s: %s has no description", file, id)
			}
			if domain != "" && !strings.HasPrefix(id, domain+".") {
				return fmt.Errorf("%s: %s does not start with %q", file, id, domain+".")
			}
			if _, found := en[id]; !found {
				return fmt.Errorf("%s holds %s, which en.toml lacks", file, id)
			}
			if err := checkOneForm(catalog[id]); err != nil {
				return fmt.Errorf("%s: %s %w", file, id, err)
			}
		}
		for _, id := range slices.Sorted(maps.Keys(en)) {
			if _, found := catalog[id]; !found {
				return fmt.Errorf("%s lacks %s, which en.toml holds", file, id)
			}
		}
	}
	return nil
}

// checkOneForm requires a one form of an entry whose other form shows a
// number from the field Count, which Printer.N fills, or N, which
// Printer.Size and Printer.Duration fill. When English writes both forms the
// same, as for a unit symbol, neither a scenario nor the pseudo-locale sees
// a missing one, and a language whose forms differ would show the other form
// for 1.
func checkOneForm(entry map[string]any) error {
	m, err := goi18n.NewMessage(entry)
	if err != nil {
		return fmt.Errorf("is not a message: %w", err)
	}
	// The printer parses each form as text/template does here, under the
	// name "", so a form that defines its own templates parses the same.
	other, err := template.New("").Delims(m.LeftDelim, m.RightDelim).Parse(m.Other)
	if err != nil {
		return fmt.Errorf("has an other form that does not parse: %w", err)
	}
	if strings.TrimSpace(m.One) != "" {
		return nil
	}
	if field := numberField(other); field != "" {
		return fmt.Errorf("reads .%s in its other form, but has no one form", field)
	}
	return nil
}

// numberField returns Count or N when the text of t changes with the value
// of that field, or "" when it changes with neither. Rendering sees every
// way that a template reads a field, such as {{ .Count }}, {{$.N}}, a
// variable, or a template that the form defines. A form that fails on these
// values keeps the text that it wrote before the failure.
func numberField(t *template.Template) string {
	render := func(count, n string) string {
		var b strings.Builder
		_ = t.Execute(&b, map[string]any{"Count": count, "N": n})
		return b.String()
	}
	switch shown := render("1", "1"); {
	case render("2", "1") != shown:
		return "Count"
	case render("1", "2") != shown:
		return "N"
	}
	return ""
}

func readCatalog(fsys fs.FS, file string) (map[string]map[string]any, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, err
	}
	var catalog map[string]map[string]any
	if err := toml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return catalog, nil
}
