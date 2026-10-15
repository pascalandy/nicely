package cli

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

var describeCommand = contract.Command{
	Path:    []string{"describe"},
	Summary: "describe.summary",
	Args:    []contract.Arg{{Name: "command", Summary: "describe.command", Variadic: true}},
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

// entry is the concise description of one command.
type entry struct {
	Path    string `json:"path"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
	// Skill is present on the entry of an extension's domain alone, where it
	// holds the folder of the extension's skill, or null as a nil *string.
	Skill any `json:"skill,omitempty"`
}

// listing is the answer of ncly describe.
type listing struct {
	Commands []entry `json:"commands"`
}

// addDescribe adds ncly describe, which reads the declarations of core and
// the manifests of extensions, and never starts an extension.
func addDescribe(p *i18n.Printer, lang string, root *cobra.Command, extensions []*extension) {
	root.AddCommand(command.Build(p, p, describeCommand, func(c *cobra.Command, args []string) error {
		words := strings.Fields(strings.Join(args, " "))
		if len(args) > 0 && len(words) == 0 {
			return command.Fail(notFound(p, ""))
		}
		if len(words) > 0 {
			if e := brokenDomain(extensions, words[0]); e != nil {
				return command.Fail(e.invalid(p))
			}
		}
		var found []entry
		for _, e := range entries(p, lang, extensions) {
			if within(strings.Fields(e.Path), words) {
				found = append(found, e)
			}
		}
		if len(found) == 0 {
			return command.Fail(notFound(p, strings.Join(words, " ")))
		}
		return &answered{
			outcome: contract.Outcome{Data: listing{found}},
			text:    func(w io.Writer) error { return writeEntries(w, found) },
		}
	}))
}

// entries returns the concise entry of every command that runs here, in
// the order of their paths: core's, then each valid extension's.
func entries(p *i18n.Printer, lang string, extensions []*extension) []entry {
	var all []entry
	for _, d := range coreCommands() {
		all = append(all, entry{Path: strings.Join(d.Path, " "), Summary: p.T(d.Summary), Source: sourceCore})
	}
	for _, e := range extensions {
		if e.broken != nil {
			continue
		}
		text := e.catalogs.Printer(lang)
		for _, d := range e.manifest.Commands {
			en := entry{Path: strings.Join(d.Path, " "), Summary: text.T(d.Summary), Source: sourceExtension}
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
	commands := []contract.Command{completionCommand, describeCommand}
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
		if err := a.text(out); err != nil || out.Err() != nil {
			return command.Report(p, g, contract.Core, command.WriteFailure(p), out, stderr)
		}
	}
	return command.Report(p, g, contract.Core, a.outcome, out, stderr)
}
