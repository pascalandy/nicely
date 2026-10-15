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

// Runs reports whether a command starts the extension. The domain alone
// prints its help, as Usage in contract.md says, and so does a command with
// children and no arguments, which only groups them.
func (m Manifest) Runs(c Command) bool {
	if len(c.Path) == 1 {
		return false
	}
	if len(c.Args) > 0 {
		return true
	}
	return !slices.ContainsFunc(m.Commands, func(other Command) bool {
		return len(other.Path) == len(c.Path)+1 && slices.Equal(other.Path[:len(c.Path)], c.Path)
	})
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
)

// IsDomain reports whether name can be a domain: lowercase letters and
// digits that start with a letter.
func IsDomain(name string) bool { return domainName.MatchString(name) }

// manifestFile is the TOML layout of a manifest. Keys that it lacks are
// ignored, as Compatibility asks of a reader.
type manifestFile struct {
	Domain   string `toml:"domain"`
	Commands []struct {
		Path     string   `toml:"path"`
		Summary  string   `toml:"summary"`
		Examples []string `toml:"examples"`
		Args     []struct {
			Name     string `toml:"name"`
			Summary  string `toml:"summary"`
			Type     string `toml:"type"`
			Required bool   `toml:"required"`
			Variadic bool   `toml:"variadic"`
		} `toml:"args"`
		Flags []struct {
			Name      string `toml:"name"`
			Shorthand string `toml:"shorthand"`
			Value     string `toml:"value"`
			Summary   string `toml:"summary"`
			Type      string `toml:"type"`
			Required  bool   `toml:"required"`
			Variadic  bool   `toml:"variadic"`
			Default   any    `toml:"default"`
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
		Result struct {
			Kind string `toml:"kind"`
		} `toml:"result"`
	} `toml:"commands"`
	Errors []struct {
		Code string `toml:"code"`
		Exit int    `toml:"exit"`
	} `toml:"errors"`
	// Programs holds one table per program, whose keys later milestones read.
	Programs map[string]toml.Primitive `toml:"programs"`
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
			Result:   Result{Kind: ResultKind(raw.Result.Kind)},
		}
		if c.Result.Kind == "" {
			c.Result.Kind = ResultObject
		}
		if raw.Effects != nil {
			c.Effects = &Effects{UserWrites: raw.Effects.UserWrites, Network: raw.Effects.Network, Paid: raw.Effects.Paid}
		}
		for _, a := range raw.Args {
			c.Args = append(c.Args, Arg{Name: a.Name, Summary: a.Summary, Type: Type(a.Type), Required: a.Required, Variadic: a.Variadic})
		}
		for _, f := range raw.Flags {
			c.Flags = append(c.Flags, Flag{
				Name: f.Name, Shorthand: f.Shorthand, Value: f.Value, Summary: f.Summary,
				Type: Type(f.Type), Required: f.Required, Variadic: f.Variadic, Default: f.Default,
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
	for _, c := range m.Commands {
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
		if err := m.checkRequires(path, c); err != nil {
			return err
		}
	}
	if err := m.checkPrograms(); err != nil {
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
		}
	}
	return nil
}

// typed reports whether a value that TOML decoded has a type: an integer
// for an integer, an integer or a float for a number.
func typed(value any, t Type) bool {
	switch value.(type) {
	case string:
		return t == TypeString
	case bool:
		return t == TypeBoolean
	case int64:
		return t == TypeInteger || t == TypeNumber
	case float64:
		return t == TypeNumber
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
// "/", and that a command names it.
func (m Manifest) checkPrograms() error {
	for _, name := range m.Programs {
		named := slices.ContainsFunc(m.Commands, func(c Command) bool { return slices.Contains(c.Requires.Bins, name) })
		if name == "" || name == "." || name == ".." || strings.Contains(name, "/") || !named {
			return broken("extension.program", map[string]any{"Name": name})
		}
	}
	return nil
}

// IsService reports whether name is a service name, as Keys in contract.md
// defines it: lowercase letters and digits that start with a letter, other
// than keychain.
func IsService(name string) bool { return domainName.MatchString(name) && name != "keychain" }

// codePrefix starts every code that the extension owns, such as HELLO_.
func (m Manifest) codePrefix() string { return strings.ToUpper(m.Domain) + "_" }

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
