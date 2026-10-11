// Package extension runs a Go extension from one value, as SDK in
// contract.md defines: started as ncly-<domain>, or compiled into ncly and
// started by it under that name. Either way it parses the call with its
// manifest, runs the handler of the command, and prints the answer by the
// kind of the command's result.
package extension

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Extension is a Go extension: its manifest, its catalogs, and one handler
// per command that runs.
type Extension struct {
	// Manifest holds manifest.toml.
	Manifest []byte
	// Locales holds locales/<lang>.toml.
	Locales fs.FS
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
	// Flags holds each flag that the caller set, by long name. A switch
	// holds "true" or "false".
	Flags   map[string]string
	JSON    bool
	Mode    contract.Mode
	Printer *i18n.Printer
	Stdin   io.Reader
	// Stdout takes human text only: the SDK prints the answer. Under --json,
	// it discards what a command other than a text command writes.
	Stdout io.Writer
	Stderr io.Writer
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
	var head struct {
		Domain string `toml:"domain"`
	}
	if _, err := toml.Decode(string(e.Manifest), &head); err != nil {
		return l, fmt.Errorf("manifest: %w", err)
	}
	var err error
	if l.manifest, err = contract.ParseManifest(e.Manifest, head.Domain); err != nil {
		return l, fmt.Errorf("manifest: %w", err)
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
		return l, fmt.Errorf("catalogs: %w", err)
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

// Check reports what Main would refuse, and a catalog that breaks the
// catalog rules of SDK in contract.md. Every Go extension has a test that
// calls it.
func Check(e Extension) error {
	l, err := load(e)
	if err != nil {
		return err
	}
	locales, err := fs.Sub(e.Locales, "locales")
	if err != nil {
		return err
	}
	return i18n.CheckCatalogs(locales, l.manifest.Domain)
}

// run runs the extension with args, without the program name, and returns
// its exit code.
func run(e Extension, args []string, stdout, stderr *os.File) int {
	l, err := load(e)
	if err != nil {
		return broken(err, args, stdout, stderr)
	}
	domain := l.manifest.Domain
	// The tree has the root ncly above the domain, so that hints and help
	// read as they do through core.
	args = append([]string{domain}, args...)
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
	switch {
	case line.Version() && !line.Help():
		_, err = fmt.Fprintf(out, "ncly-%s %s\n", domain, command.ModuleVersion())
	case line.Help():
		target, _, _ := root.Find(args)
		err = target.Help()
	default:
		err = root.Execute()
	}
	if selected != nil && err == nil {
		return l.handle(*selected, p, g, codes, out, stdout, stderr)
	}
	return command.Finish(p, g, codes, out, err, "ncly "+domain+" --help", stdout, stderr)
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
			flags := map[string]string{}
			for _, f := range d.Flags {
				if flag := c.Flags().Lookup(f.Name); flag.Changed {
					flags[f.Name] = flag.Value.String()
				}
			}
			path := strings.Join(d.Path, " ")
			return selected(selection{command: d, call: Call{Path: path, Args: args, Flags: flags}})
		}
	})
	return root
}

// handle runs the handler of a validated call and prints its result by the
// kind of the command's result. The first Ctrl-C or SIGTERM ends the
// handler's context and decides the verdict. A handler that has not
// returned once cleanup and the drain limit have passed gets no more time.
func (l loaded) handle(s selection, p *i18n.Printer, g command.Globals, codes contract.Registry, out *command.Output, stdout, stderr *os.File) int {
	kind := s.command.Result.Kind
	call := s.call
	call.JSON, call.Printer = g.JSON, p
	call.Mode = contract.DetectMode(term.IsTerminal(int(os.Stdin.Fd())), term.IsTerminal(int(stdout.Fd())), g.NoInput, g.JSON, os.LookupEnv)
	call.Stdin, call.Stdout, call.Stderr = os.Stdin, io.Writer(out), io.Writer(stderr)
	if g.JSON && kind != contract.ResultText {
		call.Stdout = io.Discard
	}
	// Run directly, the extension resolves the global flags itself, so that
	// its handler and the handler's programs see the variables that ncly
	// would set.
	for name, value := range g.Vars() {
		if value == "" {
			_ = os.Unsetenv(name)
		} else {
			_ = os.Setenv(name, value)
		}
	}
	ctx := runner.OnSignal()
	done := make(chan contract.Outcome, 1)
	go func() { done <- l.Handlers[call.Path](ctx, call) }()
	var o contract.Outcome
	select {
	case o = <-done:
	case <-ctx.Done():
		// The runner ends the handler's programs with SIGKILL when cleanup
		// ends, then reads their streams for the drain limit, so the handler
		// gets that long to return what it has.
		select {
		case o = <-done:
		case <-time.After(runner.Cleanup + runner.DrainLimit):
			o = contract.Outcome{}
		}
	}
	if ctx.Err() != nil {
		o.Signal = contract.StopSignal(runner.SignalOf(ctx))
	}
	if out.Err() != nil {
		return command.Report(p, g, codes, command.WriteFailure(p), stdout, stderr)
	}
	if kind == contract.ResultText && codes.Decide(o).Exit == contract.ExitOK {
		// A text result prints its text, never an answer, so its warnings
		// print as text too.
		g.JSON = false
	}
	return command.Report(p, g, codes, o, stdout, stderr)
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
