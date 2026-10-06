package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandLine is what the parser will find in the arguments of one command.
type commandLine struct {
	// words holds the positional words before "--".
	words []string
	// values holds the last value given to each known flag, by long name.
	values map[string]string
}

// scan reads args with the flags of cmd and the same rules as the parser: a
// flag that takes a value takes the next argument even when it starts with
// "-". An unknown flag counts as a switch, since the parser stops there.
func scan(cmd *cobra.Command, args []string) commandLine {
	cmd.InheritedFlags() // merges the persistent flags of the parents into cmd.Flags()
	flags := cmd.Flags()
	line := commandLine{values: map[string]string{}}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := func() (string, bool) {
			if i+1 < len(args) {
				i++
				return args[i], true
			}
			return "", false
		}
		switch {
		case arg == "--":
			return line
		case strings.HasPrefix(arg, "--"):
			name, value, inline := strings.Cut(arg[2:], "=")
			f := flags.Lookup(name)
			switch {
			case f == nil:
			case inline:
				line.values[f.Name] = value
			case f.NoOptDefVal != "":
				line.values[f.Name] = f.NoOptDefVal
			default:
				if value, ok := next(); ok {
					line.values[f.Name] = value
				}
			}
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			line.shorthands(flags, arg[1:], next)
		default:
			line.words = append(line.words, arg)
		}
	}
	return line
}

// shorthands reads a group such as -vh, -lfr, or -l fr.
func (line *commandLine) shorthands(flags *pflag.FlagSet, group string, next func() (string, bool)) {
	for j := 0; j < len(group); j++ {
		f := flags.ShorthandLookup(group[j : j+1])
		rest := group[j+1:]
		switch {
		case f == nil:
			return
		case strings.HasPrefix(rest, "="):
			line.values[f.Name] = rest[1:]
			return
		case f.NoOptDefVal != "":
			line.values[f.Name] = f.NoOptDefVal
		case rest != "":
			line.values[f.Name] = rest
			return
		default:
			if value, ok := next(); ok {
				line.values[f.Name] = value
			}
			return
		}
	}
}

// switchOn reports the value of a switch, and false when it is absent or its
// value is not a boolean.
func (line commandLine) switchOn(name string) (on, given bool) {
	value, present := line.values[name]
	if !present {
		return false, false
	}
	on, err := strconv.ParseBool(value)
	return on, err == nil
}

func (line commandLine) help() bool {
	on, _ := line.switchOn("help")
	return on
}

// globals holds the resolved global flags. A flag wins over its variable.
type globals struct {
	json bool
}

func resolveGlobals(line commandLine, lookupEnv func(string) (string, bool)) globals {
	on, given := line.switchOn("json")
	if !given {
		value, _ := lookupEnv("NCLY_JSON")
		on = value == "1"
	}
	return globals{json: on}
}

// addFlag registers a declared flag with its catalog description.
func addFlag(set *pflag.FlagSet, p *i18n.Printer, f contract.Flag) {
	if f.Value == "" {
		set.BoolP(f.Name, f.Shorthand, false, p.T(f.Summary))
		return
	}
	set.StringP(f.Name, f.Shorthand, "", p.T(f.Summary))
}

// flagProblem turns a parser error into USAGE_INVALID with catalog text.
func flagProblem(p *i18n.Printer, c *cobra.Command, err error) contract.Problem {
	problem := contract.Problem{Code: contract.UsageInvalid, Hint: c.CommandPath() + " --help"}
	var (
		notExist *pflag.NotExistError
		required *pflag.ValueRequiredError
		invalid  *pflag.InvalidValueError
		syntax   *pflag.InvalidSyntaxError
	)
	switch {
	case errors.As(err, &notExist):
		flag := "--" + notExist.GetSpecifiedName()
		if notExist.GetSpecifiedShortnames() != "" {
			flag = "-" + notExist.GetSpecifiedName()
		}
		problem.Message = p.T("usage.unknown_flag", map[string]any{"Flag": flag})
	case errors.As(err, &required):
		problem.Message = p.T("usage.flag_needs_value", map[string]any{"Flag": "--" + required.GetFlag().Name})
	case errors.As(err, &invalid):
		problem.Message = p.T("usage.invalid_flag_value", map[string]any{"Flag": "--" + invalid.GetFlag().Name, "Value": invalid.GetValue()})
	case errors.As(err, &syntax):
		problem.Message = p.T("usage.invalid_flag_syntax", map[string]any{"Arg": syntax.GetSpecifiedFlag()})
	default:
		problem.Message = p.T("usage.invalid")
	}
	return problem
}
