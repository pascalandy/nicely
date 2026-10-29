package cli

import (
	"context"
	_ "embed"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
)

//go:embed auth_revoke_schema.json
var authRevokeSchema []byte

var authRevokeCommand = contract.Command{
	Path:    []string{"auth", "revoke"},
	Summary: "auth.revoke.summary",
	Args:    grantArgs,
	Flags:   []contract.Flag{{Name: forceFlag.Name, Shorthand: forceFlag.Shorthand, Summary: "auth.revoke.force"}},
	Result:  contract.Result{Kind: contract.ResultObject, Schema: mustSchema(authRevokeSchema), DryRunSchema: mustSchema(authGrantDryRunSchema)},
	Effects: &contract.Effects{UserWrites: true},
	Modes:   contract.Modes{DryRun: true},
	Examples: []string{
		`ncly auth revoke deepgram transcript`,
		`ncly auth revoke deepgram transcript --dry-run --json`,
		`ncly auth revoke deepgram transcript --force --json`,
	},
}

// addAuthRevoke adds ncly auth revoke, which removes the grant of a
// service to a slug, whatever its source, even after the extension is uninstalled.
func addAuthRevoke(p *i18n.Printer, g command.Globals, group *cobra.Command, extensions []*extension) {
	c := command.Build(p, p, authRevokeCommand, func(c *cobra.Command, args []string) error {
		service, domain := args[0], args[1]
		data := map[string]any{"Service": service, "Domain": domain}
		if problem := badService(p, service); problem != nil {
			return command.Fail(*problem)
		}
		file := grantsPath()
		grants, err := readGrants(file)
		if err != nil {
			return grantsFailure(context.Background(), p, err)
		}
		existing := findGrant(grants, service, domain)
		if existing == nil {
			return command.Fail(noGrant(p, data))
		}
		force, _ := c.Flags().GetBool(forceFlag.Name)
		if dryRun, _ := c.Flags().GetBool(contract.DryRunFlag.Name); dryRun {
			plan := []plannedGrant{{Key: service, Action: "revoke", Domain: domain, Effects: plannedEffects{UserWrites: true}}}
			pending := []pendingItem{}
			if !force {
				pending = append(pending, pendingItem{ID: "auth.confirmation", Message: p.T("auth.pending_revoke", data)})
			}
			return plannedGrantChange(p, plan, pending, p.T("auth.planned_revoke", data))
		}
		ctx := runner.OnSignal()
		if !force {
			forced := []string{"ncly", "auth", "revoke", service, domain, "--force"}
			question := confirmation{
				question: p.T("auth.revoke_question", data),
				facts:    pathFacts(p, existing.Path),
				required: p.T("auth.revoke_required", data),
				declined: p.T("auth.revoke_declined", data),
				hint:     strings.Join(forced, " "),
				next:     forced,
			}
			if err := question.ask(ctx, p, g, interactive(g, c.OutOrStdout()), c.ErrOrStderr()); err != nil {
				return err
			}
		}
		answer := grantChange{Service: service, Domain: domain, Action: "revoke"}
		err = changeGrants(ctx, file, func(current []grant) ([]grant, error) {
			found := findGrant(current, service, domain)
			if found == nil {
				return nil, command.Fail(noGrant(p, data))
			}
			answer.Path = found.Path
			return slices.DeleteFunc(current, func(g grant) bool { return g.Service == service && g.Domain == domain }), nil
		})
		if err != nil {
			return grantsFailure(ctx, p, err)
		}
		return grantAnswer(p, answer, "auth.revoked")
	})
	c.ValidArgsFunction = grantCompletion(extensions, func(service string) []string {
		grants, _ := readGrants(grantsPath())
		var domains []string
		for _, g := range grants {
			if g.Service == service {
				domains = append(domains, g.Domain)
			}
		}
		slices.Sort(domains)
		return domains
	})
	group.AddCommand(c)
}
