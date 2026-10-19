package runner

import (
	"maps"
	"slices"
	"strings"
)

// Environ returns the environment of a program: its parent's, without the
// key of any service, which needs a grant, and with each variable of set
// given its value, or removed when that value is empty.
func Environ(parent []string, set map[string]string) []string {
	env := slices.DeleteFunc(slices.Clone(parent), func(entry string) bool {
		name, _, _ := strings.Cut(entry, "=")
		_, changed := set[name]
		return changed || strings.HasSuffix(name, "_API_KEY")
	})
	for _, name := range slices.Sorted(maps.Keys(set)) {
		if value := set[name]; value != "" {
			env = append(env, name+"="+value)
		}
	}
	return env
}
