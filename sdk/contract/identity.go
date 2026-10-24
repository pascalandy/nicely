package contract

import (
	"slices"
	"strings"
)

// IsDomain reports whether an identity part starts with a lowercase letter
// and contains only lowercase letters and digits.
func IsDomain(name string) bool { return domainName.MatchString(name) }

// IsSlug accepts a domain or a domain and name joined by one hyphen.
func IsSlug(slug string) bool {
	parts := strings.Split(slug, "-")
	return len(parts) <= 2 && !slices.ContainsFunc(parts, func(part string) bool { return !IsDomain(part) })
}

// Slug identifies the extension in executable and sidecar names.
func (m Manifest) Slug() string { return strings.Join(m.Root(), "-") }

// Root returns the complete command root that this extension owns.
func (m Manifest) Root() []string {
	if m.Name != "" {
		return []string{m.Domain, m.Name}
	}
	return []string{m.Domain}
}

// Owns reports whether path is the extension root or one of its descendants.
func (m Manifest) Owns(path []string) bool {
	root := m.Root()
	return len(path) >= len(root) && slices.Equal(path[:len(root)], root)
}

func manifestIdentity(domain string, name *string, slug string) (Manifest, error) {
	m := Manifest{Domain: domain}
	if !IsDomain(domain) {
		return Manifest{}, broken("extension.domain", map[string]any{"Value": domain, "Want": slug})
	}
	if name != nil {
		if !IsDomain(*name) {
			return Manifest{}, broken("extension.identity_name", map[string]any{"Value": *name})
		}
		m.Name = *name
	}
	if slug != "" && m.Slug() != slug {
		if name == nil {
			return Manifest{}, broken("extension.domain", map[string]any{"Value": domain, "Want": slug})
		}
		return Manifest{}, broken("extension.identity_slug", map[string]any{"Value": m.Slug(), "Want": slug})
	}
	return m, nil
}

// Runs reports whether a command starts the extension.
// A whole-domain root prints its help, as Usage in contract.md says.
// A command with children and no arguments only groups them.
func (m Manifest) Runs(c Command) bool {
	if len(c.Path) == 1 {
		return false
	}
	return len(c.Args) > 0 || !slices.ContainsFunc(m.Commands, func(other Command) bool {
		return len(other.Path) == len(c.Path)+1 && slices.Equal(other.Path[:len(c.Path)], c.Path)
	})
}

// codePrefix starts every code that the extension owns, such as HELLO_.
func (m Manifest) codePrefix() string {
	return strings.ToUpper(strings.ReplaceAll(m.Slug(), "-", "_")) + "_"
}
