package cli

import (
	"strings"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/spf13/cobra"
)

// build makes the command that a declaration describes, with catalog text. A
// nil run makes a command that groups others: alone it prints its help, and
// an unknown word after it is a usage error.
func build(p *i18n.Printer, d contract.Command, run func(*cobra.Command, []string) error) *cobra.Command {
	c := &cobra.Command{
		Use:     d.Path[len(d.Path)-1],
		Short:   p.T(d.Summary),
		Example: strings.Join(d.Examples, "\n"),
		Args: func(c *cobra.Command, args []string) error {
			if len(args) <= len(d.Args) {
				return nil
			}
			extra := args[len(d.Args)]
			if c.HasSubCommands() {
				return fail(unknownCommand(p, c, extra))
			}
			return fail(contract.Problem{
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
		addFlag(c.Flags(), p, f)
	}
	return c
}

// unknownCommand reports a word that names no command under c.
func unknownCommand(p *i18n.Printer, c *cobra.Command, word string) contract.Problem {
	return contract.Problem{
		Code:    contract.UsageInvalid,
		Message: p.T("usage.unknown_command", map[string]any{"Name": word}),
		Hint:    c.CommandPath() + " --help",
	}
}
