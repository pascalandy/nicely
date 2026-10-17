package cli

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/tui"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

//go:embed describe_schema.json
var describeSchema []byte

var describeCommand = contract.Command{
	Path:    []string{"describe"},
	Result:  contract.Result{Kind: contract.ResultObject, Schema: mustSchema(describeSchema)},
	Summary: "describe.summary",
	Args:    []contract.Arg{{Name: "command", Summary: "describe.command", Variadic: true}},
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly describe`,
		`ncly describe completion --json`,
		`ncly describe 'completion zsh' --json`,
	},
}

// Sources of an entry, as ncly describe in core-spec.md defines them.
const (
	sourceCore      = "core"
	sourceBundled   = "bundled"
	sourceExtension = "extension"
)

// entry is the concise description of one command, with its declaration.
type entry struct {
	Path    string `json:"path"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
	// Skill is present on the entry of an extension's domain alone, where it
	// holds the folder of the extension's skill, or null as a nil *string.
	Skill any `json:"skill,omitempty"`
	// declaration is the command's declaration, whose catalog IDs text
	// renders. A group only groups the commands under it, so describe lists
	// them instead of detailing it.
	declaration contract.Command
	text        *i18n.Printer
	group       bool
	// requirements holds the entries of the manifest's programs table.
	requirements map[string]contract.Requirement
}

// listing is the answer of ncly describe for no path or a group.
type listing struct {
	Commands []entry `json:"commands"`
}

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
	Keys []named     `json:"keys"`
}

// binDetail is a program that a command may run, with the keys of its
// entry in the manifest's programs table, each present only when declared.
type binDetail struct {
	Name string `json:"name"`
	contract.Requirement
}

type named struct {
	Name string `json:"name"`
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
		Requires:   requiresDetail{Bins: []binDetail{}, Keys: names(d.Requires.Keys)},
		Modes:      modesDetail(d.Modes),
		Config:     []config.Value{lang},
		Examples:   append([]string{}, d.Examples...),
	}
	for _, name := range d.Requires.Bins {
		out.Requires.Bins = append(out.Requires.Bins, binDetail{name, e.requirements[name]})
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

// mustSchema reads a schema of core, whose references it resolves.
func mustSchema(data []byte) map[string]any {
	var schema map[string]any
	err := json.Unmarshal(data, &schema)
	if err == nil {
		schema, err = contract.ResolveSchema(schema)
	}
	if err == nil && !contract.IsDataSchema(schema) {
		err = errors.New("describe: not a data schema")
	}
	if err != nil {
		panic(err)
	}
	return schema
}

func names(list []string) []named {
	out := []named{}
	for _, name := range list {
		out = append(out, named{name})
	}
	return out
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

// addDescribe adds ncly describe, which reads the declarations of core and
// the manifests of extensions, and never starts an extension.
func addDescribe(p *i18n.Printer, lang string, root *cobra.Command, extensions []*extension) {
	root.AddCommand(command.Build(p, p, describeCommand, func(c *cobra.Command, args []string) error {
		// Discovery works with the defaults when a config file cannot be
		// read, and says so in every answer, a failure's included.
		values, err := config.LoadValues(config.Locate(home(), os.LookupEnv))
		var warnings []contract.Problem
		if err != nil {
			values, warnings = nil, []contract.Problem{defaultsUsed(p, err)}
		}
		fail := func(problem contract.Problem) error {
			return &command.Failure{Outcome: contract.Outcome{Errors: []contract.Problem{problem}, Warnings: warnings}}
		}
		words := strings.Fields(strings.Join(args, " "))
		if len(args) > 0 && len(words) == 0 {
			return fail(notFound(p, ""))
		}
		if len(words) > 0 {
			if e := brokenDomain(extensions, words[0]); e != nil {
				return fail(e.invalid(p))
			}
		}
		var found []entry
		for _, e := range entries(p, lang, extensions) {
			if within(strings.Fields(e.Path), words) {
				found = append(found, e)
			}
		}
		if len(found) == 0 {
			return fail(notFound(p, strings.Join(words, " ")))
		}
		// The path sorts first, so a path that names a command, not a group,
		// gets its detail alone.
		var answer any = listing{found}
		if len(words) > 0 && !found[0].group && found[0].Path == strings.Join(words, " ") {
			found = found[:1]
			lang := config.Lang(c.Flag("lang").Value.String(), os.Getenv, func() config.Values { return values })
			answer = commandDetail{[]detail{found[0].detail(p, lang, values)}}
		}
		return &answered{
			outcome: contract.Outcome{Data: answer, Warnings: warnings},
			text:    func(w io.Writer) error { return writeEntries(w, found) },
		}
	}))
}

// entries returns the concise entry of every command that runs here, in
// the order of their paths: core's, then each valid extension's.
func entries(p *i18n.Printer, lang string, extensions []*extension) []entry {
	var all []entry
	core := coreCommands()
	for _, d := range core {
		// As in a manifest, a command with children and no arguments groups them.
		children := slices.ContainsFunc(core, func(other contract.Command) bool {
			return len(other.Path) > len(d.Path) && slices.Equal(other.Path[:len(d.Path)], d.Path)
		})
		all = append(all, entry{
			Path: strings.Join(d.Path, " "), Summary: p.T(d.Summary), Source: sourceCore,
			declaration: d, text: p, group: children && len(d.Args) == 0,
		})
	}
	for _, e := range extensions {
		if e.broken != nil {
			continue
		}
		text := e.catalogs.Printer(lang)
		for _, d := range e.manifest.Commands {
			en := entry{
				Path: strings.Join(d.Path, " "), Summary: text.T(d.Summary), Source: sourceExtension,
				declaration: d, text: text, group: !e.manifest.Runs(d), requirements: e.manifest.Requirements,
			}
			if e.bundled {
				en.Source = sourceBundled
			}
			if len(d.Path) == 1 {
				en.Skill = e.skill()
			}
			all = append(all, en)
		}
	}
	slices.SortFunc(all, func(a, b entry) int { return slices.Compare(strings.Fields(a.Path), strings.Fields(b.Path)) })
	return all
}

// coreCommands returns the declarations of core's commands, which
// discovery lists. help only prints the help of another command, so it is
// left out, as ncly --help leaves it out.
func coreCommands() []contract.Command {
	commands := []contract.Command{completionCommand, describeCommand, doctorCommand, authCommand, authStatusCommand, authLoginCommand, authLogoutCommand}
	for _, s := range shells {
		commands = append(commands, s.command)
	}
	return commands
}

// skill returns the folder ncly-<domain>.skill/ beside the manifest of an
// external extension, or a nil *string when it ships none or is bundled.
func (e *extension) skill() *string {
	if e.bundled {
		return nil
	}
	folder := e.path + ".skill"
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return nil
	}
	return &folder
}

// within reports whether path is the path of words or lies under it.
func within(path, words []string) bool {
	return len(path) >= len(words) && slices.Equal(path[:len(words)], words)
}

// notFound reports a path that names no command, with the command that
// lists them as next.
func notFound(p *i18n.Printer, path string) contract.Problem {
	return contract.Problem{
		Code:    contract.NotFound,
		Message: p.T("describe.not_found", map[string]any{"Path": path}),
		Hint:    "ncly describe",
		Next:    []string{"ncly", "describe"},
	}
}

// writeEntries prints one line per entry: its path, then its summary.
func writeEntries(w io.Writer, entries []entry) error {
	width := 0
	for _, e := range entries {
		width = max(width, len(e.Path))
	}
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%-*s  %s\n", width, e.Path, e.Summary)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// defaultsUsed warns that discovery used the defaults because a config
// file cannot be read, naming the file and, when the parser knows it, the
// line.
func defaultsUsed(p *i18n.Printer, err error) contract.Problem {
	file, line := "", 0
	var invalid *config.InvalidError
	if errors.As(err, &invalid) {
		file, line = invalid.File, invalid.Line
	}
	data := map[string]any{"File": file, "Line": line}
	message := p.T("describe.defaults_used", data)
	if line > 0 {
		message = p.T("describe.defaults_used_line", data)
	}
	return contract.Problem{Code: contract.ConfigDefaultsUsed, Message: message, Hint: p.T("describe.defaults_used_hint", data)}
}

// home is the home folder, or empty when it is unknown, which leaves the
// default config files unset.
func home() string {
	dir, _ := os.UserHomeDir()
	return dir
}

// answered is the outcome of a core command that succeeded, with the text
// that it prints without --json. The command returns it as its error, so
// that Main prints it with the resolved global flags.
type answered struct {
	outcome contract.Outcome
	text    func(io.Writer) error
}

func (a *answered) Error() string { return "answered" }

// print writes the text without --json, then reports the outcome: its
// answer with --json, and its warnings either way.
func (a *answered) print(p *i18n.Printer, g command.Globals, out *command.Output, stderr io.Writer) int {
	if !g.JSON {
		if err := a.text(tui.Output(out, g.NoColor)); err != nil || out.Err() != nil {
			return command.Report(p, g, contract.Core, command.WriteFailure(p), out, stderr)
		}
	}
	return command.Report(p, g, contract.Core, a.outcome, out, stderr)
}
