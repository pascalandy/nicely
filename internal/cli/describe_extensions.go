package cli

import (
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/i18n"
)

func extensionEntries(p *i18n.Printer, lang string, extensions []*extension) []entry {
	all := sharedGroups(p, extensions)
	for _, e := range extensions {
		if e.broken != nil {
			continue
		}
		text := e.catalogs.Printer(lang)
		for _, d := range e.manifest.Commands {
			en := entry{
				Path: strings.Join(d.Path, " "), Summary: text.T(d.Summary), Source: sourceExtension,
				declaration: d, text: text, group: !e.manifest.Runs(d), requirements: e.manifest.Requirements, keys: e.manifest.Keys,
			}
			if e.bundled {
				en.Source = sourceBundled
			}
			if slices.Equal(d.Path, e.manifest.Root()) {
				en.Skill = e.skill()
			}
			all = append(all, en)
		}
	}
	return all
}

// skill returns the slug-named sidecar beside the discovered executable.
// A bundled extension or an absent skill returns nil.
func (e *extension) skill() *string {
	if e.bundled {
		return nil
	}
	folder := e.path + ".skill"
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return nil
	}
	return &folder
}
