// Package command turns command declarations into a command line: the
// commands and their flags, the global flags, help pages, and printed
// answers. Core and every Go extension build their command line with it, so
// every command parses, helps, and fails alike.
package command

import (
	"errors"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Build makes the command that a declaration describes. text renders the
// catalog IDs of the declaration, from the catalogs of its owner, and p
// renders the SDK's own text, such as usage errors. A nil run makes a
// command that groups others: alone it prints its help, and an unknown word
// after it is a usage error.
func Build(p, text *i18n.Printer, d contract.Command, run func(*cobra.Command, []string) error) *cobra.Command {
	variadic := len(d.Args) > 0 && d.Args[len(d.Args)-1].Variadic
	c := &cobra.Command{
		Use:     usage(d),
		Short:   text.T(d.Summary),
		Example: strings.Join(d.Examples, "\n"),
		Args: func(c *cobra.Command, args []string) error {
			if variadic || len(args) <= len(d.Args) {
				// A group only prints its help, so nothing that a run needs
				// applies to it.
				if run == nil {
					return nil
				}
				if problem := invalid(p, c, d, args); problem != nil {
					return Fail(*problem)
				}
				return nil
			}
			extra := args[len(d.Args)]
			if c.HasSubCommands() {
				return Fail(UnknownCommand(p, c, extra))
			}
			return Fail(HelpUsage(c, p.T("usage.unexpected_argument", map[string]any{"Arg": i18n.Clip(extra)})))
		},
		RunE: run,
		// Beyond its declared arguments, a command takes no word, so the
		// shell must not offer file names either.
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) >= len(d.Args) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveDefault
		},
	}
	if run == nil {
		c.RunE = func(c *cobra.Command, _ []string) error { return c.Help() }
	} else {
		c.RunE = func(c *cobra.Command, args []string) error {
			args, err := complete(p, text, c, d, args)
			if err != nil {
				return err
			}
			return run(c, args)
		}
		c.Annotations = effectAnnotations(d)
		c.Annotations[runsAnnotation] = "true"
	}
	for _, f := range d.Flags {
		addFlag(c.Flags(), text, f)
	}
	for _, f := range d.ModeFlags() {
		addFlag(c.Flags(), p, f)
	}
	return c
}

// usage returns the usage of a command after its path, as Usage in
// contract.md shows it: each argument, <name> when required and [name]
// otherwise, each required flag outside a group, then each flag group, in
// parentheses when the group needs a flag and in brackets otherwise.
func usage(d contract.Command) string {
	words := []string{d.Path[len(d.Path)-1]}
	for _, a := range d.Args {
		name := a.Name
		if a.Variadic {
			name += "..."
		}
		if a.Required {
			words = append(words, "<"+name+">")
		} else {
			words = append(words, "["+name+"]")
		}
	}
	grouped := map[string]bool{}
	for _, g := range d.FlagGroups {
		for _, name := range g.Flags {
			grouped[name] = true
		}
	}
	for _, f := range d.Flags {
		if f.Required && !grouped[f.Name] {
			words = append(words, flagUsage(d, f.Name))
		}
	}
	for _, g := range d.FlagGroups {
		words = append(words, groupUsage(d, g))
	}
	return strings.Join(words, " ")
}

// flagUsage shows a flag as a usage line does: --name, or --name <value>
// for a flag that takes one.
func flagUsage(d contract.Command, name string) string {
	for _, f := range d.Flags {
		if f.Name == name && f.Value != "" {
			return "--" + name + " <" + f.Value + ">"
		}
	}
	return "--" + name
}

// groupUsage shows a flag group as a usage line does: in parentheses when
// the group needs a flag, and in brackets otherwise.
func groupUsage(d contract.Command, g contract.FlagGroup) string {
	var choices []string
	for _, name := range g.Flags {
		choices = append(choices, flagUsage(d, name))
	}
	open, end := "[", "]"
	if g.Min > 0 {
		open, end = "(", ")"
	}
	return open + strings.Join(choices, " | ") + end
}

// effectsAnnotation keeps the effects of a command that runs, which its
// help names in one line: user_writes, network, and paid, then dry_run when
// it supports --dry-run. runsAnnotation marks a command that runs, whose
// usage stays its own even when it has children.
const (
	effectsAnnotation = "ncly-effects"
	runsAnnotation    = "ncly-runs"
)

func effectAnnotations(d contract.Command) map[string]string {
	e := d.DeclaredEffects()
	var effects []string
	for _, effect := range []struct {
		name string
		on   bool
	}{{"user_writes", e.UserWrites}, {"network", e.Network}, {"paid", e.Paid}} {
		if effect.on {
			effects = append(effects, effect.name)
		}
	}
	if len(effects) == 0 {
		return map[string]string{}
	}
	if d.Modes.DryRun {
		effects = append(effects, "dry_run")
	}
	return map[string]string{effectsAnnotation: strings.Join(effects, " ")}
}

// AddManifest adds the commands that a manifest declares under root, each
// parent before its children, and returns the extension root. text
// renders the manifest's catalog IDs. run returns what each command that
// runs does with its call, and a command for which it returns nil prints
// its help.
func AddManifest(root *cobra.Command, p, text *i18n.Printer, m contract.Manifest, run func(contract.Command) func(*cobra.Command, []string) error) *cobra.Command {
	commands := map[string]*cobra.Command{}
	if m.Name != "" {
		parent, _, _ := root.Find([]string{m.Domain})
		if parent == root {
			parent = &cobra.Command{Use: m.Domain}
			root.AddCommand(parent)
		}
		commands[m.Domain] = parent
	}
	declarations := slices.Clone(m.Commands)
	slices.SortStableFunc(declarations, func(a, b contract.Command) int { return len(a.Path) - len(b.Path) })
	for _, d := range declarations {
		var r func(*cobra.Command, []string) error
		if m.Runs(d) {
			r = run(d)
		}
		c := Build(p, text, d, r)
		commands[strings.Join(d.Path, " ")] = c
		if parent, found := commands[strings.Join(d.Path[:len(d.Path)-1], " ")]; found {
			parent.AddCommand(c)
		} else {
			root.AddCommand(c)
		}
	}
	return commands[strings.Join(m.Root(), " ")]
}

// AddGlobalFlags gives a root command the global flags of the contract, and
// turns every parser error into USAGE_INVALID.
func AddGlobalFlags(root *cobra.Command, p *i18n.Printer) {
	for _, f := range contract.GlobalFlags {
		addFlag(root.PersistentFlags(), p, f)
	}
	// Cobra offers file names for a flag value, and a language tag is never one.
	_ = root.RegisterFlagCompletionFunc("lang", cobra.NoFileCompletions)
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return Fail(flagProblem(p, c, err))
	})
}

// UnknownCommand reports a word that names no command under c. When a
// command of c is close to the word, it suggests that command, whose help is
// the hint, as Usage in contract.md defines.
func UnknownCommand(p *i18n.Printer, c *cobra.Command, word string) contract.Problem {
	data := map[string]any{"Name": i18n.Clip(word)}
	if name := suggestion(c, word); name != "" {
		data["Suggestion"] = name
		return Usage(p.T("usage.unknown_command_suggestion", data), c.CommandPath()+" "+name+" --help")
	}
	return HelpUsage(c, p.T("usage.unknown_command", data))
}

// HelpUsage reports USAGE_INVALID with the help of c as its hint.
func HelpUsage(c *cobra.Command, message string) contract.Problem {
	return Usage(message, c.CommandPath()+" --help")
}

// Usage reports USAGE_INVALID whose hint is a command that runs as is, so
// the same command is its next, as Output in contract.md asks.
func Usage(message, hint string) contract.Problem {
	return contract.Problem{Code: contract.UsageInvalid, Message: message, Hint: hint, Next: strings.Fields(hint)}
}

// valueAnnotation keeps the declared name of a flag's value, such as tag.
const valueAnnotation = "ncly-value"

// addFlag registers a declared flag with its catalog description. A
// variadic flag keeps each value, commas included, in the caller's order.
// The values of a flag are checked against its type after parsing.
func addFlag(set *pflag.FlagSet, p *i18n.Printer, f contract.Flag) {
	switch {
	case f.Value == "":
		set.BoolP(f.Name, f.Shorthand, false, p.T(f.Summary))
		return
	case f.Variadic:
		set.StringArrayP(f.Name, f.Shorthand, nil, p.T(f.Summary))
	default:
		set.StringP(f.Name, f.Shorthand, "", p.T(f.Summary))
	}
	_ = set.SetAnnotation(f.Name, valueAnnotation, []string{f.Value})
}

// flagProblem turns a parser error into USAGE_INVALID with catalog text.
func flagProblem(p *i18n.Printer, c *cobra.Command, err error) contract.Problem {
	problem := HelpUsage(c, "")
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
		problem.Message = p.T("usage.unknown_flag", map[string]any{"Flag": i18n.Clip(flag)})
	case errors.As(err, &required):
		problem.Message = p.T("usage.flag_needs_value", map[string]any{"Flag": "--" + required.GetFlag().Name})
	case errors.As(err, &invalid):
		problem.Message = p.T("usage.invalid_flag_value", map[string]any{"Flag": "--" + invalid.GetFlag().Name, "Value": i18n.Clip(invalid.GetValue())})
	case errors.As(err, &syntax):
		problem.Message = p.T("usage.invalid_flag_syntax", map[string]any{"Arg": i18n.Clip(syntax.GetSpecifiedFlag())})
	default:
		problem.Message = p.T("usage.invalid")
	}
	return problem
}
