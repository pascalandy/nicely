package cli

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:embed doctor_schema.json
var doctorSchema []byte

var doctorCommand = contract.Command{
	Path:    []string{"doctor"},
	Result:  contract.Result{Kind: contract.ResultReport, Schema: mustSchema(doctorSchema)},
	Summary: "doctor.summary",
	Args:    []contract.Arg{{Name: "component", Summary: "doctor.component"}},
	Flags: []contract.Flag{
		{Name: "live", Summary: "doctor.live"},
		{Name: "timeout", Value: "duration", Summary: "doctor.timeout", Default: "30s"},
	},
	// --live reaches the network through the status commands that use it.
	Effects: &contract.Effects{Network: true},
	Examples: []string{
		`ncly doctor`,
		`ncly doctor core --json`,
		`ncly doctor transcript --timeout 1m`,
	},
}

// Statuses of a check, as ncly doctor in core-spec.md defines them.
const (
	statusPass = "pass"
	statusWarn = "warn"
	statusFail = "fail"
)

// check is one finding of doctor, as the check object of ncly doctor in
// core-spec.md defines it.
type check struct {
	ID        string        `json:"id"`
	Component string        `json:"component"`
	Status    string        `json:"status"`
	Code      contract.Code `json:"code,omitempty"`
	Message   string        `json:"message"`
	Hint      string        `json:"hint,omitempty"`
	Next      []string      `json:"next,omitempty"`
	// extensionKeys is set in an entry of the extensions component alone.
	*extensionKeys
}

// extensionKeys tell which extension an entry of the extensions component
// describes, and what hides it.
type extensionKeys struct {
	Domain string `json:"domain"`
	Source string `json:"source"`
	// Path is the executable as discovery found it, or nil for a bundled
	// extension.
	Path     *string   `json:"path"`
	HiddenBy *hiddenBy `json:"hidden_by,omitempty"`
}

// hiddenBy is what wins over a hidden extension: a core command, a bundled
// extension, or an earlier extension, with its path.
type hiddenBy struct {
	Source string  `json:"source"`
	Path   *string `json:"path"`
}

// report is the answer of ncly doctor.
type report struct {
	Checks []check `json:"checks"`
}

// component is a part of the machine that doctor checks, under the name
// that the component argument gives it. extension is set for the
// component of an installed extension. last marks the component that runs
// after every other, auth, whose questions to the keychain never cost a
// check that already finished. It changes when the component runs, not
// where the report lists it.
type component struct {
	name      string
	checks    func() []check
	extension *extension
	last      bool
}

// addDoctor adds ncly doctor, which reports its checks on stdout even when
// one fails, as a report command does. It runs the version probes, then
// each status command, one at a time, because the cleanup of the runner
// reaches every descendant of ncly, and starts no other program. Its auth
// component asks the keychain through b last.
func addDoctor(p *i18n.Printer, g command.Globals, root *cobra.Command, extensions []*extension, b keychain.Backend) {
	root.AddCommand(command.Build(p, p, doctorCommand, func(c *cobra.Command, args []string) error {
		timeout, problem := parseTimeout(p, c.Flag("timeout"))
		if problem != nil {
			return command.Fail(*problem)
		}
		live, _ := c.Flags().GetBool("live")
		probes := &prober{p: p, g: g, live: live, ctx: runner.OnSignal(), deadline: time.Now().Add(timeout)}
		k := keychain.New(b, interactive(g, c.OutOrStdout()))
		components := []component{
			{name: "core", checks: func() []check { return coreChecks(p, extensions) }},
			{name: "auth", checks: func() []check { return probes.auth(extensions, k) }, last: true},
			{name: "extensions", checks: func() []check { return extensionChecks(p, extensions) }},
		}
		for _, e := range extensions {
			components = append(components, component{name: e.domain, checks: func() []check {
				return append(probes.programs(e), probes.status(e)...)
			}, extension: e})
		}
		if len(args) > 0 {
			i := slices.IndexFunc(components, func(c component) bool { return c.name == args[0] })
			switch {
			case i < 0:
				return command.Fail(noComponent(p, args[0], components))
			case components[i].extension != nil && components[i].extension.broken != nil:
				return command.Fail(components[i].extension.invalid(p))
			}
			components = components[i : i+1]
		}
		r := report{Checks: []check{}}
		// Invalid config fails the report even for another component, whose
		// independent checks still run.
		if components[0].name != "core" {
			if config := coreChecks(p, extensions); config[0].Status == statusFail {
				r.Checks = append(r.Checks, config[0])
			}
		}
		found := make([][]check, len(components))
		var order []int
		for _, last := range []bool{false, true} {
			for i, c := range components {
				if c.last == last {
					order = append(order, i)
				}
			}
		}
		for _, i := range order {
			found[i] = components[i].checks()
			// A signal that arrives while no probe runs ends doctor too.
			if probes.signal == nil && probes.ctx.Err() != nil {
				probes.signal = runner.SignalOf(probes.ctx)
			}
			if probes.signal != nil {
				return &answered{outcome: contract.Outcome{Signal: contract.StopSignal(probes.signal)}, text: func(io.Writer) error { return nil }}
			}
		}
		for _, checks := range found {
			r.Checks = append(r.Checks, checks...)
		}
		failed := slices.ContainsFunc(r.Checks, func(c check) bool { return c.Status == statusFail })
		return &answered{
			outcome: contract.Outcome{Data: r, HumanNeeded: failed},
			text:    func(w io.Writer) error { return writeChecks(w, p, r.Checks) },
		}
	}))
}

// noComponent reports a component argument that names no component, with
// the names of those that exist.
func noComponent(p *i18n.Printer, name string, components []component) contract.Problem {
	var names []string
	for _, c := range components {
		names = append(names, c.name)
	}
	return contract.Problem{
		Code:    contract.NotFound,
		Message: p.T("doctor.no_component", map[string]any{"Component": name, "Components": strings.Join(names, ", ")}),
		Hint:    "ncly doctor",
		Next:    []string{"ncly", "doctor"},
	}
}

// coreChecks checks the config files: one failed core.config check when
// one cannot be read, otherwise one warning per unknown key, or one pass
// when they hold none.
func coreChecks(p *i18n.Printer, extensions []*extension) []check {
	finding := func(status string, code contract.Code, message, hint string) check {
		return check{ID: "core.config", Component: "core", Status: status, Code: code, Message: message, Hint: hint}
	}
	_, keys, err := config.Load(config.Locate(home(), os.LookupEnv))
	if err != nil {
		file, line := "", 0
		var invalid *config.InvalidError
		if errors.As(err, &invalid) {
			file, line = invalid.File, invalid.Line
		}
		data := map[string]any{"File": file, "Line": line}
		message := p.T("doctor.config_unreadable", data)
		if line > 0 {
			message = p.T("doctor.config_invalid_line", data)
		}
		return []check{finding(statusFail, contract.ConfigInvalid, message, p.T("doctor.config_invalid_hint", data))}
	}
	var checks []check
	for _, k := range unknownKeys(keys, extensions) {
		data := map[string]any{"File": k.File, "Key": k.Name}
		checks = append(checks, finding(statusWarn, contract.ConfigUnknownKey, p.T("doctor.unknown_key", data), p.T("doctor.unknown_key_hint", data)))
	}
	if len(checks) == 0 {
		checks = append(checks, finding(statusPass, "", p.T("doctor.config_valid"), ""))
	}
	return checks
}

// unknownKeys keeps the keys that doctor warns about, as Configuration in
// contract.md defines them: a top-level key that core does not know, and a
// key of an installed extension's table that its manifest does not declare.
// It skips a table that could belong to an extension not installed here,
// with its keys, and a key beneath one that it keeps, which the warning
// about that one covers.
func unknownKeys(keys []config.Key, extensions []*extension) []config.Key {
	var kept []config.Key
	for _, k := range keys {
		if slices.ContainsFunc(kept, func(o config.Key) bool { return o.File == k.File && strings.HasPrefix(k.Name, o.Name+".") }) {
			continue
		}
		domain, rest, _ := strings.Cut(k.Name, ".")
		i := slices.IndexFunc(extensions, func(e *extension) bool { return e.domain == domain })
		switch {
		case i >= 0 && rest == "" && k.Table:
			continue
		case i >= 0 && rest != "":
			// A broken manifest declares nothing that doctor can read, so its
			// keys stay its own.
			if e := extensions[i]; e.broken != nil || declares(e.manifest, k.Name) {
				continue
			}
		case i < 0 && (rest != "" || k.Table) && contract.IsDomain(domain) && !slices.Contains(reservedNames, domain):
			continue
		}
		kept = append(kept, k)
	}
	return kept
}

// declares reports whether a manifest declares key, a key beneath a
// declared one, or a table above one.
func declares(m contract.Manifest, key string) bool {
	for _, d := range m.Commands {
		for _, declared := range d.Config {
			if key == declared || strings.HasPrefix(key, declared+".") || strings.HasPrefix(declared, key+".") {
				return true
			}
		}
	}
	return false
}

// extensionChecks describes each extension that discovery finds, in the
// order of their domains, without starting any: the one that runs, then
// each one that it hides, in discovery order, and each one that a name
// reserved by core or a bundled extension hides. extensions holds the
// bundled ones and the first one found for each other domain.
func extensionChecks(p *i18n.Printer, extensions []*extension) []check {
	winners := map[string]*extension{}
	for _, e := range extensions {
		winners[e.domain] = e
	}
	found := map[string][]string{}
	seen := map[string]bool{}
	for _, dir := range searchPath() {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			domain, ok := strings.CutPrefix(entry.Name(), executablePrefix)
			path := filepath.Join(dir, entry.Name())
			if ok && contract.IsDomain(domain) && !seen[path] && isExecutable(path) {
				seen[path] = true
				found[domain] = append(found[domain], path)
			}
		}
	}
	domains := slices.Collect(maps.Keys(found))
	for domain := range winners {
		if !slices.Contains(domains, domain) {
			domains = append(domains, domain)
		}
	}
	slices.Sort(domains)
	var checks []check
	for _, domain := range domains {
		entry := func(status, source string, path *string, message string) check {
			return check{
				ID: "extensions." + domain, Component: "extensions", Status: status, Message: message,
				extensionKeys: &extensionKeys{Domain: domain, Source: source, Path: path},
			}
		}
		winner := winners[domain]
		if winner != nil {
			source, path := sourceExtension, &winner.path
			if winner.bundled {
				source, path = sourceBundled, nil
			}
			c := entry(statusPass, source, path, p.T("doctor.extension_runs", map[string]any{"Domain": domain, "Path": winner.path}))
			if winner.bundled {
				c.Message = p.T("doctor.extension_bundled", map[string]any{"Domain": domain})
			}
			if winner.broken != nil {
				problem := winner.invalid(p)
				c.Status, c.Code, c.Message, c.Hint = statusFail, problem.Code, problem.Message, problem.Hint
			}
			checks = append(checks, c)
		}
		for _, path := range found[domain] {
			data := map[string]any{"Domain": domain, "Path": path}
			var c check
			switch {
			case winner != nil && winner.bundled:
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden_by_bundled", data))
				c.HiddenBy = &hiddenBy{Source: sourceBundled}
			case winner == nil:
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden_by_core", data))
				c.HiddenBy = &hiddenBy{Source: sourceCore}
			case path == winner.path:
				continue
			default:
				data["Winner"] = winner.path
				c = entry(statusPass, sourceExtension, &path, p.T("doctor.extension_hidden", data))
				c.HiddenBy = &hiddenBy{Source: sourceExtension, Path: &winner.path}
			}
			checks = append(checks, c)
		}
	}
	return checks
}

// parseTimeout reads the value of --timeout: a duration such as 30s, 5m,
// or 2h, or a number of seconds, above 0. Without the flag, the default of
// its declaration applies.
func parseTimeout(p *i18n.Printer, flag *pflag.Flag) (time.Duration, *contract.Problem) {
	value := flag.Value.String()
	if !flag.Changed {
		value = doctorCommand.Flags[1].Default.(string)
	}
	d, err := time.ParseDuration(value)
	if seconds, convErr := strconv.ParseUint(value, 10, 31); convErr == nil {
		d, err = time.Duration(seconds)*time.Second, nil
	}
	if err != nil || d <= 0 {
		problem := command.Usage(p.T("doctor.timeout_invalid", map[string]any{"Value": value}), "ncly doctor --timeout 30s")
		return 0, &problem
	}
	return d, nil
}

// prober runs the program checks of doctor, one at a time, within the time
// left of --timeout.
type prober struct {
	p        *i18n.Printer
	g        command.Globals
	live     bool
	ctx      context.Context
	deadline time.Time
	// spent is set once a check ran out of time, so that no other starts.
	spent bool
	// signal is the signal that ended a probe.
	signal os.Signal
}

// programs checks each program that the manifest of an extension requires
// and does not mark conditional, in the order of their names. A broken
// manifest declares nothing that doctor can read.
func (pr *prober) programs(e *extension) []check {
	if e.broken != nil {
		return nil
	}
	var checks []check
	env := runner.Environ(os.Environ(), nil)
	for _, name := range e.manifest.Programs {
		r := e.manifest.Requirements[name]
		if r.Conditional || pr.signal != nil {
			continue
		}
		c := check{ID: e.domain + "." + name, Component: e.domain}
		data := map[string]any{"Name": name, "Version": r.Version, "Probe": strings.Join(append([]string{name}, r.ProbeArgs()...), " ")}
		timedOut := func(message string) {
			c.Status, c.Code, c.Message, c.Hint = statusFail, contract.CheckTimeout, message, pr.p.T("doctor.timeout_hint")
		}
		left := time.Until(pr.deadline)
		if pr.spent || left <= 0 {
			pr.spent = true
			timedOut(pr.p.T("doctor.not_started", data))
			checks = append(checks, c)
			continue
		}
		found, err := runner.CheckProgram(pr.ctx, name, r, env, left)
		data["Path"], data["Found"], data["Exit"] = found.Path, found.Version, found.Exit
		switch {
		case err != nil:
			found.Verdict = runner.ProgramProbeFailed
			data["Exit"] = -1
			found.Install = nil
		case found.Signal != nil:
			pr.signal = found.Signal
			return checks
		}
		c.Status = statusFail
		c.Code = contract.PrereqMissing
		switch found.Verdict {
		case runner.ProgramFound:
			c.Status, c.Code, c.Message = statusPass, "", pr.p.T("doctor.program_found", data)
			if found.Version != "" {
				c.Message = pr.p.T("doctor.program_version", data)
			}
		case runner.ProgramTimedOut:
			pr.spent = true
			timedOut(pr.p.T("doctor.program_timeout", data))
		case runner.ProgramMissing:
			c.Message = pr.p.T("doctor.program_missing", data)
		case runner.ProgramTooOld:
			c.Message = pr.p.T("doctor.program_old", data)
		case runner.ProgramProbeFailed:
			c.Message = pr.p.T("doctor.program_probe_failed", data)
		case runner.ProgramNoVersion:
			c.Message = pr.p.T("doctor.program_no_version", data)
		}
		if found.Failed() {
			c.Hint = pr.p.T("doctor.program_install", data)
			if found.Install != nil {
				c.Hint, c.Next = strings.Join(found.Install, " "), found.Install
			}
		}
		checks = append(checks, c)
	}
	return checks
}

// status runs the status command that an extension's manifest marks with
// doctor, after its probes, and returns its checks when its whole answer
// holds, or one <domain>.status check in place of an answer that doctor
// cannot list, as Status command in contract.md defines.
func (pr *prober) status(e *extension) []check {
	i := slices.IndexFunc(e.manifest.Commands, func(d contract.Command) bool { return d.Doctor })
	if e.broken != nil || i < 0 || pr.signal != nil {
		return nil
	}
	d := e.manifest.Commands[i]
	data := map[string]any{"Domain": e.domain, "Name": strings.Join(d.Path, " "), "Path": e.path}
	failed := func(code contract.Code, message, hint string, next []string) []check {
		return []check{{ID: e.domain + ".status", Component: e.domain, Status: statusFail, Code: code, Message: message, Hint: hint, Next: next}}
	}
	timedOut := func(message string) []check {
		pr.spent = true
		return failed(contract.CheckTimeout, message, pr.p.T("doctor.timeout_hint"), nil)
	}
	invalid := func(reason string) []check {
		data["Reason"] = reason
		return failed(contract.ExtensionInvalid, pr.p.T("doctor.status_invalid", data), pr.p.T("extension.invalid_hint", data), nil)
	}
	left := time.Until(pr.deadline)
	if pr.spent || left <= 0 {
		return timedOut(pr.p.T("doctor.not_started", data))
	}
	args := slices.Clone(d.Path[1:])
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
	program := runner.Program{
		Path:   e.path,
		Args:   append(args, "--"),
		Env:    runner.Environ(os.Environ(), extensionVars(withJSON(pr.g))),
		Stdout: &stdout,
		Stderr: stderr,
		Limit:  left,
	}
	if e.bundled {
		program.Name = executablePrefix + e.domain
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

// statusChecks reads the checks of a status command's report, or names the
// rule that one of them, or the verdict of the report, breaks: each check
// belongs to the domain, a pass holds no code nor hint, a failure's code
// maps to exit 78 and a warning's is a warning code, each holds a hint,
// and the exit code is 78 exactly when a check fails.
func statusChecks(p *i18n.Printer, e *extension, raw []json.RawMessage, exit int) ([]check, string) {
	registry := e.manifest.Registry()
	checks := []check{}
	failed := false
	for i, r := range raw {
		var c check
		data := map[string]any{"Index": i + 1, "Domain": e.domain}
		if err := json.Unmarshal(r, &c); err != nil || c.extensionKeys != nil {
			return nil, p.T("doctor.rule_shape", data)
		}
		data["ID"], data["Component"], data["Code"], data["Status"] = c.ID, c.Component, c.Code, c.Status
		code, isError := registry.Exit(c.Code)
		switch {
		case c.Component != e.domain || !strings.HasPrefix(c.ID, e.domain+".") || len(c.ID) == len(e.domain)+1:
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

// writeChecks prints the report in the look of sdk/tui: each check's status,
// id, and message in aligned columns, then the code and the hint of a
// problem under it.
func writeChecks(w io.Writer, p *i18n.Printer, checks []check) error {
	labels := map[string]string{
		statusPass: p.T("doctor.status_pass"),
		statusWarn: p.T("doctor.status_warn"),
		statusFail: p.T("doctor.status_fail"),
	}
	rows := make([]tui.Check, len(checks))
	for i, c := range checks {
		rows[i] = tui.Check{Status: c.Status, Label: labels[c.Status], ID: c.ID, Message: c.Message, Code: string(c.Code), Hint: c.Hint}
	}
	_, err := io.WriteString(w, tui.Checks(rows, p.T("error.hint")))
	return err
}
