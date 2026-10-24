package cli

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// sharedGroups declares one host-owned group for each valid shared domain.
func sharedGroups(p *i18n.Printer, extensions []*extension) []entry {
	seen := map[string]bool{}
	var groups []entry
	for _, e := range extensions {
		if e.broken != nil || e.manifest.Name == "" || seen[e.domain] {
			continue
		}
		seen[e.domain] = true
		d := contract.Command{
			Path: []string{e.domain}, Summary: "extension.group_summary", Effects: &contract.Effects{},
			Examples: []string{"ncly " + e.domain + " --help", "ncly describe " + e.domain + " --json"},
		}
		groups = append(groups, entry{
			Path: e.domain, Summary: p.T(d.Summary, map[string]any{"Domain": e.domain}), Source: sourceCore,
			declaration: d, text: p, group: true,
		})
	}
	return groups
}

// forPath loads only the selected root's declarations for a direct call.
// A domain group or an unknown child needs its siblings for help.
func (h host) forPath(words []string) []*extension {
	if e := h.find(words[:1]); e != nil {
		return []*extension{e}
	}
	if len(words) > 1 {
		if e := h.find(words[:2]); e != nil {
			return []*extension{e}
		}
	}
	selected := h.selected()
	var siblings []*extension
	for _, e := range selected {
		if e.domain != words[0] {
			continue
		}
		siblings = append(siblings, e)
	}
	if len(siblings) == 0 {
		return loadSelected(selected)
	}
	return loadSelected(siblings)
}

// find selects an exact executable name without listing its search folders.
func (h host) find(root []string) *extension {
	slug := strings.Join(root, "-")
	for _, e := range h.bundled {
		if e.slug == slug {
			return e
		}
	}
	if !contract.IsSlug(slug) || h.reserved(root[0]) {
		return nil
	}
	for _, dir := range h.dirs {
		path := filepath.Join(dir, executablePrefix+slug)
		if isExecutable(path) {
			return loadExtension(slug, path)
		}
	}
	return nil
}

func targetPath(words []string) []string {
	if len(words) > 0 && words[0] == "help" {
		return words[1:]
	}
	return words
}

// brokenPath isolates a broken sibling to its own root. A group fails only
// when none of its selected roots has valid declarations.
func brokenPath(extensions []*extension, words []string) *extension {
	if len(words) == 0 {
		return nil
	}
	var broken *extension
	valid := false
	for _, e := range extensions {
		if within(words, e.root()) {
			if e.broken != nil {
				return e
			}
			return nil
		}
		if slices.Equal(words, []string{e.domain}) {
			if e.broken != nil {
				broken = e
			} else {
				valid = true
			}
		}
	}
	if valid {
		return nil
	}
	return broken
}
