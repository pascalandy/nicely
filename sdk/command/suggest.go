package command

import (
	"slices"

	"github.com/spf13/cobra"
)

// suggestion returns the first command of c, in alphabetical order, that
// Cobra finds close to word, or "". Cobra skips hidden commands and help.
// An operand after -- can equal a command name, and then gets no suggestion.
func suggestion(c *cobra.Command, word string) string {
	// Cobra applies its default distance of 2 only when it builds its own
	// error for an unknown command.
	if c.SuggestionsMinimumDistance <= 0 {
		c.SuggestionsMinimumDistance = 2
	}
	names := c.SuggestionsFor(word)
	if word == "" || len(names) == 0 || slices.Contains(names, word) {
		return ""
	}
	return slices.Min(names)
}
