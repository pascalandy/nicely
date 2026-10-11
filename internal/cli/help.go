package cli

import (
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

var helpDeclaration = contract.Command{
	Path:     []string{"help"},
	Summary:  "help.summary",
	Examples: []string{`ncly help`, `ncly help completion zsh`},
}

// helpCommand prints the same help as --help, and fails on an unknown topic.
func helpCommand(p *i18n.Printer) *cobra.Command {
	return &cobra.Command{
		Use:     "help [command]",
		Short:   p.T(helpDeclaration.Summary),
		Example: strings.Join(helpDeclaration.Examples, "\n"),
		RunE: func(c *cobra.Command, args []string) error {
			target, rest, err := c.Root().Find(args)
			switch {
			case err != nil:
				return command.Fail(command.UnknownCommand(p, c.Root(), args[0]))
			case len(rest) > 0:
				return command.Fail(command.UnknownCommand(p, target, rest[0]))
			}
			return target.Help()
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
			var names []cobra.Completion
			if target, rest, err := c.Root().Find(args); err == nil && len(rest) == 0 {
				for _, sub := range target.Commands() {
					if sub.IsAvailableCommand() {
						names = append(names, cobra.CompletionWithDesc(sub.Name(), sub.Short))
					}
				}
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}
}
