package cli

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
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
	// Skill is present on the entry of an extension root, where it
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
	// keys holds the entries of the manifest's keys table.
	keys map[string]contract.KeyEntry
}

// listing is the answer of ncly describe for no path or a group.
type listing struct {
	Commands []entry `json:"commands"`
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
			if e := brokenPath(extensions, words); e != nil {
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
	all := extensionEntries(p, lang, extensions)
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
	slices.SortFunc(all, func(a, b entry) int { return slices.Compare(strings.Fields(a.Path), strings.Fields(b.Path)) })
	return all
}

// coreCommands returns the declarations of core's commands, which
// discovery lists. help only prints the help of another command, so it is
// left out, as ncly --help leaves it out.
func coreCommands() []contract.Command {
	commands := []contract.Command{completionCommand, describeCommand, doctorCommand, authCommand, authStatusCommand, authLoginCommand, authLogoutCommand, authGrantCommand, authRevokeCommand, runCommand, runListCommand, runViewCommand}
	for _, s := range shells {
		commands = append(commands, s.command)
	}
	return commands
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
