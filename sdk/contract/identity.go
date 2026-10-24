package contract

import (
	"slices"
	"strings"
)

// IsDomain reports whether name can be a domain: lowercase letters and
// digits that start with a letter.
func IsDomain(name string) bool { return domainName.MatchString(name) }

// Runs reports whether a command starts the extension. The domain alone
// prints its help, as Usage in contract.md says, and so does a command with
// children and no arguments, which only groups them.
func (m Manifest) Runs(c Command) bool {
	if len(c.Path) == 1 {
		return false
	}
	if len(c.Args) > 0 {
		return true
	}
	return !slices.ContainsFunc(m.Commands, func(other Command) bool {
		return len(other.Path) == len(c.Path)+1 && slices.Equal(other.Path[:len(c.Path)], c.Path)
	})
}

// codePrefix starts every code that the extension owns, such as HELLO_.
func (m Manifest) codePrefix() string { return strings.ToUpper(m.Domain) + "_" }
