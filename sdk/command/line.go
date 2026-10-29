package command

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Line is what the parser will find in the arguments of one command.
type Line struct {
	// Words holds the positional words before "--", in order.
	Words []string
	// Operands holds the words after "--", which are never flags or commands.
	Operands []string
	// values holds, by long name, the last value of each known flag that
	// takes one, and the last valid value of each known switch.
	values map[string]string
}

// Scan reads args with the flags of cmd and the same rules as the parser: a
// flag that takes a value takes the next argument even when it starts with
// "-". An unknown flag counts as a switch, and scanning goes on after it, so
// help and machine mode hold wherever the parser would stop.
func Scan(cmd *cobra.Command, args []string) Line {
	cmd.InheritedFlags() // merges the persistent flags of the parents into cmd.Flags()
	flags := cmd.Flags()
	line := Line{values: map[string]string{}}
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
			line.Operands = args[i+1:]
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
			line.Words = append(line.Words, arg)
		}
	}
	return line
}

// shorthands reads a group such as -vh, -lfr, or -l fr. An unknown letter
// counts as a switch, so -zh still asks for help. The text after = is the
// value of the letter before it, so -z=h does not.
func (line *Line) shorthands(flags *pflag.FlagSet, group string, next func() (string, bool)) {
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

func (line *Line) set(f *pflag.Flag, value string) {
	if f.Value.Type() == "bool" {
		if _, err := strconv.ParseBool(value); err != nil {
			return
		}
	}
	line.values[f.Name] = value
}

// Switch returns the last valid value of a switch, by long name, and
// whether the line sets it at all. An invalid value, such as --json=bad, is
// the parser's to report, and never cancels an earlier one. The test kit
// reads the switches of each call of a scenario with it, so it sees what
// ncly sees.
func (line Line) Switch(name string) (on, given bool) {
	value, given := line.values[name]
	on, _ = strconv.ParseBool(value)
	return on, given
}

// Help reports whether the line asks for help.
func (line Line) Help() bool {
	on, _ := line.Switch("help")
	return on
}

// Version reports whether the line asks for the version.
func (line Line) Version() bool {
	on, _ := line.Switch("version")
	return on
}

// ModuleVersion returns the version that Go stamped from the module or the
// VCS of the build, or a development marker.
func ModuleVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}
	return "v0.0.0-dev"
}

// Globals holds the resolved global flags. A flag wins over its variable.
type Globals struct {
	JSON    bool
	NoInput bool
	NoColor bool
	Verbose bool
	// Lang is the language setting before matching, such as fr_CA.UTF-8.
	Lang string
}

// Vars returns the variables that carry the resolved global flags to a
// program, as Global flags in contract.md names them. An empty value means
// that the variable is unset.
func (g Globals) Vars() map[string]string {
	on := func(set bool) string {
		if set {
			return "1"
		}
		return ""
	}
	return map[string]string{
		"NCLY_JSON":     on(g.JSON),
		"NCLY_NO_INPUT": on(g.NoInput),
		"NCLY_VERBOSE":  on(g.Verbose),
		"NO_COLOR":      on(g.NoColor),
		"NCLY_LANG":     i18n.Requested(g.Lang),
	}
}

// Resolve applies the precedence of contract.md. It reads the config
// only when neither --lang nor NCLY_LANG sets the language.
func Resolve(line Line) Globals {
	env := os.Getenv
	switchOr := func(name string, fallback bool) bool {
		if on, given := line.Switch(name); given {
			return on
		}
		return fallback
	}
	g := Globals{
		JSON:    switchOr("json", env("NCLY_JSON") == "1"),
		NoInput: switchOr("no-input", env("NCLY_NO_INPUT") == "1"),
		NoColor: switchOr("no-color", env("NO_COLOR") != "" || env("TERM") == "dumb"),
		Verbose: switchOr("verbose", env("NCLY_VERBOSE") == "1"),
	}
	g.Lang, _ = config.Lang(line.values["lang"], env, ConfigValues).Value.(string)
	return g
}

// Refused returns USAGE_INVALID for the first value that Global flags or
// Configuration in contract.md refuses, or nil. It checks the NCLY_ variable
// of each switch that the line leaves to it, which must hold 1, 0, or
// nothing. NO_COLOR follows no-color.org instead. It then checks --lang or
// NCLY_LANG, whichever sets the language.
func Refused(p *i18n.Printer, line Line) *contract.Problem {
	for _, f := range contract.GlobalFlags {
		value := os.Getenv(f.Env)
		_, given := line.Switch(f.Name)
		if f.Value != "" || !strings.HasPrefix(f.Env, "NCLY_") || given || value == "" || value == "0" || value == "1" {
			continue
		}
		data := map[string]any{"Name": f.Env, "Value": i18n.Clip(value)}
		return &contract.Problem{Code: contract.UsageInvalid, Message: p.T("usage.invalid_variable", data), Hint: p.T("usage.invalid_variable_hint", data)}
	}
	lang := config.Lang(line.values["lang"], os.Getenv, func() config.Values { return nil })
	if setting, _ := lang.Value.(string); (lang.Source == "--lang" || lang.Source == "NCLY_LANG") && !i18n.IsLanguage(setting) {
		data := map[string]any{"Name": lang.Source, "Value": i18n.Clip(setting)}
		return &contract.Problem{Code: contract.UsageInvalid, Message: p.T("usage.invalid_language", data), Hint: p.T("usage.invalid_language_hint")}
	}
	return nil
}

// ConfigValues reads the values of the config files. A load error gives
// none, so the help still works with the defaults.
func ConfigValues() config.Values {
	home, _ := os.UserHomeDir()
	values, err := config.LoadValues(config.Locate(home, os.LookupEnv))
	if err != nil {
		return nil
	}
	return values
}
