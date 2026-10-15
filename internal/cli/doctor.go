package cli

import (
	_ "embed"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

//go:embed doctor_schema.json
var doctorSchema []byte

var doctorCommand = contract.Command{
	Path:    []string{"doctor"},
	Result:  contract.Result{Kind: contract.ResultReport, Schema: mustSchema(doctorSchema)},
	Summary: "doctor.summary",
	Args:    []contract.Arg{{Name: "component", Summary: "doctor.component"}},
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly doctor`,
		`ncly doctor core --json`,
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
// that the component argument gives it.
type component struct {
	name   string
	checks func() []check
}

// addDoctor adds ncly doctor, which reports its checks on stdout even when
// one fails, as a report command does. It never starts an extension.
func addDoctor(p *i18n.Printer, root *cobra.Command, extensions []*extension) {
	root.AddCommand(command.Build(p, p, doctorCommand, func(_ *cobra.Command, args []string) error {
		components := []component{
			{"core", func() []check { return coreChecks(p, extensions) }},
			{"extensions", func() []check { return extensionChecks(p, extensions) }},
		}
		if len(args) > 0 {
			i := slices.IndexFunc(components, func(c component) bool { return c.name == args[0] })
			if i < 0 {
				return command.Fail(noComponent(p, args[0], components))
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
		for _, c := range components {
			r.Checks = append(r.Checks, c.checks()...)
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

// writeChecks prints one line per check: its status, its id, and its
// message, then its hint on the next line, under the id.
func writeChecks(w io.Writer, p *i18n.Printer, checks []check) error {
	labels := map[string]string{
		statusPass: p.T("doctor.status_pass"),
		statusWarn: p.T("doctor.status_warn"),
		statusFail: p.T("doctor.status_fail"),
	}
	width := 0
	for _, label := range labels {
		width = max(width, utf8.RuneCountInString(label))
	}
	pad := func(s string) string { return s + strings.Repeat(" ", width-utf8.RuneCountInString(s)) }
	var b strings.Builder
	for _, c := range checks {
		b.WriteString(pad(labels[c.Status]) + "  " + c.ID + "  " + c.Message + "\n")
		if c.Hint != "" {
			b.WriteString(strings.Repeat(" ", width+2) + p.T("error.hint") + " " + c.Hint + "\n")
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
