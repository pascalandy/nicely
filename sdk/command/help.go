package command

import (
	"fmt"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelpFunc prints the help page of a command from its declaration, in the
// active language. The commands of a group, such as the extensions under
// core's root, get a section of their own, titled by the group.
func HelpFunc(p *i18n.Printer, noColor bool) func(*cobra.Command, []string) {
	return func(c *cobra.Command, _ []string) {
		// Reading the flags first merges the global flags into c.Flags(),
		// which UseLine needs to add [flags]. Cobra skips that merge when it
		// reaches c without parsing, as in ncly help completion zsh.
		local, inherited := c.LocalFlags(), c.InheritedFlags()
		page := tui.HelpPage{
			Summary:       c.Short,
			Note:          effectsLine(p, c.Annotations[effectsAnnotation]),
			UsageTitle:    p.T("help.usage"),
			Usage:         c.UseLine(),
			ExamplesTitle: p.T("help.examples"),
		}
		if c.Example != "" {
			page.Examples = strings.Split(c.Example, "\n")
		}
		sections := []tui.Section{{Title: p.T("help.commands"), Colon: true}}
		groups := map[string]int{"": 0}
		for _, g := range c.Groups() {
			groups[g.ID] = len(sections)
			sections = append(sections, tui.Section{Title: g.Title, Colon: true})
		}
		for _, sub := range c.Commands() {
			if sub.IsAvailableCommand() {
				s := &sections[groups[sub.GroupID]]
				s.Entries = append(s.Entries, tui.Entry{Name: sub.Name(), Summary: sub.Short})
			}
		}
		for _, s := range sections {
			if len(s.Entries) > 0 {
				// A group's usage names its commands, while a command that
				// runs keeps its own, which its declaration gives.
				if c.Annotations[runsAnnotation] == "" {
					page.Usage = c.CommandPath() + " <command> [flags]"
				}
				page.Sections = append(page.Sections, s)
			}
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

// effectIDs holds the catalog ID of each effect that a help page names.
var effectIDs = map[string]string{
	"user_writes": "help.effect_user_writes",
	"network":     "help.effect_network",
	"paid":        "help.effect_paid",
}

// effectsLine names the effects of a command in one line, and --dry-run
// when the command supports it, or returns "" for a command without effects.
func effectsLine(p *i18n.Printer, annotation string) string {
	var effects []string
	dryRun := false
	for _, name := range strings.Fields(annotation) {
		if name == "dry_run" {
			dryRun = true
			continue
		}
		effects = append(effects, p.T(effectIDs[name]))
	}
	if len(effects) == 0 {
		return ""
	}
	data := map[string]any{"Effects": strings.Join(effects, p.T("help.effect_separator")), "Flag": "--" + contract.DryRunFlag.Name}
	if dryRun {
		return p.T("help.effects_dry_run", data)
	}
	return p.T("help.effects", data)
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
