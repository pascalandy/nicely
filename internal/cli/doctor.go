package cli

import (
	"context"
	_ "embed"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
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
	c := command.Build(p, p, doctorCommand, func(c *cobra.Command, args []string) error {
		timeout, problem := parseTimeout(p, c.Flag("timeout"))
		if problem != nil {
			return command.Fail(*problem)
		}
		live, _ := c.Flags().GetBool("live")
		// One keychain serves the whole report, so a keychain that a status
		// command's key found unavailable stays so for the auth component.
		k := keychain.New(b, interactive(g, c.OutOrStdout()))
		probes := &prober{p: p, g: g, live: live, ctx: runner.OnSignal(), deadline: time.Now().Add(timeout), keys: k}
		components := []component{
			{name: "core", checks: func() []check { return coreChecks(p, extensions) }},
			{name: "auth", checks: func() []check { return probes.auth(extensions, k) }, last: true},
			{name: "extensions", checks: func() []check { return extensionChecks(p, extensions) }},
		}
		for _, e := range extensions {
			components = append(components, component{name: e.slug, checks: func() []check {
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
	})
	c.ValidArgsFunction = func(_ *cobra.Command, args []string, prefix string) ([]cobra.Completion, cobra.ShellCompDirective) {
		var names []cobra.Completion
		if len(args) == 0 {
			components := []string{"core", "auth", "extensions"}
			for _, e := range extensions {
				components = append(components, e.slug)
			}
			for _, name := range components {
				if strings.HasPrefix(name, prefix) {
					names = append(names, name)
				}
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
	root.AddCommand(c)
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
	// keys reads the keys that a status command needs under --live.
	keys *keychain.Keychain
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
		c := check{ID: e.slug + "." + name, Component: e.slug}
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
