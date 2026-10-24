package cli

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// extensionChecks reports selected and hidden executables in slug order.
// Each selected extension precedes its hidden copies, in discovery order.
// It starts no extension and reads no hidden declarations.
func extensionChecks(p *i18n.Printer, extensions []*extension) []check {
	winners := map[string]*extension{}
	for _, e := range extensions {
		winners[e.slug] = e
	}
	found := map[string][]string{}
	seen := map[string]bool{}
	for _, dir := range searchPath() {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			slug, ok := strings.CutPrefix(entry.Name(), executablePrefix)
			path := filepath.Join(dir, entry.Name())
			if ok && contract.IsSlug(slug) && !seen[path] && isExecutable(path) {
				seen[path] = true
				found[slug] = append(found[slug], path)
			}
		}
	}
	slugs := slices.Collect(maps.Keys(found))
	for slug := range winners {
		if !slices.Contains(slugs, slug) {
			slugs = append(slugs, slug)
		}
	}
	slices.Sort(slugs)
	var checks []check
	for _, slug := range slugs {
		entry := func(status, source string, path *string, message string) check {
			return check{
				ID: "extensions." + slug, Component: "extensions", Status: status, Message: message,
				extensionKeys: &extensionKeys{Domain: slug, Source: source, Path: path},
			}
		}
		winner := winners[slug]
		if winner != nil {
			source, path := sourceExtension, &winner.path
			if winner.bundled {
				source, path = sourceBundled, nil
			}
			c := entry(statusPass, source, path, p.T("doctor.extension_runs", map[string]any{"Domain": slug, "Path": winner.path}))
			if winner.bundled {
				c.Message = p.T("doctor.extension_bundled", map[string]any{"Domain": slug})
			}
			if winner.broken != nil {
				problem := winner.invalid(p)
				c.Status, c.Code, c.Message, c.Hint = statusFail, problem.Code, problem.Message, problem.Hint
			}
			checks = append(checks, c)
		}
		domain, _, _ := strings.Cut(slug, "-")
		if winner == nil {
			winner = winners[domain]
		}
		for _, path := range found[slug] {
			data := map[string]any{"Domain": domain, "Path": path}
			var c check
			switch {
			case winner != nil && winner.bundled:
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden_by_bundled", data))
				c.HiddenBy = &hiddenBy{Source: sourceBundled}
			case winner == nil:
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden_by_core", data))
				c.HiddenBy = &hiddenBy{Source: sourceCore}
			case path == winner.path:
				continue
			default:
				data["Winner"] = winner.path
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden", data))
				c.HiddenBy = &hiddenBy{Source: sourceExtension, Path: &winner.path}
			}
			checks = append(checks, c)
		}
	}
	return checks
}
