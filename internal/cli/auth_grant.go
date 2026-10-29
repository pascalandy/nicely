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
	"github.com/spf13/cobra"
)

//go:embed auth_grant_schema.json
var authGrantSchema []byte

//go:embed auth_grant_dry_run_schema.json
var authGrantDryRunSchema []byte

// grantArgs are the arguments of grant and revoke: the service, then the
// slug.
var grantArgs = []contract.Arg{
	{Name: "service", Summary: "auth.service", Required: true},
	{Name: "slug", Summary: "auth.slug", Required: true},
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

// addAuthGrant adds ncly auth grant for an installed extension's slug.
// Grants in contract.md requires human confirmation in a terminal.
// It reads no key.
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
		path := sourceOf(e)
		existing := findGrant(grants, service, domain)
		if existing != nil && sameSource(existing.Path, path) {
			if dryRun {
				return plannedGrantChange(p, nil, nil, p.T("auth.planned_nothing", data))
			}
			return grantAnswer(p, grantChange{Service: service, Domain: domain, Path: path, Action: "unchanged"}, "auth.grant_unchanged")
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
		ctx := runner.OnSignal()
		facts := pathFacts(p, path)
		if existing != nil {
			// The grant moves from another source, so the human sees both.
			facts = replacedFacts(p, path, existing.Path)
		}
		question := confirmation{
			question: p.T("auth.grant_question", data),
			facts:    facts,
			declined: p.T("auth.grant_declined", data),
			hint:     p.T("auth.grant_again_hint", data),
		}
		if err := question.ask(ctx, p, g, human, c.ErrOrStderr()); err != nil {
			return err
		}
		answer := grantChange{Service: service, Domain: domain, Path: path, Action: "grant"}
		err = changeGrants(ctx, file, func(current []grant) ([]grant, error) {
			if found := findGrant(current, service, domain); found != nil && sameSource(found.Path, path) {
				answer.Action = "unchanged"
				return nil, nil
			}
			// A grant at another source gives way to this one, which the
			// human confirmed.
			current = slices.DeleteFunc(current, func(g grant) bool { return g.Service == service && g.Domain == domain })
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
				domains = append(domains, e.slug)
			}
		}
		return domains
	})
	group.AddCommand(c)
}

// grantTarget selects an installed slug that declares the service.
// An unknown slug or undeclared service gets NOT_FOUND.
// An extension that cannot run gets EXTENSION_INVALID.
func grantTarget(p *i18n.Printer, extensions []*extension, service, domain string) (*extension, *contract.Problem) {
	data := map[string]any{"Service": service, "Domain": domain}
	status := []string{"ncly", "auth", "status"}
	i := slices.IndexFunc(extensions, func(e *extension) bool { return e.slug == domain })
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

// grantCompletion suggests declared services, then the slugs that domains gives for each service.
// It never asks the keychain or starts an extension.
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
