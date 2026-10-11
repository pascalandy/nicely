// Package command turns command declarations into a command line: the
// commands and their flags, the global flags, help pages, and printed
// answers. Core and every Go extension build their command line with it, so
// every command parses, helps, and fails alike.
package command

import (
	"errors"
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
	use := []string{d.Path[len(d.Path)-1]}
	for _, a := range d.Args {
		use = append(use, "["+a.Name+"]")
	}
	c := &cobra.Command{
		Use:     strings.Join(use, " "),
		Short:   text.T(d.Summary),
		Example: strings.Join(d.Examples, "\n"),
		Args: func(c *cobra.Command, args []string) error {
			if len(args) <= len(d.Args) {
				return nil
			}
			extra := args[len(d.Args)]
			if c.HasSubCommands() {
				return Fail(UnknownCommand(p, c, extra))
			}
			return Fail(contract.Problem{
				Code:    contract.UsageInvalid,
				Message: p.T("usage.unexpected_argument", map[string]any{"Arg": extra}),
				Hint:    c.CommandPath() + " --help",
			})
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
	}
	for _, f := range d.Flags {
		addFlag(c.Flags(), text, f)
	}
	return c
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

// UnknownCommand reports a word that names no command under c.
func UnknownCommand(p *i18n.Printer, c *cobra.Command, word string) contract.Problem {
	return contract.Problem{
		Code:    contract.UsageInvalid,
		Message: p.T("usage.unknown_command", map[string]any{"Name": word}),
		Hint:    c.CommandPath() + " --help",
	}
}

// valueAnnotation keeps the declared name of a flag's value, such as tag.
const valueAnnotation = "ncly-value"

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
