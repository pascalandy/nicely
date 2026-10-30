package contract

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// ManifestFormat is the only manifest format that this version reads.
const ManifestFormat = 1

// Manifest declares an extension's commands and its own codes, as the
// Manifest section of contract.md defines.
type Manifest struct {
	Domain   string
	Name     string
	Commands []Command
	// Errors holds the own codes in the order that decides which one leads.
	Errors []OwnCode
	// DeclaredCodes holds errors and warnings with their scopes and TOML positions.
	DeclaredCodes []CodeCoverage
	// Programs holds the name of each program that a command may run, the
	// keys of the manifest's programs table, sorted.
	Programs []string
	// Requirements holds the entry of each program, by name, from M05.
	Requirements map[string]Requirement
	// Keys holds the entry of each service of the keys table, by name, from
	// M08.
	Keys map[string]KeyEntry
}

// KeyEntry is a service's entry in the keys table of a manifest, as
// Manifest in contract.md defines it.
type KeyEntry struct {
	// Optional marks a key that each command that declares the service
	// runs without, such as a harness that falls back to its own login.
	Optional bool
}

// Optional reports whether the keys table marks the key of a service
// optional.
func (m Manifest) Optional(service string) bool { return m.Keys[service].Optional }

// Requirement is a program's entry in the programs table of a manifest, as
// Manifest in contract.md defines it.
type Requirement struct {
	// Conditional marks a program that only some configs or invocations
	// need, which doctor never checks.
	Conditional bool `toml:"conditional" json:"conditional,omitempty"`
	// Version is the oldest version that works, such as 6.0, or empty when
	// finding the program is enough.
	Version string `toml:"version" json:"version,omitempty"`
	// Probe holds the arguments that make the program print its version,
	// or nil for the default, --version.
	Probe []string `toml:"probe" json:"probe,omitempty"`
	// Install holds one install command per package manager, by its name.
	Install map[string][]string `toml:"install" json:"install,omitempty"`
}

// ProbeArgs returns the arguments of the version probe.
func (r Requirement) ProbeArgs() []string {
	if r.Probe == nil {
		return []string{"--version"}
	}
	return r.Probe
}

// Command returns the declaration of a path, such as "hello run".
func (m Manifest) Command(path string) (Command, bool) {
	i := slices.IndexFunc(m.Commands, func(c Command) bool { return strings.Join(c.Path, " ") == path })
	if i < 0 {
		return Command{}, false
	}
	return m.Commands[i], true
}

// ids returns every catalog ID that the manifest names, in order.
func (m Manifest) ids() []string {
	var ids []string
	for _, c := range m.Commands {
		ids = append(ids, c.Summary)
		for _, a := range c.Args {
			ids = append(ids, a.Summary)
		}
		for _, f := range c.Flags {
			ids = append(ids, f.Summary)
		}
	}
	return ids
}

// RuleError names the first rule that a manifest or an answer breaks. Rule
// is the catalog ID of the message that explains it, and Data fills that
// message.
type RuleError struct {
	Rule string
	Data map[string]any
}

func (e *RuleError) Error() string { return fmt.Sprintf("%s %v", e.Rule, e.Data) }

func broken(rule string, data map[string]any) *RuleError {
	return &RuleError{Rule: rule, Data: data}
}

var (
	domainName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	pathWord   = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	shorthand  = regexp.MustCompile(`^[A-Za-z]$`)
	codeForm   = regexp.MustCompile(`^[A-Z0-9_]+$`)
	configWord = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

// ParseManifest reads an extension manifest and checks
// every rule of the Manifest section of contract.md, except the catalog IDs,
// which need the catalog. An empty slug takes the identity that the manifest
// declares, as for an extension whose manifest is embedded. The versions
// come first, so a manifest of another format is rejected before its other
// keys are read.
func ParseManifest(data []byte, slug string) (Manifest, error) {
	var versions struct {
		FormatVersion   int `toml:"format_version"`
		ContractVersion int `toml:"contract_version"`
	}
	if _, err := toml.Decode(string(data), &versions); err != nil {
		return Manifest{}, syntaxError(err)
	}
	switch {
	case versions.FormatVersion != ManifestFormat:
		return Manifest{}, broken("extension.format_version", map[string]any{"Value": versions.FormatVersion, "Want": ManifestFormat})
	case versions.ContractVersion != Version:
		return Manifest{}, broken("extension.contract_version", map[string]any{"Value": versions.ContractVersion, "Want": Version})
	}
	var file manifestFile
	if _, err := toml.Decode(string(data), &file); err != nil {
		return Manifest{}, syntaxError(err)
	}
	m, err := manifestIdentity(file.Domain, file.Name, slug)
	if err != nil {
		return Manifest{}, err
	}
	for _, raw := range file.Commands {
		c := Command{
			Path:     strings.Split(raw.Path, " "),
			Summary:  raw.Summary,
			Examples: raw.Examples,
			Modes:    Modes(raw.Modes),
			Requires: Requires(raw.Requires),
			Config:   raw.Config,
			Doctor:   raw.Doctor,
			Result:   Result{Kind: ResultKind(raw.Result.Kind), Schema: raw.Result.Schema, DryRunSchema: raw.Result.DryRunSchema},
		}
		if c.Result.Kind == "" {
			c.Result.Kind = ResultObject
		}
		if raw.Effects != nil {
			c.Effects = &Effects{UserWrites: raw.Effects.UserWrites, Network: raw.Effects.Network, Paid: raw.Effects.Paid}
		}
		for _, a := range raw.Args {
			c.Args = append(c.Args, Arg{
				Name: a.Name, Summary: a.Summary, Type: Type(a.Type), Required: a.Required, Variadic: a.Variadic,
				Values: a.Values, ValuesFrom: a.ValuesFrom.declared(),
			})
		}
		for _, f := range raw.Flags {
			c.Flags = append(c.Flags, Flag{
				Name: f.Name, Shorthand: f.Shorthand, Value: f.Value, Summary: f.Summary,
				Type: Type(f.Type), Required: f.Required, Variadic: f.Variadic, Default: f.Default,
				Values: f.Values, ValuesFrom: f.ValuesFrom.declared(),
			})
		}
		for _, g := range raw.FlagGroups {
			group := FlagGroup{Flags: g.Flags, Min: 0, Max: len(g.Flags)}
			if g.Min != nil {
				group.Min = *g.Min
			}
			if g.Max != nil {
				group.Max = *g.Max
			}
			c.FlagGroups = append(c.FlagGroups, group)
		}
		m.Commands = append(m.Commands, c)
	}
	for _, e := range file.Errors {
		var exit int
		if e.Exit != nil {
			exit = *e.Exit
		}
		m.Errors = append(m.Errors, OwnCode{Code: Code(e.Code), Exit: ExitCode(exit)})
	}
	if m.DeclaredCodes, err = parseDeclaredCodes(data, file); err != nil {
		return Manifest{}, err
	}
	m.Programs = slices.Sorted(maps.Keys(file.Programs))
	m.Requirements = file.Programs
	for _, name := range slices.Sorted(maps.Keys(file.Keys)) {
		optional, isBool := file.Keys[name].Optional.(bool)
		if file.Keys[name].Optional != nil && !isBool {
			return Manifest{}, broken("extension.key_optional", map[string]any{"Name": name})
		}
		if m.Keys == nil {
			m.Keys = map[string]KeyEntry{}
		}
		m.Keys[name] = KeyEntry{Optional: optional}
	}
	if err := m.check(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func syntaxError(err error) *RuleError {
	var parse toml.ParseError
	if errors.As(err, &parse) {
		return broken("extension.file_syntax", map[string]any{"Line": parse.Position.Line})
	}
	return broken("extension.file_invalid", map[string]any{"Value": err.Error()})
}

// IsService reports whether name is a service name, as Keys in contract.md
// defines it: lowercase letters and digits that start with a letter, other
// than keychain and grants, which name checks of doctor.
func IsService(name string) bool {
	return domainName.MatchString(name) && name != "keychain" && name != "grants"
}

// CheckCatalog reports the first catalog ID of the manifest that the en
// catalog lacks.
func (m Manifest) CheckCatalog(has func(id string) bool) error {
	for _, id := range m.ids() {
		if !has(id) {
			return broken("extension.catalog_id", map[string]any{"Value": id})
		}
	}
	return nil
}
