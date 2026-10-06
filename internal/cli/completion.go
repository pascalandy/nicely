package cli

import (
	"io"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/spf13/cobra"
)

// The completion command lives beside the root, because Cobra writes each
// script from the whole command tree.

var completionCommand = contract.Command{
	Path:    []string{"completion"},
	Summary: "completion.summary",
	Examples: []string{
		`ncly completion zsh > "${fpath[1]}/_ncly"`,
		`ncly completion bash > ~/.local/share/bash-completion/completions/ncly`,
		`ncly completion fish > ~/.config/fish/completions/ncly.fish`,
	},
}

// shells pairs each shell's declaration with the writer of its script.
var shells = []struct {
	command contract.Command
	write   func(root *cobra.Command, w io.Writer) error
}{
	{
		contract.Command{
			Path:     []string{"completion", "zsh"},
			Summary:  "completion.zsh_summary",
			Examples: []string{`ncly completion zsh > "${fpath[1]}/_ncly"`, `source <(ncly completion zsh)`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenZshCompletion(w) },
	},
	{
		contract.Command{
			Path:     []string{"completion", "bash"},
			Summary:  "completion.bash_summary",
			Examples: []string{`ncly completion bash > ~/.local/share/bash-completion/completions/ncly`, `source <(ncly completion bash)`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenBashCompletionV2(w, true) },
	},
	{
		contract.Command{
			Path:     []string{"completion", "fish"},
			Summary:  "completion.fish_summary",
			Examples: []string{`ncly completion fish > ~/.config/fish/completions/ncly.fish`, `ncly completion fish | source`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenFishCompletion(w, true) },
	},
}

func addCompletion(b builder, root *cobra.Command) {
	group := b.build(completionCommand, nil)
	for _, s := range shells {
		group.AddCommand(b.build(s.command, func(c *cobra.Command, _ []string) error {
			return s.write(root, c.OutOrStdout())
		}))
	}
	root.AddCommand(group)
}
