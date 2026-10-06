package cli

import (
	"cmp"
	"errors"
	"os"
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
	// values holds, by long name, the last value of each known flag that
	// takes one, and the last valid value of each known switch.
	values map[string]string
}

// scan reads args with the flags of cmd and the same rules as the parser: a
// flag that takes a value takes the next argument even when it starts with
// "-". An unknown flag counts as a switch, and scanning goes on after it, so
// help and machine mode hold wherever the parser would stop.
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
// counts as a switch, so -zh still asks for help. The text after = is the
// value of the letter before it, so -z=h does not.
func (line *commandLine) shorthands(flags *pflag.FlagSet, group string, next func() (string, bool)) {
	for j := 0; j < len(group); j++ {
		f := flags.ShorthandLookup(group[j : j+1])
		rest := group[j+1:]
		switch {
		case f == nil && strings.HasPrefix(rest, "="):
			return
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
	if f.Value.Type() == "bool" {
		if _, err := strconv.ParseBool(value); err != nil {
			return
		}
	}
	line.values[f.Name] = value
}

// switchOn reports the last valid value of a switch. An invalid value, such
// as --json=bad, is the parser's to report, and never cancels an earlier one.
func (line commandLine) switchOn(name string) (on, given bool) {
	value, given := line.values[name]
	on, _ = strconv.ParseBool(value)
	return on, given
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
	json    bool
	noColor bool
	// lang is the language setting before matching, such as fr_CA.UTF-8.
	lang string
}

// resolveGlobals applies the precedence of cli-spec.md. It reads the config
// only when neither --lang nor NCLY_LANG sets the language.
func resolveGlobals(line commandLine) globals {
	env := os.Getenv
	switchOr := func(name string, fallback bool) bool {
		if on, given := line.switchOn(name); given {
			return on
		}
		return fallback
	}
	g := globals{
		json:    switchOr("json", env("NCLY_JSON") == "1"),
		noColor: switchOr("no-color", env("NO_COLOR") != "" || env("TERM") == "dumb"),
		lang:    cmp.Or(line.values["lang"], env("NCLY_LANG")),
	}
	if g.lang == "" {
		g.lang = cmp.Or(configLang(), env("LC_ALL"), env("LC_MESSAGES"), env("LANG"))
	}
	return g
}

// addFlag registers a declared flag with its catalog description.
func addFlag(set *pflag.FlagSet, p *i18n.Printer, f contract.Flag) {
	if f.Value == "" {
		set.BoolP(f.Name, f.Shorthand, false, p.T(f.Summary))
		return
	}
	set.StringP(f.Name, f.Shorthand, "", p.T(f.Summary))
	_ = set.SetAnnotation(f.Name, valueAnnotation, []string{f.Value})
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
