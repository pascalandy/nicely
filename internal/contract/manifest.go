package contract

import (
	"errors"
	"fmt"
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

// ManifestError names the first rule that a manifest breaks. Rule is the
// catalog ID of the message that explains it, and Data fills that message.
type ManifestError struct {
	Rule string
	Data map[string]any
}

func (e *ManifestError) Error() string { return fmt.Sprintf("%s %v", e.Rule, e.Data) }

func broken(rule string, data map[string]any) *ManifestError {
	return &ManifestError{Rule: rule, Data: data}
}

var (
	domainName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	pathWord   = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	shorthand  = regexp.MustCompile(`^[A-Za-z]$`)
	ownCode    = regexp.MustCompile(`^[A-Z0-9_]+$`)
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
			Name    string `toml:"name"`
			Summary string `toml:"summary"`
		} `toml:"args"`
		Flags []struct {
			Name      string `toml:"name"`
			Shorthand string `toml:"shorthand"`
			Value     string `toml:"value"`
			Summary   string `toml:"summary"`
		} `toml:"flags"`
		Result struct {
			Kind string `toml:"kind"`
		} `toml:"result"`
	} `toml:"commands"`
	Errors []struct {
		Code string `toml:"code"`
		Exit int    `toml:"exit"`
	} `toml:"errors"`
}

// ParseManifest reads the manifest of the extension of a domain and checks
// every rule of the Manifest section of contract.md, except the catalog IDs,
// which need the catalog. The versions come first, so a manifest of another
// format is rejected before its other keys are read.
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
	if file.Domain != domain {
		return Manifest{}, broken("extension.domain", map[string]any{"Value": file.Domain, "Want": domain})
	}
	m := Manifest{Domain: domain}
	for _, raw := range file.Commands {
		c := Command{
			Path:     strings.Split(raw.Path, " "),
			Summary:  raw.Summary,
			Examples: raw.Examples,
			Result:   Result{Kind: ResultKind(raw.Result.Kind)},
		}
		if c.Result.Kind == "" {
			c.Result.Kind = ResultObject
		}
		for _, a := range raw.Args {
			c.Args = append(c.Args, Arg{Name: a.Name, Summary: a.Summary})
		}
		for _, f := range raw.Flags {
			c.Flags = append(c.Flags, Flag{Name: f.Name, Shorthand: f.Shorthand, Value: f.Value, Summary: f.Summary})
		}
		m.Commands = append(m.Commands, c)
	}
	for _, e := range file.Errors {
		m.Errors = append(m.Errors, OwnCode{Code: Code(e.Code), Exit: ExitCode(e.Exit)})
	}
	if err := m.check(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func syntaxError(err error) *ManifestError {
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
	for _, g := range GlobalFlags {
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

func (m Manifest) checkCodes() error {
	prefix := strings.ToUpper(m.Domain) + "_"
	seen := map[Code]bool{}
	for _, e := range m.Errors {
		_, core := e.Code.Exit()
		switch {
		case !strings.HasPrefix(string(e.Code), prefix) || !ownCode.MatchString(string(e.Code)):
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
