package cli

import (
	"io"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

// The completion command lives beside the root, because Cobra writes each
// script from the whole command tree.

var completionCommand = contract.Command{
	Path:    []string{"completion"},
	Summary: "completion.summary",
	Effects: &contract.Effects{},
	Examples: []string{
		`eval "$(ncly completion zsh)"`,
		`ncly completion bash > ~/.local/share/bash-completion/completions/ncly`,
		`ncly completion fish > ~/.config/fish/completions/ncly.fish`,
	},
}

var shells = []struct {
	command contract.Command
	write   func(root *cobra.Command, w io.Writer) error
}{
	{
		contract.Command{
			Path:     []string{"completion", "zsh"},
			Summary:  "completion.zsh_summary",
			Effects:  &contract.Effects{},
			Result:   contract.Result{Kind: contract.ResultText},
			Examples: []string{`eval "$(ncly completion zsh)"`, `ncly completion zsh > "${fpath[1]}/_ncly"`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenZshCompletion(w) },
	},
	{
		contract.Command{
			Path:     []string{"completion", "bash"},
			Summary:  "completion.bash_summary",
			Effects:  &contract.Effects{},
			Result:   contract.Result{Kind: contract.ResultText},
			Examples: []string{`ncly completion bash > ~/.local/share/bash-completion/completions/ncly`, `eval "$(ncly completion bash)"`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenBashCompletionV2(w, true) },
	},
	{
		contract.Command{
			Path:     []string{"completion", "fish"},
			Summary:  "completion.fish_summary",
			Effects:  &contract.Effects{},
			Result:   contract.Result{Kind: contract.ResultText},
			Examples: []string{`ncly completion fish > ~/.config/fish/completions/ncly.fish`, `ncly completion fish | source`},
		},
		func(root *cobra.Command, w io.Writer) error { return root.GenFishCompletion(w, true) },
	},
}

func addCompletion(p *i18n.Printer, root *cobra.Command) {
	group := command.Build(p, p, completionCommand, nil)
	for _, s := range shells {
		group.AddCommand(command.Build(p, p, s.command, func(c *cobra.Command, _ []string) error {
			return s.write(root, c.OutOrStdout())
		}))
	}
	root.AddCommand(group)
}
