package contract

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// ManifestFormat is the only manifest format that this version reads.
const ManifestFormat = 1

// Manifest declares an extension's commands and its own error codes, as the
// Manifest section of contract.md defines.
type Manifest struct {
	Domain   string
	Commands []Command
	// Errors holds the own codes in the order that decides which one leads.
	Errors []OwnCode
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

// OwnCode is an error code that an extension declares.
type OwnCode struct {
	Code Code
	Exit ExitCode
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

// valuesFromFile is the TOML layout of a values_from table.
type valuesFromFile struct {
	Command string `toml:"command"`
	Key     string `toml:"key"`
	Field   string `toml:"field"`
}

// declared returns the ValuesFrom of a values_from table, or nil without
// one.
func (v *valuesFromFile) declared() *ValuesFrom {
	if v == nil {
		return nil
	}
	return &ValuesFrom{Command: strings.Fields(v.Command), Key: v.Key, Field: v.Field}
}

// manifestFile is the TOML layout of a manifest. Keys that it lacks are
// ignored, as Compatibility asks of a reader.
type manifestFile struct {
	Domain   string `toml:"domain"`
	Commands []struct {
		Path     string   `toml:"path"`
		Summary  string   `toml:"summary"`
		Examples []string `toml:"examples"`
		Args     []struct {
			Name       string          `toml:"name"`
			Summary    string          `toml:"summary"`
			Type       string          `toml:"type"`
			Required   bool            `toml:"required"`
			Variadic   bool            `toml:"variadic"`
			Values     []string        `toml:"values"`
			ValuesFrom *valuesFromFile `toml:"values_from"`
		} `toml:"args"`
		Flags []struct {
			Name       string          `toml:"name"`
			Shorthand  string          `toml:"shorthand"`
			Value      string          `toml:"value"`
			Summary    string          `toml:"summary"`
			Type       string          `toml:"type"`
			Required   bool            `toml:"required"`
			Variadic   bool            `toml:"variadic"`
			Values     []string        `toml:"values"`
			Default    any             `toml:"default"`
			ValuesFrom *valuesFromFile `toml:"values_from"`
		} `toml:"flags"`
		FlagGroups []struct {
			Flags []string `toml:"flags"`
			Min   *int     `toml:"min"`
			Max   *int     `toml:"max"`
		} `toml:"flag_groups"`
		Effects *struct {
			UserWrites bool `toml:"user_writes"`
			Network    bool `toml:"network"`
			Paid       bool `toml:"paid"`
		} `toml:"effects"`
		Modes struct {
			DryRun   bool `toml:"dry_run"`
			Recorded bool `toml:"recorded"`
			Resume   bool `toml:"resume"`
		} `toml:"modes"`
		Requires struct {
			Bins []string `toml:"bins"`
			Keys []string `toml:"keys"`
		} `toml:"requires"`
		Config []string `toml:"config"`
		Doctor bool     `toml:"doctor"`
		Result struct {
			Kind         string         `toml:"kind"`
			Schema       map[string]any `toml:"schema"`
			DryRunSchema map[string]any `toml:"dry_run_schema"`
		} `toml:"result"`
	} `toml:"commands"`
	Errors []struct {
		Code string `toml:"code"`
		Exit int    `toml:"exit"`
	} `toml:"errors"`
	Programs map[string]Requirement `toml:"programs"`
	// Keys reads optional as any, so that a value that is not a boolean
	// gets its own reason.
	Keys map[string]struct {
		Optional any `toml:"optional"`
	} `toml:"keys"`
}

// ParseManifest reads the manifest of the extension of a domain and checks
// every rule of the Manifest section of contract.md, except the catalog IDs,
// which need the catalog. An empty domain takes the one that the manifest
// declares, as for an extension whose manifest is embedded. The versions
// come first, so a manifest of another format is rejected before its other
// keys are read.
func ParseManifest(data []byte, domain string) (Manifest, error) {
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
	if domain == "" {
		domain = file.Domain
	}
	if file.Domain != domain || !IsDomain(domain) {
		return Manifest{}, broken("extension.domain", map[string]any{"Value": file.Domain, "Want": domain})
	}
	m := Manifest{Domain: domain}
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
		m.Errors = append(m.Errors, OwnCode{Code: Code(e.Code), Exit: ExitCode(e.Exit)})
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

func (m Manifest) check() error {
	if _, found := m.Command(m.Domain); !found {
		return broken("extension.domain_command", map[string]any{"Path": m.Domain})
	}
	seen := map[string]bool{}
	for i, c := range m.Commands {
		path := strings.Join(c.Path, " ")
		for _, word := range c.Path {
			if !pathWord.MatchString(word) {
				return broken("extension.path", map[string]any{"Path": path})
			}
		}
		switch {
		case c.Path[0] != m.Domain:
			return broken("extension.path_domain", map[string]any{"Path": path, "Domain": m.Domain})
		case seen[path]:
			return broken("extension.path_twice", map[string]any{"Path": path})
		}
		seen[path] = true
		if parent := strings.Join(c.Path[:len(c.Path)-1], " "); parent != "" {
			if _, found := m.Command(parent); !found {
				return broken("extension.parent", map[string]any{"Path": path, "Parent": parent})
			}
		}
		if !slices.Contains([]ResultKind{ResultObject, ResultBatch, ResultReport, ResultText}, c.Result.Kind) {
			return broken("extension.result", map[string]any{"Path": path, "Value": string(c.Result.Kind)})
		}
		if err := checkNames(path, c); err != nil {
			return err
		}
		if err := checkValues(path, c); err != nil {
			return err
		}
		if err := checkGroups(path, c); err != nil {
			return err
		}
		if err := m.checkValuesFrom(path, c); err != nil {
			return err
		}
		if err := m.checkRequires(path, c); err != nil {
			return err
		}
		if err := m.checkConfig(path, c); err != nil {
			return err
		}
		if c.Doctor && !m.IsStatus(c) {
			return broken("extension.doctor", map[string]any{"Path": path})
		}
		if err := checkResult(path, &m.Commands[i]); err != nil {
			return err
		}
	}
	if err := m.checkPrograms(); err != nil {
		return err
	}
	if err := m.checkKeys(); err != nil {
		return err
	}
	return m.checkCodes()
}

// checkNames applies the rules of names and shorthands to the arguments and
// flags of one command, so that no declaration can break the parser.
func checkNames(path string, c Command) error {
	for _, a := range c.Args {
		if !pathWord.MatchString(a.Name) {
			return broken("extension.name", map[string]any{"Path": path, "Name": a.Name})
		}
	}
	names, shorthands := map[string]bool{}, map[string]bool{}
	for _, g := range append(slices.Clone(GlobalFlags), DryRunFlag, RunIDFlag) {
		names[g.Name] = true
		if g.Shorthand != "" {
			shorthands[g.Shorthand] = true
		}
	}
	for _, f := range c.Flags {
		switch {
		case !pathWord.MatchString(f.Name):
			return broken("extension.name", map[string]any{"Path": path, "Name": f.Name})
		case f.Value != "" && !pathWord.MatchString(f.Value):
			return broken("extension.name", map[string]any{"Path": path, "Name": f.Value})
		case f.Shorthand != "" && !shorthand.MatchString(f.Shorthand):
			return broken("extension.shorthand", map[string]any{"Path": path, "Name": f.Shorthand})
		case names[f.Name]:
			return broken("extension.flag_taken", map[string]any{"Path": path, "Name": "--" + f.Name})
		case f.Shorthand != "" && shorthands[f.Shorthand]:
			return broken("extension.flag_taken", map[string]any{"Path": path, "Name": "-" + f.Shorthand})
		}
		names[f.Name] = true
		if f.Shorthand != "" {
			shorthands[f.Shorthand] = true
		}
	}
	return nil
}

// checkValuesFrom applies the rules of values_from to one command: it sits
// on an argument or on a flag that takes a value, names a command, a key,
// and a field, and, when the command starts with this domain, names a
// command of this manifest that qualifies.
func (m Manifest) checkValuesFrom(path string, c Command) error {
	type named struct {
		name string
		from *ValuesFrom
	}
	var declared []named
	for _, a := range c.Args {
		declared = append(declared, named{a.Name, a.ValuesFrom})
	}
	for _, f := range c.Flags {
		if f.ValuesFrom != nil && f.Value == "" {
			return broken("extension.values_from_switch", map[string]any{"Path": path, "Name": f.Name})
		}
		declared = append(declared, named{f.Name, f.ValuesFrom})
	}
	for _, d := range declared {
		data := map[string]any{"Path": path, "Name": d.name}
		switch from := d.from; {
		case from == nil:
		case len(from.Command) == 0 || from.Key == "" || from.Field == "":
			return broken("extension.values_from_empty", data)
		case from.Command[0] == m.Domain:
			if list, found := m.Command(strings.Join(from.Command, " ")); !found || !m.Qualifies(list) {
				data["Command"] = strings.Join(from.Command, " ")
				return broken("extension.values_from_command", data)
			}
		}
	}
	return nil
}

// Qualifies reports whether a command of the manifest may give completion
// its values, as Completion values in contract.md defines: a list command
// that runs, takes no argument, answers an object, and declares effects
// that neither write user files nor bill. A command of core never
// qualifies, which the caller checks.
func (m Manifest) Qualifies(c Command) bool {
	return m.Runs(c) && c.Path[len(c.Path)-1] == "list" && len(c.Args) == 0 &&
		c.Result.Kind == ResultObject && c.Effects != nil && !c.Effects.UserWrites && !c.Effects.Paid
}

// IsStatus reports whether a command qualifies as the status command of
// its domain, as Status command in contract.md defines it: the domain
// followed by status, which runs, answers a report, takes no argument and
// no flag group, has live and timeout as its only own flags, neither
// required nor limited to values, with live a switch that defaults to
// false and timeout a string, so each call that doctor builds passes
// validation, declares effects that neither write nor bill, and live when
// they use the network, and sets no mode.
func (m Manifest) IsStatus(c Command) bool {
	live := false
	for _, f := range c.Flags {
		switch {
		case f.Required || len(f.Values) > 0:
			return false
		case f.Name == "live" && f.Value == "" && f.FlagType() == TypeBoolean && (f.Default == nil || f.Default == false):
			live = true
		case f.Name == "timeout" && f.Value != "" && f.FlagType() == TypeString:
		default:
			return false
		}
	}
	return slices.Equal(c.Path, []string{m.Domain, "status"}) && m.Runs(c) &&
		c.Result.Kind == ResultReport && len(c.Args) == 0 && len(c.FlagGroups) == 0 &&
		c.Effects != nil && !c.Effects.UserWrites && !c.Effects.Paid && (!c.Effects.Network || live) &&
		!c.Modes.DryRun && !c.Modes.Recorded && !c.Modes.Resume
}

// checkValues applies the rules of types, requiredness, variadic values,
// and defaults to the arguments and flags of one command.
func checkValues(path string, c Command) error {
	data := func(name string) map[string]any { return map[string]any{"Path": path, "Name": name} }
	optional := false
	for i, a := range c.Args {
		switch {
		case !slices.Contains(Types, a.ArgType()):
			return broken("extension.type", map[string]any{"Path": path, "Name": a.Name, "Value": string(a.Type)})
		case a.Required && optional:
			return broken("extension.required", data(a.Name))
		case a.Variadic && i != len(c.Args)-1:
			return broken("extension.variadic", data(a.Name))
		case !fitValues(a.Values, a.ArgType()):
			return broken("extension.values", data(a.Name))
		}
		optional = optional || !a.Required
	}
	for _, f := range c.Flags {
		switch {
		case !slices.Contains(Types, f.FlagType()), f.Value == "" && f.FlagType() != TypeBoolean:
			return broken("extension.type", map[string]any{"Path": path, "Name": f.Name, "Value": string(f.FlagType())})
		case f.Required && f.Value == "":
			return broken("extension.required", data(f.Name))
		case f.Variadic && f.Value == "":
			return broken("extension.variadic", data(f.Name))
		case f.Default != nil && (f.Required || f.Variadic || !typed(f.Default, f.FlagType()) || f.Value == "" && f.Default != false):
			return broken("extension.default", data(f.Name))
		case f.Values != nil && (f.Value == "" || !fitValues(f.Values, f.FlagType())):
			return broken("extension.values", data(f.Name))
		}
	}
	return nil
}

// fitValues reports whether the accepted values of an argument or a flag
// are absent, or hold at least one value, each once, that its type reads.
func fitValues(values []string, t Type) bool {
	if values == nil {
		return true
	}
	for i, v := range values {
		if v == "" || !t.Reads(v) || slices.Index(values, v) != i {
			return false
		}
	}
	return len(values) > 0
}

// typed reports whether a value that TOML decoded has a type: an integer
// for an integer, an integer or a finite float for a number, since JSON has
// no NaN or infinity.
func typed(value any, t Type) bool {
	switch v := value.(type) {
	case string:
		return t == TypeString
	case bool:
		return t == TypeBoolean
	case int64:
		return t == TypeInteger || t == TypeNumber
	case float64:
		return t == TypeNumber && !math.IsNaN(v) && !math.IsInf(v, 0)
	}
	return false
}

// checkGroups applies the rules of flag groups to one command: each names
// own flags, at least two and each once, with 0 <= min <= max <= flags and
// max >= 1.
func checkGroups(path string, c Command) error {
	for _, g := range c.FlagGroups {
		bad := len(g.Flags) < 2 || g.Min < 0 || g.Min > g.Max || g.Max < 1 || g.Max > len(g.Flags)
		for j, name := range g.Flags {
			own := slices.ContainsFunc(c.Flags, func(f Flag) bool { return f.Name == name })
			bad = bad || !own || slices.Index(g.Flags, name) != j
		}
		if bad {
			return broken("extension.flag_group", map[string]any{"Path": path, "Value": "--" + strings.Join(g.Flags, ", --")})
		}
	}
	return nil
}

// checkRequires checks the programs and the services that one command
// names: each program is a key of programs, each service a service name,
// and each name appears once.
func (m Manifest) checkRequires(path string, c Command) error {
	for i, name := range c.Requires.Bins {
		if !slices.Contains(m.Programs, name) || slices.Index(c.Requires.Bins, name) != i {
			return broken("extension.bin", map[string]any{"Path": path, "Name": name})
		}
	}
	for i, name := range c.Requires.Keys {
		if !IsService(name) || slices.Index(c.Requires.Keys, name) != i {
			return broken("extension.key", map[string]any{"Path": path, "Name": name})
		}
	}
	return nil
}

// checkPrograms checks that each key of programs is a file name, without
// "/", and that a command names it, then its entry: a version of numbers
// joined by dots, a probe that is not empty, and install commands that are
// not empty.
func (m Manifest) checkPrograms() error {
	for _, name := range m.Programs {
		named := slices.ContainsFunc(m.Commands, func(c Command) bool { return slices.Contains(c.Requires.Bins, name) })
		if name == "" || name == "." || name == ".." || strings.Contains(name, "/") || !named {
			return broken("extension.program", map[string]any{"Name": name})
		}
		r := m.Requirements[name]
		switch {
		case r.Version != "" && !versionNumber.MatchString(r.Version):
			return broken("extension.program_version", map[string]any{"Name": name, "Value": r.Version})
		case r.Probe != nil && len(r.Probe) == 0:
			return broken("extension.program_probe", map[string]any{"Name": name})
		}
		for _, manager := range slices.Sorted(maps.Keys(r.Install)) {
			if len(r.Install[manager]) == 0 {
				return broken("extension.program_install", map[string]any{"Name": name, "Manager": manager})
			}
		}
	}
	return nil
}

// checkKeys checks that each service of the keys table is one that the
// requires.keys of a command names.
func (m Manifest) checkKeys() error {
	for _, name := range slices.Sorted(maps.Keys(m.Keys)) {
		if !slices.ContainsFunc(m.Commands, func(c Command) bool { return slices.Contains(c.Requires.Keys, name) }) {
			return broken("extension.key_table", map[string]any{"Name": name})
		}
	}
	return nil
}

// versionNumber is a version as a manifest declares it: one or more numbers
// joined by dots.
var versionNumber = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

// checkConfig checks the config keys that one command reads: each is a
// dotted path of at least two key words under the domain, and appears once.
func (m Manifest) checkConfig(path string, c Command) error {
	for i, key := range c.Config {
		words := strings.Split(key, ".")
		bad := len(words) < 2 || words[0] != m.Domain || slices.Index(c.Config, key) != i
		for _, word := range words {
			bad = bad || !configWord.MatchString(word)
		}
		if bad {
			return broken("extension.config_key", map[string]any{"Path": path, "Name": key, "Domain": m.Domain})
		}
	}
	return nil
}

// checkResult checks the schemas of one command's result, and resolves
// their references. A text result has none, and only a command that
// supports dry run has dry_run_schema. Each is an object schema whose
// references are local, and whose top level holds only the keywords that
// AnswerSchema carries into the answer's schema.
func checkResult(path string, c *Command) error {
	for _, s := range []struct {
		schema *map[string]any
		name   string
		allow  bool
	}{
		{&c.Result.Schema, "schema", c.Result.Kind != ResultText},
		{&c.Result.DryRunSchema, "dry_run_schema", c.Result.Kind != ResultText && c.Modes.DryRun},
	} {
		if *s.schema == nil {
			continue
		}
		resolved, err := ResolveSchema(*s.schema)
		if !s.allow || err != nil || !IsDataSchema(resolved) {
			return broken("extension.schema", map[string]any{"Path": path, "Name": s.name})
		}
		*s.schema = resolved
	}
	return nil
}

// IsService reports whether name is a service name, as Keys in contract.md
// defines it: lowercase letters and digits that start with a letter, other
// than keychain and grants, which name checks of doctor.
func IsService(name string) bool {
	return domainName.MatchString(name) && name != "keychain" && name != "grants"
}

func (m Manifest) checkCodes() error {
	prefix := m.codePrefix()
	seen := map[Code]bool{}
	for _, e := range m.Errors {
		_, core := e.Code.Exit()
		switch {
		case !strings.HasPrefix(string(e.Code), prefix) || !codeForm.MatchString(string(e.Code)):
			return broken("extension.code_prefix", map[string]any{"Code": string(e.Code), "Prefix": prefix})
		case core || e.Code.IsWarning():
			return broken("extension.code_core", map[string]any{"Code": string(e.Code)})
		case seen[e.Code]:
			return broken("extension.code_twice", map[string]any{"Code": string(e.Code)})
		case !slices.Contains([]ExitCode{ExitRuntime, ExitUsage, ExitHuman}, e.Exit):
			return broken("extension.code_exit", map[string]any{"Code": string(e.Code), "Value": int(e.Exit)})
		}
		seen[e.Code] = true
	}
	return nil
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
