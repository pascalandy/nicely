package cli

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
)

//go:embed auth_grant_schema.json
var authGrantSchema []byte

//go:embed auth_revoke_schema.json
var authRevokeSchema []byte

//go:embed auth_grant_dry_run_schema.json
var authGrantDryRunSchema []byte

// grantArgs are the arguments of grant and revoke: the service, then the
// domain.
var grantArgs = []contract.Arg{
	{Name: "service", Summary: "auth.service", Required: true},
	{Name: "domain", Summary: "auth.domain", Required: true},
}

var authGrantCommand = contract.Command{
	Path:    []string{"auth", "grant"},
	Summary: "auth.grant.summary",
	Args:    grantArgs,
	Result:  contract.Result{Kind: contract.ResultObject, Schema: mustSchema(authGrantSchema), DryRunSchema: mustSchema(authGrantDryRunSchema)},
	Effects: &contract.Effects{UserWrites: true},
	Modes:   contract.Modes{DryRun: true},
	Examples: []string{
		`ncly auth grant deepgram transcript`,
		`ncly auth grant deepgram transcript --dry-run --json`,
	},
}

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

// grantChange is the answer of a grant or a revocation that succeeded.
type grantChange struct {
	Service string  `json:"service"`
	Domain  string  `json:"domain"`
	Path    *string `json:"path"`
	Action  string  `json:"action"`
}

// plannedGrant is the plan of a dry run of grant or revoke.
type plannedGrant struct {
	Key     string         `json:"key"`
	Action  string         `json:"action"`
	Domain  string         `json:"domain"`
	Effects plannedEffects `json:"effects"`
}

// grantDryRun is the answer of grant or revoke under --dry-run, as ncly auth
// in core-spec.md defines it.
type grantDryRun struct {
	DryRun        bool           `json:"dry_run"`
	Plan          []plannedGrant `json:"plan"`
	PendingChecks []pendingItem  `json:"pending_checks"`
}

// addAuthGrant adds ncly auth grant, which lets the installed extension of
// a domain receive the key of a service once a human confirms it in a
// terminal, as Grants in contract.md defines. It reads no key.
func addAuthGrant(p *i18n.Printer, g command.Globals, group *cobra.Command, extensions []*extension) {
	c := command.Build(p, p, authGrantCommand, func(c *cobra.Command, args []string) error {
		service, domain := args[0], args[1]
		data := map[string]any{"Service": service, "Domain": domain}
		if problem := badService(p, service); problem != nil {
			return command.Fail(*problem)
		}
		e, problem := grantTarget(p, extensions, service, domain)
		if problem != nil {
			return command.Fail(*problem)
		}
		file := grantsPath()
		grants, err := readGrants(file)
		if err != nil {
			return grantsFailure(context.Background(), p, err)
		}
		dryRun, _ := c.Flags().GetBool(contract.DryRunFlag.Name)
		if existing := findGrant(grants, service, domain); existing != nil {
			if dryRun {
				return plannedGrantChange(p, nil, nil, p.T("auth.planned_nothing", data))
			}
			return grantAnswer(p, grantChange{Service: service, Domain: domain, Path: existing.Path, Action: "unchanged"}, "auth.grant_unchanged")
		}
		if dryRun {
			plan := []plannedGrant{{Key: service, Action: "grant", Domain: domain, Effects: plannedEffects{UserWrites: true}}}
			pending := []pendingItem{{ID: "auth.confirmation", Message: p.T("auth.pending_grant", data)}}
			return plannedGrantChange(p, plan, pending, p.T("auth.planned_grant", data))
		}
		call := []string{"ncly", "auth", "grant", service, domain}
		data["Command"] = strings.Join(call, " ")
		human := interactive(g, c.OutOrStdout())
		if !human {
			return command.Fail(contract.Problem{Code: contract.TerminalRequired, Message: p.T("auth.grant_terminal_required", data), Hint: p.T("auth.terminal_required_hint", data), Next: call})
		}
		path := sourceOf(e)
		ctx := runner.OnSignal()
		question := confirmation{
			question: p.T("auth.grant_question", data),
			facts:    pathFacts(p, path),
			declined: p.T("auth.grant_declined", data),
			hint:     p.T("auth.grant_again_hint", data),
		}
		if err := question.ask(ctx, p, g, human, c.ErrOrStderr()); err != nil {
			return err
		}
		answer := grantChange{Service: service, Domain: domain, Path: path, Action: "grant"}
		err = changeGrants(ctx, file, func(current []grant) ([]grant, error) {
			if existing := findGrant(current, service, domain); existing != nil {
				answer.Path, answer.Action = existing.Path, "unchanged"
				return nil, nil
			}
			granted := grant{Service: service, Domain: domain, Path: path, GrantedAt: time.Now().UTC().Format(time.RFC3339)}
			return append(current, granted), nil
		})
		if err != nil {
			return grantsFailure(ctx, p, err)
		}
		message := "auth.granted"
		if answer.Action == "unchanged" {
			message = "auth.grant_unchanged"
		}
		return grantAnswer(p, answer, message)
	})
	c.ValidArgsFunction = grantCompletion(extensions, func(service string) []string {
		var domains []string
		for _, e := range extensions {
			if e.broken == nil && slices.Contains(declaredServices([]*extension{e}), service) {
				domains = append(domains, e.domain)
			}
		}
		return domains
	})
	group.AddCommand(c)
}

// addAuthRevoke adds ncly auth revoke, which removes the grant of a
// service to a domain, whatever its source, even when no installed
// extension has the domain anymore.
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

// grantTarget returns the installed extension of a domain that declares a
// service, or the problem that keeps grant from naming it: NOT_FOUND for a
// domain that no installed extension has or a service that its commands
// do not declare, and EXTENSION_INVALID for an extension that cannot run.
func grantTarget(p *i18n.Printer, extensions []*extension, service, domain string) (*extension, *contract.Problem) {
	data := map[string]any{"Service": service, "Domain": domain}
	status := []string{"ncly", "auth", "status"}
	i := slices.IndexFunc(extensions, func(e *extension) bool { return e.domain == domain })
	switch {
	case i < 0:
		return nil, &contract.Problem{Code: contract.NotFound, Message: p.T("auth.no_domain", data), Hint: strings.Join(status, " "), Next: status}
	case extensions[i].broken != nil:
		problem := extensions[i].invalid(p)
		return nil, &problem
	case !slices.Contains(declaredServices(extensions[i:i+1]), service):
		return nil, &contract.Problem{Code: contract.NotFound, Message: p.T("auth.not_declared", data), Hint: strings.Join(status, " "), Next: status}
	}
	return extensions[i], nil
}

// noGrant is the NOT_FOUND of a grant that does not exist.
func noGrant(p *i18n.Printer, data map[string]any) contract.Problem {
	status := []string{"ncly", "auth", "status"}
	return contract.Problem{Code: contract.NotFound, Message: p.T("auth.revoke_none", data), Hint: strings.Join(status, " "), Next: status}
}

// pathFacts shows the source of a grant as a fact of its confirmation. A
// bundled extension has no path to show.
func pathFacts(p *i18n.Printer, path *string) []tui.Fact {
	if path == nil {
		return nil
	}
	return []tui.Fact{{Label: p.T("auth.fact_path"), Value: *path}}
}

// grantsFailure answers a failure to read or change the grants file:
// CONFIG_INVALID for a file that core cannot read, TEMPORARY for a lock
// that another process holds, the signal that ended the wait for the lock,
// and RUNTIME for a write that failed, whose effect may be unknown.
func grantsFailure(ctx context.Context, p *i18n.Printer, err error) error {
	var (
		invalid *grantsError
		failed  *grantsWriteError
		failure *command.Failure
	)
	switch {
	case errors.As(err, &invalid):
		return command.Fail(invalid.problem(p))
	case errors.Is(err, errGrantsBusy):
		return command.Fail(contract.Problem{Code: contract.Temporary, Message: p.T("auth.grants_busy"), Hint: p.T("auth.grants_busy_hint")})
	case errors.As(err, &failed):
		data := map[string]any{"File": failed.file}
		return command.Fail(contract.Problem{Code: contract.Runtime, Message: p.T("auth.grants_unwritten", data), Hint: p.T("auth.grants_unwritten_hint", data)})
	case errors.As(err, &failure):
		return failure
	case ctx.Err() != nil:
		return stopped(ctx)
	}
	return err
}

// grantAnswer answers a grant or a revocation, with the text of message.
func grantAnswer(p *i18n.Printer, a grantChange, message string) *answered {
	data := map[string]any{"Service": a.Service, "Domain": a.Domain}
	return &answered{
		outcome: contract.Outcome{Data: a},
		text: func(w io.Writer) error {
			_, err := io.WriteString(w, p.T(message, data)+"\n")
			return err
		},
	}
}

// plannedGrantChange answers a dry run of grant or revoke, whose text is
// summary, then one line per pending check.
func plannedGrantChange(p *i18n.Printer, plan []plannedGrant, pending []pendingItem, summary string) *answered {
	a := grantDryRun{DryRun: true, Plan: plan, PendingChecks: pending}
	if a.Plan == nil {
		a.Plan = []plannedGrant{}
	}
	if a.PendingChecks == nil {
		a.PendingChecks = []pendingItem{}
	}
	return &answered{
		outcome: contract.Outcome{Data: a},
		text: func(w io.Writer) error {
			text := summary + "\n"
			for _, c := range a.PendingChecks {
				text += p.T("auth.pending", map[string]any{"ID": c.ID, "Message": c.Message}) + "\n"
			}
			_, err := io.WriteString(w, text)
			return err
		},
	}
}

// grantCompletion completes the service of grant and revoke with the
// services that the manifests declare, as login does, then the domain with
// the domains that domains gives for that service, without asking the
// keychain or starting an extension.
func grantCompletion(extensions []*extension, domains func(service string) []string) cobra.CompletionFunc {
	services := serviceCompletion(extensions)
	return func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return services(c, args, toComplete)
		case 1:
			var suggested []cobra.Completion
			for _, d := range domains(args[0]) {
				if strings.HasPrefix(d, toComplete) {
					suggested = append(suggested, d)
				}
			}
			return suggested, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
