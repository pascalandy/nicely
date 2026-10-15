package extension

import (
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Frontmatter holds the keys of the frontmatter of a SKILL.md that Nicely
// reads.
type Frontmatter struct {
	Name        string
	Description string
}

// ErrNoFrontmatter reports a SKILL.md without a frontmatter, or whose
// frontmatter is not a YAML mapping of strings.
var ErrNoFrontmatter = errors.New("no valid frontmatter")

// MissingKeyError reports a SKILL.md whose frontmatter lacks a key that a
// skill needs, or leaves it empty.
type MissingKeyError struct {
	Key string
}

func (e *MissingKeyError) Error() string { return "the frontmatter lacks " + e.Key }

// ParseSkill reads the frontmatter of a SKILL.md: the YAML between a first
// line --- and the next line ---, as the Agent Skills specification defines
// it. A file without one, whose YAML is invalid, or whose name or
// description is not a string is ErrNoFrontmatter, and one whose name or
// description is missing or empty is a MissingKeyError.
func ParseSkill(data []byte) (Frontmatter, error) {
	lines := strings.SplitAfter(string(data), "\n")
	fence := func(line string) bool { return strings.TrimRight(line, "\r\n") == "---" }
	if !fence(lines[0]) {
		return Frontmatter{}, ErrNoFrontmatter
	}
	end := 1
	for end < len(lines) && !fence(lines[end]) {
		end++
	}
	if end == len(lines) {
		return Frontmatter{}, ErrNoFrontmatter
	}
	var keys map[string]any
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "")), &keys); err != nil {
		return Frontmatter{}, ErrNoFrontmatter
	}
	var fm Frontmatter
	for _, key := range []struct {
		name string
		into *string
	}{{"name", &fm.Name}, {"description", &fm.Description}} {
		switch value := keys[key.name].(type) {
		case nil:
			return Frontmatter{}, &MissingKeyError{Key: key.name}
		case string:
			if value == "" {
				return Frontmatter{}, &MissingKeyError{Key: key.name}
			}
			*key.into = value
		default:
			return Frontmatter{}, ErrNoFrontmatter
		}
	}
	return fm, nil
}
