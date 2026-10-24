package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// status runs an extension's declared status command after its program probes.
// A rejected answer becomes one failed <slug>.status check, as Status command in contract.md defines.
func (pr *prober) status(e *extension) []check {
	i := slices.IndexFunc(e.manifest.Commands, func(d contract.Command) bool { return d.Doctor })
	if e.broken != nil || i < 0 || pr.signal != nil {
		return nil
	}
	d := e.manifest.Commands[i]
	data := map[string]any{"Domain": e.slug, "Name": strings.Join(d.Path, " "), "Path": e.path}
	failed := func(code contract.Code, message, hint string, next []string) []check {
		return []check{{ID: e.slug + ".status", Component: e.slug, Status: statusFail, Code: code, Message: message, Hint: hint, Next: next}}
	}
	timedOut := func(message string) []check {
		pr.spent = true
		return failed(contract.CheckTimeout, message, pr.p.T("doctor.timeout_hint"), nil)
	}
	invalid := func(reason string) []check {
		data["Reason"] = reason
		return failed(contract.ExtensionInvalid, pr.p.T("doctor.status_invalid", data), pr.p.T("extension.invalid_hint", data), nil)
	}
	if pr.spent || time.Until(pr.deadline) <= 0 {
		return timedOut(pr.p.T("doctor.not_started", data))
	}
	// Under --live, a status command that declares live gets its granted
	// keys, which the keychain reads within the time left of --timeout, as
	// Required keys in core-spec.md defines. Without it, it gets none.
	read := pr.live && slices.ContainsFunc(d.Flags, func(f contract.Flag) bool { return f.Name == "live" })
	bounded, cancel := context.WithDeadline(pr.ctx, pr.deadline)
	keys, err := keyVars(bounded, pr.keys, e, d.Requires.Keys, read)
	cancel()
	switch {
	case err != nil && pr.ctx.Err() != nil:
		// The signal also reaches the program that the keychain call may
		// have started, which the call leaves running.
		pr.signal = runner.SignalOf(pr.ctx)
		runner.Stop(pr.signal)
		return nil
	case err != nil:
		return timedOut(pr.p.T("doctor.status_key_timeout", data))
	}
	left := time.Until(pr.deadline)
	if left <= 0 {
		return timedOut(pr.p.T("doctor.status_key_timeout", data))
	}
	args := slices.Clone(d.Path[len(e.root()):])
	for _, f := range d.Flags {
		switch {
		case f.Name == "live" && pr.live:
			args = append(args, "--live=true")
		case f.Name == "timeout":
			args = append(args, "--timeout="+strconv.FormatInt(int64(math.Ceil(left.Seconds())), 10)+"s")
		}
	}
	var stdout bytes.Buffer
	stderr := &heldLines{dest: io.Discard}
	vars := extensionVars(withJSON(pr.g))
	maps.Copy(vars, keys)
	program := runner.Program{
		Path:   e.path,
		Args:   append(args, "--"),
		Env:    runner.Environ(os.Environ(), vars),
		Stdout: &stdout,
		Stderr: stderr,
		Limit:  left,
	}
	if e.bundled {
		program.Name = executablePrefix + e.slug
	}
	result, err := runner.Run(pr.ctx, program)
	data["Exit"] = result.Exit
	switch {
	case err != nil:
		return invalid(pr.p.T("extension.start_failed", data))
	case result.Signal != nil:
		pr.signal = result.Signal
		return nil
	case result.TimedOut:
		return timedOut(pr.p.T("doctor.status_timeout", data))
	}
	answer, err := contract.CheckAnswer(e.manifest, contract.ResultReport, contract.Reply{Exit: result.Exit, Stdout: stdout.Bytes(), Last: stderr.finish()})
	var rule *contract.RuleError
	switch {
	case errors.As(err, &rule):
		return invalid(pr.p.T(rule.Rule, rule.Data))
	case err != nil:
		return invalid(pr.p.T("extension.answer_json", data))
	case result.Exit != int(contract.ExitOK) && result.Exit != int(contract.ExitHuman):
		return failed(contract.ExtensionInvalid, pr.p.T("doctor.status_failed", data), pr.p.T("extension.invalid_hint", data), nil)
	}
	var reply struct {
		Checks json.RawMessage    `json:"checks"`
		Errors []contract.Problem `json:"errors"`
	}
	if err := json.Unmarshal(answer.Line, &reply); err != nil {
		return invalid(pr.p.T("extension.answer_json", data))
	}
	if len(reply.Errors) > 0 {
		lead := reply.Errors[0]
		if exit, mapped := e.manifest.Registry().Exit(lead.Code); !mapped || exit != contract.ExitHuman {
			return invalid(pr.p.T("doctor.rule_lead", map[string]any{"Code": lead.Code}))
		}
		return failed(lead.Code, lead.Message, lead.Hint, lead.Next)
	}
	var raw []json.RawMessage
	if json.Unmarshal(reply.Checks, &raw) != nil || raw == nil {
		return invalid(pr.p.T("doctor.rule_checks"))
	}
	checks, reason := statusChecks(pr.p, e, raw, result.Exit)
	if reason != "" {
		return invalid(reason)
	}
	return checks
}

// codeForm is the form of every code: uppercase letters, digits, and
// underscores.
var codeForm = regexp.MustCompile(`^[A-Z0-9_]+$`)

// statusChecks validates a complete status report before doctor accepts any check.
// Check ownership, codes, hints, and the exit verdict follow Status command in contract.md.
func statusChecks(p *i18n.Printer, e *extension, raw []json.RawMessage, exit int) ([]check, string) {
	registry := e.manifest.Registry()
	checks := []check{}
	failed := false
	for i, r := range raw {
		var c check
		data := map[string]any{"Index": i + 1, "Domain": e.slug}
		if err := json.Unmarshal(r, &c); err != nil || c.extensionKeys != nil {
			return nil, p.T("doctor.rule_shape", data)
		}
		data["ID"], data["Component"], data["Code"], data["Status"] = c.ID, c.Component, c.Code, c.Status
		code, isError := registry.Exit(c.Code)
		switch {
		case c.Component != e.slug || !strings.HasPrefix(c.ID, e.slug+".") || len(c.ID) == len(e.slug)+1:
			return nil, p.T("doctor.rule_component", data)
		case c.Message == "" || !slices.Contains([]string{statusPass, statusWarn, statusFail}, c.Status):
			return nil, p.T("doctor.rule_shape", data)
		case c.Status != statusPass && !codeForm.MatchString(string(c.Code)),
			c.Status == statusPass && (c.Code != "" || c.Hint != "" || c.Next != nil),
			c.Status == statusFail && (!isError || code != contract.ExitHuman),
			c.Status == statusWarn && !registry.IsWarning(c.Code):
			return nil, p.T("doctor.rule_code", data)
		case c.Status != statusPass && c.Hint == "",
			c.Next != nil && (len(c.Next) == 0 || slices.Contains(c.Next, "")):
			return nil, p.T("doctor.rule_hint", data)
		}
		failed = failed || c.Status == statusFail
		checks = append(checks, c)
	}
	if failed != (exit == int(contract.ExitHuman)) {
		return nil, p.T("doctor.rule_verdict", map[string]any{"Exit": exit})
	}
	return checks, ""
}
