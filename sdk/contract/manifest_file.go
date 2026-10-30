package contract

import "strings"

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
	Domain   string  `toml:"domain"`
	Name     *string `toml:"name"`
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
