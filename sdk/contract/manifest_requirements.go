package contract

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

// checkRequires checks the programs and the services that one command
// names: each program is a key of programs, each service a service name,
// and each name appears once.
func (m Manifest) checkRequires(path string, c Command) error {
	for i, name := range c.Requires.Bins {
		if !slices.Contains(m.Programs, name) || slices.Index(c.Requires.Bins, name) != i {
			return broken("extension.bin", map[string]any{"Path": path, "Name": name})
		}
	}
	for i, name := range c.Requires.Keys {
		if !IsService(name) || slices.Index(c.Requires.Keys, name) != i {
			return broken("extension.key", map[string]any{"Path": path, "Name": name})
		}
	}
	return nil
}

// checkPrograms checks that each key of programs is a file name, without
// "/", and that a command names it, then its entry: a version of numbers
// joined by dots, a probe that is not empty, and install commands that are
// not empty.
func (m Manifest) checkPrograms() error {
	for _, name := range m.Programs {
		named := slices.ContainsFunc(m.Commands, func(c Command) bool { return slices.Contains(c.Requires.Bins, name) })
		if name == "" || name == "." || name == ".." || strings.Contains(name, "/") || !named {
			return broken("extension.program", map[string]any{"Name": name})
		}
		r := m.Requirements[name]
		switch {
		case r.Version != "" && !versionNumber.MatchString(r.Version):
			return broken("extension.program_version", map[string]any{"Name": name, "Value": r.Version})
		case r.Probe != nil && len(r.Probe) == 0:
			return broken("extension.program_probe", map[string]any{"Name": name})
		}
		for _, manager := range slices.Sorted(maps.Keys(r.Install)) {
			if len(r.Install[manager]) == 0 {
				return broken("extension.program_install", map[string]any{"Name": name, "Manager": manager})
			}
		}
	}
	return nil
}

// checkKeys checks that each service of the keys table is one that the
// requires.keys of a command names.
func (m Manifest) checkKeys() error {
	for _, name := range slices.Sorted(maps.Keys(m.Keys)) {
		if !slices.ContainsFunc(m.Commands, func(c Command) bool { return slices.Contains(c.Requires.Keys, name) }) {
			return broken("extension.key_table", map[string]any{"Name": name})
		}
	}
	return nil
}

// versionNumber is a version as a manifest declares it: one or more numbers
// joined by dots.
var versionNumber = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

// checkConfig checks each command's declared config keys.
// Each key appears once and has at least two words joined by dots, starting with the slug.
func (m Manifest) checkConfig(path string, c Command) error {
	for i, key := range c.Config {
		words := strings.Split(key, ".")
		bad := len(words) < 2 || words[0] != m.Slug() || slices.Index(c.Config, key) != i
		for _, word := range words {
			bad = bad || !configWord.MatchString(word)
		}
		if bad {
			return broken("extension.config_key", map[string]any{"Path": path, "Name": key, "Domain": m.Slug()})
		}
	}
	return nil
}
