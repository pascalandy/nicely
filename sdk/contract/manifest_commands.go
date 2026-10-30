package contract

import (
	"slices"
	"strings"
)

func (m Manifest) check() error {
	if _, found := m.Command(strings.Join(m.Root(), " ")); !found {
		return broken("extension.domain_command", map[string]any{"Path": strings.Join(m.Root(), " ")})
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
		case !m.Owns(c.Path):
			return broken("extension.path_domain", map[string]any{"Path": path, "Domain": strings.Join(m.Root(), " ")})
		case seen[path]:
			return broken("extension.path_twice", map[string]any{"Path": path})
		}
		seen[path] = true
		if parent := strings.Join(c.Path[:len(c.Path)-1], " "); len(c.Path) > len(m.Root()) {
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

// checkValuesFrom checks each argument or value-taking flag's provider command, key, and field.
// A provider beneath this extension root must name a qualifying command of this manifest.
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
		case m.Owns(from.Command):
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

// IsStatus checks the status command under the extension root, as Status command in contract.md defines it.
// It runs, answers a report, and takes no argument or flag group.
// Its only own flags are live and timeout, neither required nor limited to values.
// Live is a switch that defaults to false, and timeout is a string, so each doctor call passes validation.
// Declared effects neither write nor bill, and network use requires live.
// It sets no mode.
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
	return slices.Equal(c.Path, append(m.Root(), "status")) && m.Runs(c) &&
		c.Result.Kind == ResultReport && len(c.Args) == 0 && len(c.FlagGroups) == 0 &&
		c.Effects != nil && !c.Effects.UserWrites && !c.Effects.Paid && (!c.Effects.Network || live) &&
		!c.Modes.DryRun && !c.Modes.Recorded && !c.Modes.Resume
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
