package cli

import (
	"context"
	_ "embed"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

//go:embed auth_status_schema.json
var authStatusSchema []byte

var authCommand = contract.Command{
	Path:    []string{"auth"},
	Summary: "auth.summary",
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly auth status`,
		`ncly auth login deepgram`,
		`<command that prints the key> | ncly auth login deepgram --stdin`,
		`ncly auth grant deepgram transcript`,
	},
}

var authStatusCommand = contract.Command{
	Path:    []string{"auth", "status"},
	Result:  contract.Result{Kind: contract.ResultReport, Schema: mustSchema(authStatusSchema)},
	Summary: "auth.status.summary",
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly auth status`,
		`ncly auth status --json`,
	},
}

// Sources of a key, as ncly auth in core-spec.md defines them.
const (
	sourceEnv      = "env"
	sourceKeychain = "keychain"
	sourceNone     = "none"
)

// statusReport is the answer of ncly auth status.
type statusReport struct {
	// Keychain is available or unavailable.
	Keychain string         `json:"keychain"`
	Services []serviceEntry `json:"services"`
}

// serviceEntry is the verdict of one service: where its key comes from, the
// grants of the extensions that declare it, and the keys of a problem when
// an extension that needs the key lacks it or its grant.
type serviceEntry struct {
	Service    string        `json:"service"`
	Configured bool          `json:"configured"`
	Source     string        `json:"source"`
	Code       contract.Code `json:"code,omitempty"`
	Message    string        `json:"message,omitempty"`
	Hint       string        `json:"hint,omitempty"`
	Next       []string      `json:"next,omitempty"`
	Grants     []grantEntry  `json:"grants"`
	// found is the message of an entry whose key was found, for the check of
	// its service in ncly doctor.
	found string
	// ungranted lists, in name order, the domains whose extension needs the
	// key and lacks a grant, for the checks of ncly doctor.
	ungranted []string
}

// grantEntry says whether the extension of a domain that declares a
// service holds a grant of its key, as ncly auth in core-spec.md defines.
type grantEntry struct {
	Domain   string `json:"domain"`
	Granted  bool   `json:"granted"`
	Optional bool   `json:"optional"`
}

// addAuth adds ncly auth: status, which reports the keys that the installed
// extensions declare, login, logout, grant, and revoke. b is the keychain's
// backend.
func addAuth(p *i18n.Printer, g command.Globals, root *cobra.Command, extensions []*extension, b keychain.Backend) {
	group := command.Build(p, p, authCommand, nil)
	group.AddCommand(command.Build(p, p, authStatusCommand, func(c *cobra.Command, _ []string) error {
		ctx := runner.OnSignal()
		grants, err := readGrants(grantsPath())
		if err != nil {
			return grantsFailure(ctx, p, err)
		}
		k := keychain.New(b, interactive(g, c.OutOrStdout()))
		r, err := authStatus(ctx, p, k, extensions, grants)
		if err != nil {
			return stopped(ctx)
		}
		failed := slices.ContainsFunc(r.Services, func(s serviceEntry) bool { return s.Code != "" })
		return &answered{
			outcome: contract.Outcome{Data: r, HumanNeeded: failed},
			text:    func(w io.Writer) error { return writeServices(w, p, r) },
		}
	}))
	addAuthLogin(p, g, group, extensions, b)
	addAuthLogout(p, g, group, extensions, b)
	addAuthGrant(p, g, group, extensions)
	addAuthRevoke(p, g, group, extensions)
	root.AddCommand(group)
}

// interactive reports whether the command runs in interactive mode, as
// Modes in contract.md defines it.
func interactive(g command.Globals, stdout io.Writer) bool {
	out, isFile := stdout.(interface{ Fd() uintptr })
	return isFile && contract.DetectMode(term.IsTerminal(int(os.Stdin.Fd())), term.IsTerminal(int(out.Fd())), g.NoInput, g.JSON, os.LookupEnv) == contract.Interactive
}

// declaredServices lists, once each and in name order, the services that
// the commands of the valid extensions declare in requires.keys.
func declaredServices(extensions []*extension) []string {
	var services []string
	for _, e := range extensions {
		if e.broken != nil {
			continue
		}
		for _, c := range e.manifest.Commands {
			for _, s := range c.Requires.Keys {
				if !slices.Contains(services, s) {
					services = append(services, s)
				}
			}
		}
	}
	slices.Sort(services)
	return services
}

// authStatus asks the keychain once whether it answers, then gives each
// service that the extensions declare its verdict, with the grants that the
// grants file holds. It returns the error of ctx when a signal ends it.
func authStatus(ctx context.Context, p *i18n.Printer, k *keychain.Keychain, extensions []*extension, grants []grant) (statusReport, error) {
	r := statusReport{Keychain: "available", Services: []serviceEntry{}}
	if _, err := k.Available(ctx); err != nil {
		return r, err
	}
	for _, s := range declaredServices(extensions) {
		entry, err := serviceVerdict(ctx, p, k, s, extensions, grants, true)
		if err != nil {
			return r, err
		}
		r.Services = append(r.Services, entry)
	}
	// A read can find the keychain unavailable after it answered the first
	// question, so the report gives its state after the last read.
	if available, _ := k.Available(ctx); !available {
		r.Keychain = "unavailable"
	}
	return r, nil
}

// serviceVerdict gives one service its verdict, which ncly auth status and
// the auth component of doctor share: the key from its variable first, then
// from the keychain, which it reads only when the variable is empty, then,
// with judgeGrants, the grant of each extension that needs the key. A
// missing key reports its own code alone, because the key comes before its
// grants. It returns the error of ctx when ctx ends before the keychain
// answers.
func serviceVerdict(ctx context.Context, p *i18n.Printer, k *keychain.Keychain, service string, extensions []*extension, grants []grant, judgeGrants bool) (serviceEntry, error) {
	data := map[string]any{"Service": service, "Variable": keyVariable(service)}
	entry := serviceEntry{Service: service, Source: sourceNone, Grants: serviceGrants(service, extensions, grants)}
	_, source, code, err := lookupKey(ctx, k, service)
	if err != nil {
		return entry, err
	}
	entry.Source = source
	switch code {
	case contract.KeyringUnavailable:
		entry.Code, entry.Message, entry.Hint = code, p.T("auth.keyring_unavailable", data), p.T("auth.keyring_unavailable_hint", data)
	case contract.AuthMissing:
		login := []string{"ncly", "auth", "login", service}
		entry.Code, entry.Message, entry.Hint, entry.Next = code, p.T("auth.missing", data), strings.Join(login, " "), login
	case "":
		entry.Configured, entry.found = true, p.T("auth.from_"+source, data)
	}
	if code != "" || !judgeGrants {
		return entry, nil
	}
	for _, g := range entry.Grants {
		if !g.Granted && !g.Optional {
			entry.ungranted = append(entry.ungranted, g.Domain)
		}
	}
	if len(entry.ungranted) > 0 {
		problem := notGranted(p, service, entry.ungranted[0])
		entry.Code, entry.Message, entry.Hint, entry.Next = problem.Code, problem.Message, problem.Hint, problem.Next
	}
	return entry, nil
}

// notGranted is the KEY_NOT_GRANTED problem of a key that the extension of
// a domain needs and that no human granted to it, whose hint and next
// grant it.
func notGranted(p *i18n.Printer, service, domain string) contract.Problem {
	grant := []string{"ncly", "auth", "grant", service, domain}
	data := map[string]any{"Service": service, "Domain": domain}
	return contract.Problem{Code: contract.KeyNotGranted, Message: p.T("auth.not_granted", data), Hint: strings.Join(grant, " "), Next: grant}
}

// serviceGrants lists, in name order, the grant of each valid extension
// that declares a service: whether a human granted the key to that
// extension at its source.
func serviceGrants(service string, extensions []*extension, grants []grant) []grantEntry {
	entries := []grantEntry{}
	for _, e := range extensions {
		if e.broken != nil || !slices.Contains(declaredServices([]*extension{e}), service) {
			continue
		}
		g := findGrant(grants, service, e.domain)
		entries = append(entries, grantEntry{Domain: e.domain, Granted: g != nil && sameSource(g.Path, sourceOf(e))})
	}
	slices.SortFunc(entries, func(a, b grantEntry) int { return strings.Compare(a.Domain, b.Domain) })
	return entries
}

// keyVariable is the environment variable of a service's key, such as
// DEEPGRAM_API_KEY, as Keys in contract.md names it.
func keyVariable(service string) string { return strings.ToUpper(service) + "_API_KEY" }

// writeServices prints the report as text: the keychain and its state, an
// empty line, then one line per service, in aligned columns, its name in
// bold, its source, then its grants, separated by commas. A service's
// problem follows its line, in the column of its source, with the label
// styles of an error.
func writeServices(w io.Writer, p *i18n.Printer, r statusReport) error {
	width := ansi.StringWidth(sourceKeychain)
	for _, s := range r.Services {
		width = max(width, ansi.StringWidth(s.Service))
	}
	state := p.T("auth.state_available")
	if r.Keychain != "available" {
		state = p.T("auth.state_unavailable")
	}
	sources := map[string]string{
		sourceEnv:      p.T("auth.source_env"),
		sourceKeychain: p.T("auth.source_keychain"),
		sourceNone:     p.T("auth.source_none"),
	}
	sourceWidth := 0
	for _, s := range r.Services {
		sourceWidth = max(sourceWidth, ansi.StringWidth(sources[s.Source]))
	}
	indent := strings.Repeat(" ", width+2)
	var b strings.Builder
	b.WriteString(sourceKeychain + strings.Repeat(" ", width-ansi.StringWidth(sourceKeychain)+2) + state + "\n")
	if len(r.Services) > 0 {
		b.WriteString("\n")
	}
	for _, s := range r.Services {
		source := sources[s.Source]
		b.WriteString(tui.Bold(s.Service) + strings.Repeat(" ", width-ansi.StringWidth(s.Service)+2) + source)
		if len(s.Grants) > 0 {
			b.WriteString(strings.Repeat(" ", sourceWidth-ansi.StringWidth(source)+2) + grantsText(p, s.Grants))
		}
		b.WriteString("\n")
		if s.Code != "" {
			b.WriteString(indent + tui.ErrorLabel(p.T("auth.fail_label", map[string]any{"Code": s.Code})) + " " + s.Message + "\n")
			b.WriteString(indent + tui.HintLabel(p.T("error.hint")) + " " + s.Hint + "\n")
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// grantsText writes the grants of a service as auth status shows them:
// each one as <domain>: granted or <domain>: not granted, with , optional
// for an optional key, separated by commas.
func grantsText(p *i18n.Printer, grants []grantEntry) string {
	texts := make([]string, 0, len(grants))
	for _, g := range grants {
		id := "auth.grant_not_granted"
		if g.Granted {
			id = "auth.grant_granted"
		}
		if g.Optional {
			id += "_optional"
		}
		texts = append(texts, p.T(id, map[string]any{"Domain": g.Domain}))
	}
	return strings.Join(texts, ", ")
}

// serviceCompletion completes the service argument of login and logout with
// the services that the manifests that discovery finds declare, as status
// lists them, without asking the keychain or starting an extension.
func serviceCompletion(extensions []*extension) cobra.CompletionFunc {
	return func(_ *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var services []cobra.Completion
		for _, s := range declaredServices(extensions) {
			if strings.HasPrefix(s, toComplete) {
				services = append(services, s)
			}
		}
		return services, cobra.ShellCompDirectiveNoFileComp
	}
}
