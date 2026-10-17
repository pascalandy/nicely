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

// invalid checks the values of a parsed call against its declaration
// before the command runs, as Command descriptions in contract.md asks: the
// type and the accepted values of each value that the call gave, and the
// flag groups that it sets too many flags of. A value that the call lacks
// is for missing to report. It returns the first problem, or nil.
func invalid(p *i18n.Printer, c *cobra.Command, d contract.Command, args []string) *contract.Problem {
	usage := func(message string) *contract.Problem {
		problem := HelpUsage(c, message)
		return &problem
	}
	for i, a := range d.Args {
		values := args[min(i, len(args)):]
		if !a.Variadic {
			values = values[:min(1, len(values))]
		}
		for _, v := range values {
			if !accepts(a.ArgType(), a.Values, v) {
				return usage(p.T("usage.invalid_argument_value", map[string]any{"Name": a.Name, "Value": v}))
			}
		}
	}
	for _, f := range d.Flags {
		flag := c.Flags().Lookup(f.Name)
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
		if setFlags(c, g) > g.Max {
			return usage(groupMessage(p, g, g.Max))
		}
	}
	return nil
}

// lack is a value that a call lacks: a required argument, a required flag,
// or a flag group whose min the call does not reach. One field is set.
type lack struct {
	arg   *contract.Arg
	flag  *contract.Flag
	group *contract.FlagGroup
}

// lacking lists the values that a call lacks, in the order of the usage
// line: each argument, each flag, then each group.
func lacking(c *cobra.Command, d contract.Command, args []string) []lack {
	var list []lack
	for i := range d.Args {
		if d.Args[i].Required && len(args) <= i {
			list = append(list, lack{arg: &d.Args[i]})
		}
	}
	for i := range d.Flags {
		if d.Flags[i].Required && !c.Flags().Lookup(d.Flags[i].Name).Changed {
			list = append(list, lack{flag: &d.Flags[i]})
		}
	}
	for i := range d.FlagGroups {
		if setFlags(c, d.FlagGroups[i]) < d.FlagGroups[i].Min {
			list = append(list, lack{group: &d.FlagGroups[i]})
		}
	}
	return list
}

// missing reports the values that a call lacks as USAGE_INVALID: the
// message names the first one, and the hint shows the full command, as
// Modes in contract.md asks. The hint does not run as is, so the problem has
// no next.
func missing(p *i18n.Printer, c *cobra.Command, d contract.Command, args []string, lacks []lack) contract.Problem {
	var message string
	switch first := lacks[0]; {
	case first.arg != nil:
		message = p.T("usage.missing_argument", map[string]any{"Name": first.arg.Name})
	case first.flag != nil:
		message = p.T("usage.missing_flag", map[string]any{"Flag": "--" + first.flag.Name})
	default:
		message = groupMessage(p, *first.group, first.group.Min)
	}
	return contract.Problem{Code: contract.UsageInvalid, Message: message, Hint: commandLine(c, d, args, lacks)}
}

// groupMessage says how many flags of a group to give, where count is the
// bound that the call breaks.
func groupMessage(p *i18n.Printer, g contract.FlagGroup, count int) string {
	data := map[string]any{"Flags": "--" + strings.Join(g.Flags, ", --")}
	switch {
	case g.Min == g.Max:
		return p.N("usage.flag_group_exactly", g.Min, data)
	case count == g.Min:
		return p.N("usage.flag_group_at_least", g.Min, data)
	}
	return p.N("usage.flag_group_at_most", g.Max, data)
}

// setFlags counts the flags of a group that the call set.
func setFlags(c *cobra.Command, g contract.FlagGroup) int {
	set := 0
	for _, name := range g.Flags {
		if c.Flags().Lookup(name).Changed {
			set++
		}
	}
	return set
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
