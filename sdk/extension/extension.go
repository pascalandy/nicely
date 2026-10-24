// Package extension runs a Go extension from one value, as SDK in contract.md defines.
// It starts as ncly-<slug>, or as ncly compiled with the extension and started under that name.
// Either way it parses the call with its
// manifest, runs the handler of the command, and prints the answer by the
// kind of the command's result.
//
// A run that asks for a worker gets every worker of Resources in a private
// temporary folder. The SDK removes that folder once the handler returns,
// so no handler deletes it. After a signal, the folder goes a second before
// cleanup ends at the latest, because core then sends SIGKILL.
package extension

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

// Extension is a Go extension: its manifest, its catalogs, and one handler
// per command that runs.
type Extension struct {
	// Manifest holds manifest.toml.
	Manifest []byte
	// Locales holds locales/<lang>.toml.
	Locales fs.FS
	// Skill holds skill/SKILL.md and the files that it names, which teach
	// an agent the extension's commands. Installing an external extension
	// writes them as ncly-<slug>.skill/. It is nil for an extension that
	// ships no skill.
	Skill fs.FS
	// Resources holds worker/, the workers that the extension embeds, such
	// as a user's script. Call.Worker extracts them for a run. It is nil for
	// an extension that embeds no worker.
	Resources fs.FS
	// Handlers holds one handler per command that runs, keyed by path,
	// such as "hello run".
	Handlers map[string]Handler
}

// Handler runs one validated call. ctx ends on the first Ctrl-C or SIGTERM,
// and the handler then returns what it has.
type Handler func(ctx context.Context, call Call) contract.Outcome

// Call is one validated call of a command.
type Call struct {
	// Path is the command path, such as "hello run".
	Path string
	// Args holds the positional arguments.
	Args []string
	// Flags holds each flag that the caller set, by long name, except a
	// variadic one. A switch holds "true" or "false".
	Flags map[string]string
	// Lists holds each variadic flag that the caller set, by long name,
	// with its values in the caller's order.
	Lists   map[string][]string
	JSON    bool
	Mode    contract.Mode
	Printer *i18n.Printer
	Stdin   io.Reader
	// Stdout takes human text only: the SDK prints the answer. Under --json,
	// it discards what a command other than a text command writes.
	Stdout io.Writer
	Stderr io.Writer
	// domain holds the extension slug for Key's grant hint.
	// services holds the command's requires.keys.
	domain   string
	services []string
	// programs holds the entry of each program that the command declares
	// in requires.bins, which Requirement reads.
	programs map[string]contract.Requirement
	// workers extracts the workers of the run for Worker.
	workers *workers
}

// Main runs the extension as this process with the arguments of the
// process, then exits with its exit code.
func Main(e Extension) {
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	os.Exit(run(e, os.Args[1:], os.Stdout, os.Stderr))
}

// loaded is an extension whose manifest and catalogs were read.
type loaded struct {
	Extension
	manifest contract.Manifest
	catalogs *i18n.Catalogs
}

// load reads the manifest and the catalogs, and checks the handlers, as
// Main does when it starts.
func load(e Extension) (loaded, error) {
	l := loaded{Extension: e}
	var err error
	if l.manifest, err = contract.ParseManifest(e.Manifest, ""); err != nil {
		return l, ruleError(err, "manifest.toml")
	}
	locales, err := fs.Sub(e.Locales, "locales")
	if err != nil {
		return l, err
	}
	own, err := i18n.Load(locales)
	if err != nil {
		return l, fmt.Errorf("catalogs: %w", err)
	}
	if err := l.manifest.CheckCatalog(own.Has); err != nil {
		return l, ruleError(err, "locales/en.toml")
	}
	if l.catalogs, err = i18n.Load(i18n.Locales(), locales); err != nil {
		return l, fmt.Errorf("catalogs: %w", err)
	}
	for _, c := range l.manifest.Commands {
		path := strings.Join(c.Path, " ")
		if l.manifest.Runs(c) && e.Handlers[path] == nil {
			return l, fmt.Errorf("the command %q runs but has no handler", path)
		}
	}
	for path := range e.Handlers {
		if c, found := l.manifest.Command(path); !found || !l.manifest.Runs(c) {
			return l, fmt.Errorf("the handler %q has no command that runs", path)
		}
	}
	return l, nil
}

// ruleError renders a contract rule with its repository file name.
func ruleError(err error, file string) error {
	var rule *contract.RuleError
	if !errors.As(err, &rule) {
		return err
	}
	catalogs, err := i18n.Load(i18n.Locales())
	if err != nil {
		return err
	}
	data := map[string]any{"File": file}
	maps.Copy(data, rule.Data)
	id := rule.Rule
	// load takes the domain from the manifest, so this rule means its name is invalid.
	if id == "extension.domain" {
		id = "extension.domain_name"
	}
	return errors.New(strings.TrimSuffix(catalogs.Printer("en").T(id, data), "."))
}

// Check reports what Main would refuse.
// It also checks the catalog rules of SDK in contract.md and requires a skill's name to equal the slug.
// Every Go extension has a test that calls it.
func Check(e Extension) error {
	l, err := load(e)
	if err != nil {
		return err
	}
	locales, err := fs.Sub(e.Locales, "locales")
	if err != nil {
		return err
	}
	if err := i18n.CheckCatalogs(locales, l.manifest.Slug()); err != nil {
		return err
	}
	return checkSkill(e.Skill, l.manifest.Slug())
}

// checkSkill validates the skill and requires its name to equal the slug.
// An agent can then go from the extension root to its skill without a lookup.
func checkSkill(skill fs.FS, slug string) error {
	if skill == nil {
		return nil
	}
	data, err := fs.ReadFile(skill, "skill/SKILL.md")
	if err != nil {
		return fmt.Errorf("skill: %w", err)
	}
	fm, err := ParseSkill(data)
	if err != nil {
		return fmt.Errorf("skill/SKILL.md: %w", err)
	}
	if fm.Name != slug {
		return fmt.Errorf("skill/SKILL.md: the skill is named %q, not after the slug %q", fm.Name, slug)
	}
	return nil
}

// run runs the extension with args, without the program name, and returns
// its exit code.
func run(e Extension, args []string, stdout, stderr *os.File) int {
	l, err := load(e)
	if err != nil {
		return broken(err, args, stdout, stderr)
	}
	path := strings.Join(l.manifest.Root(), " ")
	// The tree has ncly above the extension root, so that hints and help
	// read as they do through core.
	args = append(l.manifest.Root(), args...)
	en := l.catalogs.Printer("en")
	cmd, _, _ := l.tree(en, nil).Find(args)
	line := command.Scan(cmd, args)
	g := command.Resolve(line)
	p := l.catalogs.Printer(g.Lang)
	var selected *selection
	root := l.tree(p, func(s selection) error {
		selected = &s
		return nil
	})
	out := command.NewOutput(stdout)
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(stderr)
	root.SetHelpFunc(command.HelpFunc(p, g.NoColor))
	codes := l.manifest.Registry()
	form := command.NewForm(p, g, os.Stdin, stdout, stderr)
	switch {
	case line.Version() && !line.Help():
		_, err = fmt.Fprintf(out, "ncly-%s %s\n", l.manifest.Slug(), command.ModuleVersion())
	case line.Help():
		target, _, _ := root.Find(args)
		err = target.Help()
	default:
		err = root.ExecuteContext(command.WithForm(context.Background(), form))
	}
	code := 0
	if selected != nil && err == nil {
		code = l.handle(*selected, p, g, codes, out, stdout, stderr)
	} else {
		code = command.Finish(p, g, codes, out, err, "ncly "+path+" --help", stdout, stderr)
	}
	if err := form.PrintNext(code); err != nil {
		return command.Report(p, g, codes, command.WriteFailure(p), stdout, stderr)
	}
	return code
}

// selection is the command and the call that the parser accepted.
type selection struct {
	command contract.Command
	call    Call
}

// tree builds the command tree of the manifest under a root named ncly.
// Each command that runs reports its call to selected, when it is set.
func (l loaded) tree(p *i18n.Printer, selected func(selection) error) *cobra.Command {
	root := &cobra.Command{Use: "ncly", SilenceErrors: true, SilenceUsage: true}
	root.CompletionOptions.DisableDefaultCmd = true
	command.AddGlobalFlags(root, p)
	command.AddManifest(root, p, p, l.manifest, func(d contract.Command) func(*cobra.Command, []string) error {
		if selected == nil {
			return nil
		}
		return func(c *cobra.Command, args []string) error {
			flags, lists := map[string]string{}, map[string][]string{}
			for _, f := range append(slices.Clone(d.Flags), d.ModeFlags()...) {
				switch flag := c.Flags().Lookup(f.Name); {
				case !flag.Changed:
				case f.Variadic:
					lists[f.Name] = command.FlagValues(flag)
				default:
					flags[f.Name] = flag.Value.String()
				}
			}
			path := strings.Join(d.Path, " ")
			return selected(selection{command: d, call: Call{Path: path, Args: args, Flags: flags, Lists: lists}})
		}
	})
	return root
}

// broken answers RUNTIME for an extension that does not match its own
// manifest, with the SDK's text alone. With --verbose, the reason comes
// first on stderr.
func broken(reason error, args []string, stdout, stderr *os.File) int {
	sdk, err := i18n.Load(i18n.Locales())
	if err != nil {
		return int(contract.ExitRuntime)
	}
	root := &cobra.Command{Use: "ncly"}
	command.AddGlobalFlags(root, sdk.Printer("en"))
	g := command.Resolve(command.Scan(root, args))
	p := sdk.Printer(g.Lang)
	if g.Verbose {
		_, _ = fmt.Fprintln(stderr, reason)
	}
	problem := contract.Problem{Code: contract.Runtime, Message: p.T("outcome.extension_broken"), Hint: p.T("outcome.extension_broken_hint")}
	return command.Report(p, g, contract.Core, contract.Outcome{Errors: []contract.Problem{problem}}, stdout, stderr)
}
