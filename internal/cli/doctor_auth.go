package cli

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// auth checks the keychain, then the grants, then each service that the
// manifests declare, in the order of auth status and with the verdict that
// status gives it, as Required keys in core-spec.md defines. Doctor runs it
// last, and each keychain call gets the smaller of the keychain's limit and
// the time left of --timeout.
func (pr *prober) auth(extensions []*extension, k *keychain.Keychain) []check {
	if pr.signal != nil {
		return nil
	}
	ctx, cancel := context.WithDeadline(pr.ctx, pr.deadline)
	defer cancel()
	timedOut := func(id, message string) check {
		pr.spent = true
		return check{ID: id, Component: "auth", Status: statusFail, Code: contract.CheckTimeout, Message: message, Hint: pr.p.T("doctor.timeout_hint")}
	}
	notStarted := func(id, name string) check {
		return timedOut(id, pr.p.T("doctor.not_started", map[string]any{"Name": name}))
	}
	// stop ends doctor on a signal, which also reaches the program that a
	// keychain call may have started, which the call leaves running.
	stop := func() []check {
		pr.signal = runner.SignalOf(pr.ctx)
		runner.Stop(pr.signal)
		return nil
	}
	checks := []check{{}}
	answered := false
	if pr.spent || time.Until(pr.deadline) <= 0 {
		checks[0] = notStarted("auth.keychain", "keychain")
	} else {
		_, err := k.Available(ctx)
		switch {
		case pr.ctx.Err() != nil:
			return stop()
		case err != nil:
			checks[0] = timedOut("auth.keychain", pr.p.T("doctor.keychain_timeout"))
		default:
			answered = true
		}
	}
	var grants []grant
	readable := false
	if pr.spent || time.Until(pr.deadline) <= 0 {
		checks = append(checks, notStarted("auth.grants", "grants"))
	} else {
		var found []check
		found, grants, readable = grantChecks(pr.p, extensions)
		checks = append(checks, found...)
	}
	for _, s := range declaredServices(extensions) {
		id := "auth." + s
		if pr.spent || time.Until(pr.deadline) <= 0 {
			checks = append(checks, notStarted(id, s))
			continue
		}
		// A grants file that core cannot read leaves the keys to be judged
		// alone.
		entry, err := serviceVerdict(ctx, pr.p, k, s, extensions, grants, readable)
		switch {
		case pr.ctx.Err() != nil:
			return stop()
		case err != nil:
			checks = append(checks, timedOut(id, pr.p.T("doctor.key_timeout", map[string]any{"Service": s})))
		case entry.Code == "":
			checks = append(checks, check{ID: id, Component: "auth", Status: statusPass, Message: entry.found})
		case entry.Code == contract.KeyNotGranted:
			// One entry for each extension that needs the key and lacks a
			// grant, each with the command that grants it.
			for _, domain := range entry.ungranted {
				problem := notGranted(pr.p, s, domain)
				checks = append(checks, check{ID: id, Component: "auth", Status: statusFail, Code: problem.Code, Message: problem.Message, Hint: problem.Hint, Next: problem.Next})
			}
		default:
			checks = append(checks, check{ID: id, Component: "auth", Status: statusFail, Code: entry.Code, Message: entry.Message, Hint: entry.Hint, Next: entry.Next})
		}
	}
	// A read can find the keychain unavailable after it answered the first
	// question, so auth.keychain gives its state after the last read.
	if answered {
		checks[0] = check{ID: "auth.keychain", Component: "auth", Status: statusPass, Message: pr.p.T("auth.keychain_available")}
		if available, _ := k.Available(ctx); !available {
			checks[0] = check{ID: "auth.keychain", Component: "auth", Status: statusWarn, Code: contract.KeyringEnvOnly, Message: pr.p.T("auth.keychain_unavailable"), Hint: pr.p.T("doctor.keyring_env_only_hint")}
		}
	}
	return checks
}

// grantChecks reports GRANT_UNUSED when an extension slug no longer declares the service or no longer exists at its source.
// It reports one pass when every grant is in use, as Required keys in core-spec.md defines.
// A grant of an extension that cannot run counts as in use, because
// doctor cannot read what it declares. A grants file that core cannot read
// fails the check with CONFIG_INVALID, and readable is then false. It also
// returns the grants, which the checks of the services read.
func grantChecks(p *i18n.Printer, extensions []*extension) (checks []check, grants []grant, readable bool) {
	grants, err := readGrants(grantsPath())
	var invalid *grantsError
	if errors.As(err, &invalid) {
		problem := invalid.problem(p)
		return []check{{ID: "auth.grants", Component: "auth", Status: statusFail, Code: problem.Code, Message: problem.Message, Hint: problem.Hint}}, nil, false
	}
	for _, g := range grants {
		data := map[string]any{"Service": g.Service, "Domain": g.Domain, "Path": p.T("auth.fact_bundled")}
		if g.Path != nil {
			data["Path"] = *g.Path
		}
		i := slices.IndexFunc(extensions, func(e *extension) bool { return e.slug == g.Domain })
		var message string
		switch {
		case i >= 0 && extensions[i].broken != nil:
			continue
		case i < 0 || !sameSource(g.Path, sourceOf(extensions[i])):
			message = p.T("auth.grant_unused_source", data)
		case !slices.Contains(declaredServices(extensions[i:i+1]), g.Service):
			message = p.T("auth.grant_unused_service", data)
		default:
			continue
		}
		revoke := []string{"ncly", "auth", "revoke", g.Service, g.Domain}
		checks = append(checks, check{ID: "auth.grants", Component: "auth", Status: statusWarn, Code: contract.GrantUnused, Message: message, Hint: strings.Join(revoke, " "), Next: revoke})
	}
	if len(checks) == 0 {
		checks = append(checks, check{ID: "auth.grants", Component: "auth", Status: statusPass, Message: p.T("auth.grants_in_use")})
	}
	return checks, grants, true
}
