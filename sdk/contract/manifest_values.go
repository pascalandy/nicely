package contract

import (
	"math"
	"slices"
	"strings"
)

// checkValues applies the rules of types, requiredness, variadic values,
// and defaults to the arguments and flags of one command.
func checkValues(path string, c Command) error {
	data := func(name string) map[string]any { return map[string]any{"Path": path, "Name": name} }
	optional := false
	for i, a := range c.Args {
		switch {
		case !slices.Contains(Types, a.ArgType()):
			return broken("extension.type", map[string]any{"Path": path, "Name": a.Name, "Value": string(a.Type)})
		case a.Required && optional:
			return broken("extension.required", data(a.Name))
		case a.Variadic && i != len(c.Args)-1:
			return broken("extension.variadic", data(a.Name))
		case !fitValues(a.Values, a.ArgType()):
			return broken("extension.values", data(a.Name))
		}
		optional = optional || !a.Required
	}
	for _, f := range c.Flags {
		switch {
		case !slices.Contains(Types, f.FlagType()), f.Value == "" && f.FlagType() != TypeBoolean:
			return broken("extension.type", map[string]any{"Path": path, "Name": f.Name, "Value": string(f.FlagType())})
		case f.Required && f.Value == "":
			return broken("extension.required", data(f.Name))
		case f.Variadic && f.Value == "":
			return broken("extension.variadic", data(f.Name))
		case f.Default != nil && (f.Required || f.Variadic || !typed(f.Default, f.FlagType()) || f.Value == "" && f.Default != false):
			return broken("extension.default", data(f.Name))
		case f.Values != nil && (f.Value == "" || !fitValues(f.Values, f.FlagType())):
			return broken("extension.values", data(f.Name))
		}
	}
	return nil
}

// fitValues reports whether the accepted values of an argument or a flag
// are absent, or hold at least one value, each once, that its type reads.
func fitValues(values []string, t Type) bool {
	if values == nil {
		return true
	}
	for i, v := range values {
		if v == "" || !t.Reads(v) || slices.Index(values, v) != i {
			return false
		}
	}
	return len(values) > 0
}

// typed reports whether a value that TOML decoded has a type: an integer
// for an integer, an integer or a finite float for a number, since JSON has
// no NaN or infinity.
func typed(value any, t Type) bool {
	switch v := value.(type) {
	case string:
		return t == TypeString
	case bool:
		return t == TypeBoolean
	case int64:
		return t == TypeInteger || t == TypeNumber
	case float64:
		return t == TypeNumber && !math.IsNaN(v) && !math.IsInf(v, 0)
	}
	return false
}

// checkGroups applies the rules of flag groups to one command: each names
// own flags, at least two and each once, with 0 <= min <= max <= flags and
// max >= 1.
func checkGroups(path string, c Command) error {
	for _, g := range c.FlagGroups {
		bad := len(g.Flags) < 2 || g.Min < 0 || g.Min > g.Max || g.Max < 1 || g.Max > len(g.Flags)
		for j, name := range g.Flags {
			own := slices.ContainsFunc(c.Flags, func(f Flag) bool { return f.Name == name })
			bad = bad || !own || slices.Index(g.Flags, name) != j
		}
		if bad {
			return broken("extension.flag_group", map[string]any{"Path": path, "Value": "--" + strings.Join(g.Flags, ", --")})
		}
	}
	return nil
}
