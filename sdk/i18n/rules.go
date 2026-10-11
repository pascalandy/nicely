package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// CheckCatalogs applies the catalog rules of SDK in contract.md to the
// catalogs at the top of fsys: every entry has a description, and every
// language holds exactly the IDs of en. A non-empty domain, an extension's,
// also requires every ID to start with the domain and a dot.
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
		}
		for _, id := range slices.Sorted(maps.Keys(en)) {
			if _, found := catalog[id]; !found {
				return fmt.Errorf("%s lacks %s, which en.toml holds", file, id)
			}
		}
	}
	return nil
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
