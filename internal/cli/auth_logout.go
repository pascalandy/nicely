package cli

import (
	_ "embed"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
)

//go:embed auth_logout_schema.json
var authLogoutSchema []byte

var authLogoutCommand = contract.Command{
	Path:    []string{"auth", "logout"},
	Summary: "auth.logout.summary",
	Args:    []contract.Arg{{Name: "service", Summary: "auth.service", Required: true}},
	Flags:   []contract.Flag{{Name: forceFlag.Name, Shorthand: forceFlag.Shorthand, Summary: "auth.logout.force"}},
	Result:  contract.Result{Kind: contract.ResultObject, Schema: mustSchema(authLogoutSchema), DryRunSchema: mustSchema(authDryRunSchema)},
	Effects: &contract.Effects{UserWrites: true},
	Modes:   contract.Modes{DryRun: true},
	Examples: []string{
		`ncly auth logout deepgram`,
		`ncly auth logout deepgram --dry-run --json`,
		`ncly auth logout deepgram --force --json`,
	},
}

// addAuthLogout adds ncly auth logout, which removes the key of a service
// from the keychain. It never changes the environment.
func addAuthLogout(p *i18n.Printer, g command.Globals, group *cobra.Command, extensions []*extension, b keychain.Backend) {
	c := command.Build(p, p, authLogoutCommand, func(c *cobra.Command, args []string) error {
		service := args[0]
		if problem := badService(p, service); problem != nil {
			return command.Fail(*problem)
		}
		if dryRun, _ := c.Flags().GetBool(contract.DryRunFlag.Name); dryRun {
			return plannedChange(p, service, "logout")
		}
		force, _ := c.Flags().GetBool(forceFlag.Name)
		out := c.OutOrStdout()
		human := interactive(g, out)
		data := map[string]any{"Service": service, "Variable": keyVariable(service)}
		ctx := runner.OnSignal()
		k := keychain.New(b, human)
		_, stored, err := k.Read(ctx, service)
		switch {
		case errors.Is(err, keychain.ErrUnavailable):
			return command.Fail(contract.Problem{Code: contract.KeyringUnavailable, Message: p.T("auth.logout_unavailable", data), Hint: p.T("auth.logout_unavailable_hint", data)})
		case err != nil:
			return stopped(ctx)
		case !stored:
			hint := "ncly auth status"
			if os.Getenv(keyVariable(service)) != "" {
				hint = p.T("auth.logout_env_hint", data)
			}
			return command.Fail(contract.Problem{Code: contract.NotFound, Message: p.T("auth.logout_none", data), Hint: hint, Next: []string{"ncly", "auth", "status"}})
		}
		if !force {
			hint := strings.Join([]string{"ncly", "auth", "logout", service, "--force"}, " ")
			if !human {
				return command.Fail(contract.Problem{Code: contract.ConfirmationRequired, Message: p.T("auth.remove_required", data), Hint: hint})
			}
			yes, err := confirm(ctx, os.Stdin, out, p.T("auth.remove_question", data), yesWords(p.T("auth.yes")))
			switch {
			case ctx.Err() != nil:
				return stopped(ctx)
			case err != nil:
				return err
			case !yes:
				return command.Fail(contract.Problem{Code: contract.ConfirmationRequired, Message: p.T("auth.remove_declined", data), Hint: hint})
			}
		}
		if err := k.Remove(ctx, service); err != nil {
			unknown := writeUnknown(p, service, declaredServices(extensions), "auth.remove_unknown")
			if errors.Is(err, keychain.ErrWriteUnknown) {
				return command.Fail(unknown)
			}
			return &answered{outcome: contract.Outcome{Signal: contract.StopSignal(runner.SignalOf(ctx)), Errors: []contract.Problem{unknown}}, text: func(io.Writer) error { return nil }}
		}
		return &answered{
			outcome: contract.Outcome{Data: changed{Service: service, Action: "logout"}},
			text: func(w io.Writer) error {
				_, err := io.WriteString(w, p.T("auth.removed", data)+"\n")
				return err
			},
		}
	})
	c.ValidArgsFunction = serviceCompletion(extensions)
	group.AddCommand(c)
}
