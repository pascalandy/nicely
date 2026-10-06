package cli

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandLine is what the parser will find in the arguments of one command.
type commandLine struct {
	// words holds the positional words before "--", in order.
	words []string
	// operands holds the words after "--", which are never flags or commands.
	operands []string
	// values holds every value given to each known flag, in order, by long
	// name.
	values map[string][]string
}

// scan reads args with the flags of cmd and the same rules as the parser: a
// flag that takes a value takes the next argument even when it starts with
// "-". An unknown flag counts as a switch, and scanning goes on after it, so
// help and machine mode hold wherever the parser would stop.
func scan(cmd *cobra.Command, args []string) commandLine {
	cmd.InheritedFlags() // merges the persistent flags of the parents into cmd.Flags()
	flags := cmd.Flags()
	line := commandLine{values: map[string][]string{}}
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
			line.operands = args[i+1:]
			return line
		case strings.HasPrefix(arg, "--"):
			name, value, inline := strings.Cut(arg[2:], "=")
			f := flags.Lookup(name)
			switch {
			case f == nil:
			case inline:
				line.set(f, value)
			case f.NoOptDefVal != "":
				line.set(f, f.NoOptDefVal)
			default:
				if value, ok := next(); ok {
					line.set(f, value)
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

// shorthands reads a group such as -vh, -lfr, or -l fr. An unknown letter
// counts as a switch, so -zh still asks for help.
func (line *commandLine) shorthands(flags *pflag.FlagSet, group string, next func() (string, bool)) {
	for j := 0; j < len(group); j++ {
		f := flags.ShorthandLookup(group[j : j+1])
		rest := group[j+1:]
		switch {
		case f == nil:
		case strings.HasPrefix(rest, "="):
			line.set(f, rest[1:])
			return
		case f.NoOptDefVal != "":
			line.set(f, f.NoOptDefVal)
		case rest != "":
			line.set(f, rest)
			return
		default:
			if value, ok := next(); ok {
				line.set(f, value)
			}
			return
		}
	}
}

func (line *commandLine) set(f *pflag.Flag, value string) {
	line.values[f.Name] = append(line.values[f.Name], value)
}

// switchOn reports the last valid value of a switch. An invalid value, such
// as --json=bad, is the parser's to report, and never cancels an earlier one.
func (line commandLine) switchOn(name string) (on, given bool) {
	for _, value := range slices.Backward(line.values[name]) {
		if on, err := strconv.ParseBool(value); err == nil {
			return on, true
		}
	}
	return false, false
}

func (line commandLine) help() bool {
	on, _ := line.switchOn("help")
	return on
}

func (line commandLine) version() bool {
	on, _ := line.switchOn("version")
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
