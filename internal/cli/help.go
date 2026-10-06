package cli

import (
	"fmt"
	"strings"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/pascalandy/nicely/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCommand = contract.Command{
	Summary: "root.summary",
	Examples: []string{
		`ncly completion zsh > "${fpath[1]}/_ncly"`,
		`ncly --version`,
		`NCLY_JSON=1 ncly help`,
	},
}

var helpDeclaration = contract.Command{
	Path:     []string{"help"},
	Summary:  "help.summary",
	Examples: []string{`ncly help`, `ncly help completion zsh`},
}

// valueAnnotation keeps the declared name of a flag's value, such as tag.
const valueAnnotation = "ncly-value"

// helpFunc prints the help page of a command from its declaration, in the
// active language.
func helpFunc(p *i18n.Printer, noColor *bool) func(*cobra.Command, []string) {
	return func(c *cobra.Command, _ []string) {
		// Reading the flags first merges the global flags into c.Flags(),
		// which UseLine needs to add [flags]. Cobra skips that merge when it
		// reaches c without parsing, as in ncly help completion zsh.
		local, inherited := c.LocalFlags(), c.InheritedFlags()
		page := tui.HelpPage{
			Summary:       c.Short,
			UsageTitle:    p.T("help.usage"),
			Usage:         c.UseLine(),
			ExamplesTitle: p.T("help.examples"),
		}
		if c.Example != "" {
			page.Examples = strings.Split(c.Example, "\n")
		}
		var commands []tui.Entry
		for _, sub := range c.Commands() {
			if sub.IsAvailableCommand() {
				commands = append(commands, tui.Entry{Name: sub.Name(), Summary: sub.Short})
			}
		}
		if len(commands) > 0 {
			page.Usage = c.CommandPath() + " <command> [flags]"
			page.Sections = append(page.Sections, tui.Section{Title: p.T("help.commands"), Entries: commands, Colon: true})
		}
		if flags := flagEntries(local); len(flags) > 0 {
			page.Sections = append(page.Sections, tui.Section{Title: p.T("help.flags"), Entries: flags})
		}
		if flags := flagEntries(inherited); len(flags) > 0 {
			page.Sections = append(page.Sections, tui.Section{Title: p.T("help.global_flags"), Entries: flags})
		}
		_, _ = fmt.Fprint(tui.Output(c.OutOrStdout(), noColor), tui.Help(page))
	}
}

func flagEntries(set *pflag.FlagSet) []tui.Entry {
	var entries []tui.Entry
	set.VisitAll(func(f *pflag.Flag) {
		name := "    --" + f.Name
		if f.Shorthand != "" {
			name = "-" + f.Shorthand + ", --" + f.Name
		}
		if value := f.Annotations[valueAnnotation]; len(value) > 0 {
			name += " <" + value[0] + ">"
		}
		entries = append(entries, tui.Entry{Name: name, Summary: f.Usage})
	})
	return entries
}
