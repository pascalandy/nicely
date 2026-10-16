package cli

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

//go:embed auth_login_schema.json
var authLoginSchema []byte

//go:embed auth_dry_run_schema.json
var authDryRunSchema []byte

var authLoginCommand = contract.Command{
	Path:    []string{"auth", "login"},
	Summary: "auth.login.summary",
	Args:    []contract.Arg{{Name: "service", Summary: "auth.service", Required: true}},
	Flags: []contract.Flag{
		{Name: "stdin", Summary: "auth.login.stdin"},
		forceFlag,
	},
	Result:  contract.Result{Kind: contract.ResultObject, Schema: mustSchema(authLoginSchema), DryRunSchema: mustSchema(authDryRunSchema)},
	Effects: &contract.Effects{UserWrites: true},
	Modes:   contract.Modes{DryRun: true},
	Examples: []string{
		`ncly auth login deepgram`,
		`ncly auth login deepgram --dry-run --json`,
		`<command that prints the key> | ncly auth login deepgram --stdin`,
	},
}

// forceFlag answers yes to the confirmation of a replacement.
var forceFlag = contract.Flag{Name: "force", Shorthand: "f", Summary: "auth.force"}

// changed is the answer of a login or a logout that succeeded: never a key.
type changed struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

// dryRunAnswer is the answer of ncly auth login or logout under --dry-run,
// as ncly auth in core-spec.md defines it.
type dryRunAnswer struct {
	DryRun        bool          `json:"dry_run"`
	Plan          []plannedKey  `json:"plan"`
	PendingChecks []pendingItem `json:"pending_checks"`
}

type plannedKey struct {
	Key     string         `json:"key"`
	Action  string         `json:"action"`
	Effects plannedEffects `json:"effects"`
}

// plannedEffects are the effects of a planned change to the keychain.
type plannedEffects struct {
	UserWrites bool `json:"user_writes"`
	Network    bool `json:"network"`
	Paid       bool `json:"paid"`
}

type pendingItem struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// addAuthLogin adds ncly auth login, which stores the key of a service. It
// asks the keychain before it reads a key or asks for a confirmation, so a
// human never types a key that cannot be stored.
func addAuthLogin(p *i18n.Printer, g command.Globals, group *cobra.Command, extensions []*extension, b keychain.Backend) {
	group.AddCommand(command.Build(p, p, authLoginCommand, func(c *cobra.Command, args []string) error {
		service := args[0]
		if problem := badService(p, service); problem != nil {
			return command.Fail(*problem)
		}
		if dryRun, _ := c.Flags().GetBool(contract.DryRunFlag.Name); dryRun {
			return plannedChange(p, service, "login")
		}
		fromStdin, _ := c.Flags().GetBool("stdin")
		force, _ := c.Flags().GetBool(forceFlag.Name)
		out := c.OutOrStdout()
		human := interactive(g, out)
		call := []string{"ncly", "auth", "login", service}
		if fromStdin {
			call = append(call, "--stdin")
		}
		data := map[string]any{"Service": service, "Variable": keyVariable(service), "Command": strings.Join(call[:4], " ")}
		switch {
		case fromStdin && term.IsTerminal(int(os.Stdin.Fd())):
			return command.Fail(command.Usage(p.T("auth.stdin_terminal", data), strings.Join(call[:4], " ")))
		case !fromStdin && !human:
			return command.Fail(contract.Problem{Code: contract.TerminalRequired, Message: p.T("auth.terminal_required", data), Hint: p.T("auth.terminal_required_hint", data)})
		}
		ctx := runner.OnSignal()
		k := keychain.New(b, human)
		_, stored, err := k.Read(ctx, service)
		switch {
		case errors.Is(err, keychain.ErrUnavailable):
			return command.Fail(contract.Problem{Code: contract.KeyringUnavailable, Message: p.T("auth.login_unavailable", data), Hint: p.T("auth.login_unavailable_hint", data)})
		case err != nil:
			return stopped(ctx)
		}
		if stored && !force {
			hint := strings.Join(append(call, "--force"), " ")
			if !human {
				return command.Fail(contract.Problem{Code: contract.ConfirmationRequired, Message: p.T("auth.replace_required", data), Hint: hint})
			}
			yes, err := confirm(ctx, os.Stdin, out, p.T("auth.replace_question", data), yesWords(p.T("auth.yes")))
			switch {
			case ctx.Err() != nil:
				return stopped(ctx)
			case err != nil:
				return err
			case !yes:
				return command.Fail(contract.Problem{Code: contract.ConfirmationRequired, Message: p.T("auth.replace_declined", data), Hint: hint})
			}
		}
		var key string
		if fromStdin {
			key, err = readAll(ctx, os.Stdin)
		} else {
			key, err = readSecret(ctx, os.Stdin, out, p.T("auth.key_prompt", data))
		}
		switch {
		case ctx.Err() != nil:
			return stopped(ctx)
		case err != nil && !errors.Is(err, io.EOF):
			return err
		}
		if key = strings.TrimSpace(key); key == "" {
			hint := p.T("auth.empty_key_hint", map[string]any{"Command": strings.Join(call, " ")})
			return command.Fail(contract.Problem{Code: contract.UsageInvalid, Message: p.T("auth.empty_key", data), Hint: hint})
		}
		if err := k.Store(ctx, service, key); err != nil {
			unknown := writeUnknown(p, service, declaredServices(extensions), "auth.store_unknown")
			if errors.Is(err, keychain.ErrWriteUnknown) {
				return command.Fail(unknown)
			}
			return &answered{outcome: contract.Outcome{Signal: contract.StopSignal(runner.SignalOf(ctx)), Errors: []contract.Problem{unknown}}, text: func(io.Writer) error { return nil }}
		}
		return &answered{
			outcome: contract.Outcome{Data: changed{Service: service, Action: "login"}},
			text: func(w io.Writer) error {
				_, err := io.WriteString(w, p.T("auth.stored", data)+"\n")
				return err
			},
		}
	}))
}

// badService refuses a name that is no service name, keychain included,
// before any other step of ncly auth.
func badService(p *i18n.Printer, service string) *contract.Problem {
	if contract.IsService(service) {
		return nil
	}
	problem := contract.Problem{
		Code:    contract.UsageInvalid,
		Message: p.T("auth.invalid_service", map[string]any{"Service": service}),
		Hint:    "ncly auth status",
		Next:    []string{"ncly", "auth", "status"},
	}
	return &problem
}

// plannedChange answers a dry run of login or logout. It reads no key, asks
// the keychain nothing, and prompts for nothing, so whether the keychain
// answers and holds a key stays a pending check.
func plannedChange(p *i18n.Printer, service, action string) *answered {
	data := map[string]any{"Service": service}
	a := dryRunAnswer{
		DryRun:        true,
		Plan:          []plannedKey{{Key: service, Action: action, Effects: plannedEffects{UserWrites: true}}},
		PendingChecks: []pendingItem{{ID: "auth.keychain", Message: p.T("auth.pending_"+action, data)}},
	}
	return &answered{
		outcome: contract.Outcome{Data: a},
		text: func(w io.Writer) error {
			_, err := io.WriteString(w, p.T("auth.planned_"+action, data)+"\n"+p.T("auth.pending", map[string]any{"ID": a.PendingChecks[0].ID, "Message": a.PendingChecks[0].Message})+"\n")
			return err
		},
	}
}

// writeUnknown is the RUNTIME of a write whose effect is unknown. For a
// service that an installed extension declares, ncly auth status shows the
// key's state. For any other, status lists nothing, so the hint sends the
// human to Nicely's entry in the OS keychain.
func writeUnknown(p *i18n.Printer, service string, declared []string, message string) contract.Problem {
	data := map[string]any{"Service": service, "Keychain": keychain.Service}
	problem := contract.Problem{Code: contract.Runtime, Message: p.T(message, data)}
	if slices.Contains(declared, service) {
		problem.Hint, problem.Next = p.T("auth.write_unknown_hint", data), []string{"ncly", "auth", "status"}
	} else {
		problem.Hint = p.T("auth.write_unknown_entry_hint", data)
	}
	return problem
}

// stopped answers a command that a signal ended. The signal also reaches
// the program that a keychain call may have started, which the call leaves
// running.
func stopped(ctx context.Context) *answered {
	signal := runner.SignalOf(ctx)
	runner.Stop(signal)
	return &answered{outcome: contract.Outcome{Signal: contract.StopSignal(signal)}, text: func(io.Writer) error { return nil }}
}
