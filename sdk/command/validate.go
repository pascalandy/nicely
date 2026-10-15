package command

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// validate checks a parsed call against its declaration before the command
// runs, as Command descriptions in contract.md asks: the required arguments
// and flags, the type and the accepted values of each value, and the flag
// groups. It returns the first problem, or nil.
func validate(p *i18n.Printer, c *cobra.Command, d contract.Command, args []string) *contract.Problem {
	usage := func(message string) *contract.Problem {
		problem := HelpUsage(c, message)
		return &problem
	}
	for i, a := range d.Args {
		values := args[min(i, len(args)):]
		if !a.Variadic {
			values = values[:min(1, len(values))]
		}
		if a.Required && len(values) == 0 {
			return usage(p.T("usage.missing_argument", map[string]any{"Name": a.Name}))
		}
		for _, v := range values {
			if !accepts(a.ArgType(), a.Values, v) {
				return usage(p.T("usage.invalid_argument_value", map[string]any{"Name": a.Name, "Value": v}))
			}
		}
	}
	for _, f := range d.Flags {
		flag := c.Flags().Lookup(f.Name)
		if f.Required && !flag.Changed {
			return usage(p.T("usage.missing_flag", map[string]any{"Flag": "--" + f.Name}))
		}
		if !flag.Changed || f.Value == "" {
			continue
		}
		for _, v := range FlagValues(flag) {
			if !accepts(f.FlagType(), f.Values, v) {
				return usage(p.T("usage.invalid_flag_value", map[string]any{"Flag": "--" + f.Name, "Value": v}))
			}
		}
	}
	for _, g := range d.FlagGroups {
		set := 0
		for _, name := range g.Flags {
			if c.Flags().Lookup(name).Changed {
				set++
			}
		}
		data := map[string]any{"Flags": "--" + strings.Join(g.Flags, ", --")}
		switch {
		case set >= g.Min && set <= g.Max:
		case g.Min == g.Max:
			return usage(p.N("usage.flag_group_exactly", g.Min, data))
		case set < g.Min:
			return usage(p.N("usage.flag_group_at_least", g.Min, data))
		default:
			return usage(p.N("usage.flag_group_at_most", g.Max, data))
		}
	}
	return nil
}

// FlagValues returns the values of a flag that the caller set: each value
// of a variadic flag in the caller's order, or the last value of any other.
func FlagValues(flag *pflag.Flag) []string {
	if list, isList := flag.Value.(pflag.SliceValue); isList {
		return list.GetSlice()
	}
	return []string{flag.Value.String()}
}

// accepts reports whether a word is a value of a type, among the accepted
// values when the declaration lists them.
func accepts(t contract.Type, values []string, word string) bool {
	if len(values) > 0 && !slices.Contains(values, word) {
		return false
	}
	switch t {
	case contract.TypeBoolean:
		_, err := strconv.ParseBool(word)
		return err == nil
	case contract.TypeInteger:
		_, err := strconv.ParseInt(word, 10, 64)
		return err == nil
	case contract.TypeNumber:
		n, err := strconv.ParseFloat(word, 64)
		return err == nil && !math.IsNaN(n) && !math.IsInf(n, 0)
	}
	return true
}
