package cli

import (
	"context"
	"time"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/runner"
)

// auth checks the keychain, then each service that the manifests declare,
// in the order of auth status and with the verdict that status gives it,
// as Required keys in core-spec.md defines. Doctor runs it last, and each
// keychain call gets the smaller of the keychain's limit and the time left
// of --timeout. Grants, whose check sits between the two, arrive in M08.
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
	for _, s := range declaredServices(extensions) {
		id := "auth." + s
		if pr.spent || time.Until(pr.deadline) <= 0 {
			checks = append(checks, notStarted(id, s))
			continue
		}
		entry, err := serviceVerdict(ctx, pr.p, k, s)
		switch {
		case pr.ctx.Err() != nil:
			return stop()
		case err != nil:
			checks = append(checks, timedOut(id, pr.p.T("doctor.key_timeout", map[string]any{"Service": s})))
		case entry.Code == "":
			checks = append(checks, check{ID: id, Component: "auth", Status: statusPass, Message: entry.found})
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
