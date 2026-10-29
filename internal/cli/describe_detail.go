package cli

import (
	"cmp"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// commandDetail is the answer of ncly describe for one command.
type commandDetail struct {
	Commands []detail `json:"commands"`
}

// detail is the description of one command, as the detailed entry of ncly
// describe in core-spec.md defines it.
type detail struct {
	entry
	Args       []argDetail    `json:"args"`
	Flags      []flagDetail   `json:"flags"`
	FlagGroups []groupDetail  `json:"flag_groups"`
	Result     resultDetail   `json:"result"`
	Effects    effectsDetail  `json:"effects"`
	Requires   requiresDetail `json:"requires"`
	Modes      modesDetail    `json:"modes"`
	Config     []config.Value `json:"config"`
	Examples   []string       `json:"examples"`
}

type argDetail struct {
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Variadic bool   `json:"variadic"`
	Values   []any  `json:"values"`
	// ValuesFrom is present only when the declaration has one.
	ValuesFrom *valuesFromDetail `json:"values_from,omitempty"`
}

// valuesFromDetail names the list command whose answer completion suggests
// values from, as Completion values in contract.md defines.
type valuesFromDetail struct {
	Command string `json:"command"`
	Key     string `json:"key"`
	Field   string `json:"field"`
}

func describeValuesFrom(v *contract.ValuesFrom) *valuesFromDetail {
	if v == nil {
		return nil
	}
	return &valuesFromDetail{Command: strings.Join(v.Command, " "), Key: v.Key, Field: v.Field}
}

type flagDetail struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand"`
	Value     string `json:"value"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	Variadic  bool   `json:"variadic"`
	Values    []any  `json:"values"`
	Default   any    `json:"default"`
	Summary   string `json:"summary"`
	// ValuesFrom is present only when the declaration has one.
	ValuesFrom *valuesFromDetail `json:"values_from,omitempty"`
}

type resultDetail struct {
	Kind         contract.ResultKind `json:"kind"`
	Keys         []string            `json:"keys"`
	Schema       map[string]any      `json:"schema,omitempty"`
	DryRunSchema map[string]any      `json:"dry_run_schema,omitempty"`
}

type groupDetail struct {
	Flags []string `json:"flags"`
	Min   int      `json:"min"`
	Max   int      `json:"max"`
}

type effectsDetail struct {
	UserWrites bool `json:"user_writes"`
	Network    bool `json:"network"`
	Paid       bool `json:"paid"`
}

type requiresDetail struct {
	Bins []binDetail `json:"bins"`
	Keys []keyDetail `json:"keys"`
}

// binDetail is a program that a command may run, with the keys of its
// entry in the manifest's programs table, each present only when declared.
type binDetail struct {
	Name string `json:"name"`
	contract.Requirement
}

// keyDetail is a service whose key a command may need, with optional from
// the service's entry in the manifest's keys table, or false without one.
type keyDetail struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
}

type modesDetail struct {
	DryRun   bool `json:"dry_run"`
	Recorded bool `json:"recorded"`
	Resume   bool `json:"resume"`
}

// detail describes the command of an entry. p renders the text of the
// flags that its modes add, which the SDK's catalog holds. Its config lists
// lang, which core resolves for the text of every command, then each key
// that the command declares, from values.
func (e entry) detail(p *i18n.Printer, lang config.Value, values config.Values) detail {
	d := e.declaration
	out := detail{
		entry:      e,
		Args:       []argDetail{},
		Flags:      []flagDetail{},
		FlagGroups: []groupDetail{},
		Result:     describeResult(d),
		Effects:    effectsDetail(d.DeclaredEffects()),
		Requires:   requiresDetail{Bins: []binDetail{}, Keys: []keyDetail{}},
		Modes:      modesDetail(d.Modes),
		Config:     []config.Value{lang},
		Examples:   append([]string{}, d.Examples...),
	}
	for _, name := range d.Requires.Bins {
		out.Requires.Bins = append(out.Requires.Bins, binDetail{name, e.requirements[name]})
	}
	for _, name := range d.Requires.Keys {
		out.Requires.Keys = append(out.Requires.Keys, keyDetail{name, e.keys[name].Optional})
	}
	for _, key := range d.Config {
		out.Config = append(out.Config, values.Under(key)...)
	}
	for _, a := range d.Args {
		out.Args = append(out.Args, argDetail{
			Name: a.Name, Summary: e.text.T(a.Summary), Type: string(a.ArgType()),
			Required: a.Required, Variadic: a.Variadic, Values: typedValues(a.ArgType(), a.Values),
			ValuesFrom: describeValuesFrom(a.ValuesFrom),
		})
	}
	flag := func(f contract.Flag, text *i18n.Printer) flagDetail {
		return flagDetail{
			Name: f.Name, Shorthand: f.Shorthand, Value: f.Value, Type: string(f.FlagType()),
			Required: f.Required, Variadic: f.Variadic, Values: typedValues(f.FlagType(), f.Values),
			Default: f.FlagDefault(), Summary: text.T(f.Summary),
			ValuesFrom: describeValuesFrom(f.ValuesFrom),
		}
	}
	for _, f := range d.Flags {
		out.Flags = append(out.Flags, flag(f, e.text))
	}
	for _, f := range d.ModeFlags() {
		out.Flags = append(out.Flags, flag(f, p))
	}
	for _, g := range d.FlagGroups {
		out.FlagGroups = append(out.FlagGroups, groupDetail{Flags: g.Flags, Min: g.Min, Max: g.Max})
	}
	return out
}

// describeResult gives the kind of a command's result, its keys, and the
// schemas of its answers: the normal one, and the dry-run one when the
// command supports dry run. A text result has neither. A dry-run answer
// describes the work instead of doing it, so it has no results even for a
// batch.
func describeResult(d contract.Command) resultDetail {
	kind := cmp.Or(d.Result.Kind, contract.ResultObject)
	r := resultDetail{Kind: kind, Keys: contract.Keys(kind, d.Result.Schema), Schema: contract.AnswerSchema(kind, d.Result.Schema)}
	if d.Modes.DryRun && kind != contract.ResultText {
		r.DryRunSchema = contract.AnswerSchema(contract.ResultObject, d.Result.DryRunSchema)
	}
	return r
}

// typedValues turns accepted values into JSON values of their type.
func typedValues(t contract.Type, values []string) []any {
	out := []any{}
	for _, v := range values {
		var typed any = v
		switch t {
		case contract.TypeBoolean:
			typed, _ = strconv.ParseBool(v)
		case contract.TypeInteger:
			typed, _ = strconv.ParseInt(v, 10, 64)
		case contract.TypeNumber:
			typed, _ = strconv.ParseFloat(v, 64)
		}
		out = append(out, typed)
	}
	return out
}
