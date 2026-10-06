// Package cli builds the ncly command tree and runs it.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"

	"github.com/pascalandy/nicely/internal/config"
	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/pascalandy/nicely/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
)

// version is set at link time by release builds:
// -ldflags "-X github.com/pascalandy/nicely/internal/cli.version=vX.Y.Z".
var version string

// globalFlags declares the flags that every command accepts.
var globalFlags = []contract.Flag{
	{Name: "help", Shorthand: "h", Summary: "flag.help"},
	{Name: "json", Summary: "flag.json", Env: "NCLY_JSON"},
	{Name: "no-input", Summary: "flag.no_input", Env: "NCLY_NO_INPUT"},
	{Name: "no-color", Summary: "flag.no_color", Env: "NO_COLOR"},
	{Name: "lang", Summary: "flag.lang", Value: "tag", Env: "NCLY_LANG"},
	{Name: "verbose", Shorthand: "v", Summary: "flag.verbose", Env: "NCLY_VERBOSE"},
	{Name: "version", Summary: "flag.version"},
}

// reservedNames are the names that only core may use, so an extension never
// runs under them, even before core ships their command.
var reservedNames = []string{
	"completion", "describe", "doctor", "auth", "skill", "transcript", "run",
	"agent", "tap", "markdown", "video", "image", "docs",
	"help", "version", "config",
}

// extensionNames lists the domains that installed extensions add. Discovery
// arrives in M3, so it stays empty.
var extensionNames []string

// Main runs ncly with args, without the program name, and returns its exit code.
func Main(args []string, stdout, stderr *os.File) int {
	// Go ends the process with SIGPIPE when the reader of stdout is gone,
	// before ncly can report the lost output. Asking for the signal turns
	// that write into an EPIPE error instead.
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	// The parser stops at its first error, so read the command line first:
	// the language, help, and machine mode must hold even when parsing fails.
	// The help text comes from the catalog, so the tree is built again once
	// the language is known.
	cmd, _, _ := newRoot(i18n.New(language.English)).Find(args)
	line := scan(cmd, args)
	g := resolveGlobals(line, os.LookupEnv, configLang)
	p := i18n.New(i18n.Match(g.lang))
	root := newRoot(p)
	out := &output{file: stdout}
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(stderr)
	root.SetHelpFunc(helpFunc(p, g.noColor))
	cmd, _, _ = root.Find(args)
	var err error
	switch {
	case line.help():
		err = cmd.Help()
	case line.version():
		_, err = fmt.Fprintf(out, "ncly %s\n", releaseVersion())
	case len(line.words) == 0 && len(line.operands) > 0:
		// Operands after -- reach the root, which takes none.
		return report(p, g, contract.Outcome{Errors: []contract.Problem{unknownCommand(p, line.operands[0])}}, stdout, stderr)
	case len(line.words) > 0 && !isBuiltIn(root, line.words[0]) && !isExtension(line.words[0]):
		return report(p, g, contract.Outcome{Errors: []contract.Problem{unknownCommand(p, line.words[0])}}, stdout, stderr)
	default:
		err = root.Execute()
	}
	if out.err != nil {
		return report(p, g, contract.Outcome{Errors: []contract.Problem{writeFailed(p)}}, stdout, stderr)
	}
	if err == nil {
		return 0
	}
	var failed *failure
	if !errors.As(err, &failed) {
		failed = &failure{contract.Outcome{Errors: []contract.Problem{{
			Code:    contract.UsageInvalid,
			Message: p.T("usage.invalid"),
			Hint:    "ncly --help",
		}}}}
	}
	return report(p, g, failed.outcome, stdout, stderr)
}

// output passes writes to stdout and remembers the first that failed, so a
// command whose output was lost never reports success. It keeps the methods
// that terminal detection needs, and no other method of the file, so every
// write goes through Write.
type output struct {
	file *os.File
	err  error
}

func (o *output) Write(b []byte) (int, error) {
	n, err := o.file.Write(b)
	if err != nil && o.err == nil {
		o.err = err
	}
	return n, err
}

func (o *output) Read(b []byte) (int, error) { return o.file.Read(b) }

func (o *output) Close() error { return o.file.Close() }

func (o *output) Fd() uintptr { return o.file.Fd() }

func writeFailed(p *i18n.Printer) contract.Problem {
	return contract.Problem{Code: contract.Runtime, Message: p.T("output.write_failed"), Hint: p.T("output.write_failed_hint")}
}

func newRoot(p *i18n.Printer) *cobra.Command {
	root := &cobra.Command{
		Use:           "ncly",
		Short:         p.T(rootCommand.Summary),
		Example:       strings.Join(rootCommand.Examples, "\n"),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	for _, f := range globalFlags {
		addFlag(root.PersistentFlags(), p, f)
	}
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return &failure{contract.Outcome{Errors: []contract.Problem{flagProblem(p, c, err)}}}
	})
	help := helpCommand(p)
	root.SetHelpCommand(help)
	root.AddCommand(help)
	return root
}

// helpCommand prints the same help as --help, and fails on an unknown topic.
func helpCommand(p *i18n.Printer) *cobra.Command {
	return &cobra.Command{
		Use:     "help [command]",
		Short:   p.T(helpDeclaration.Summary),
		Example: strings.Join(helpDeclaration.Examples, "\n"),
		RunE: func(c *cobra.Command, args []string) error {
			target, rest, err := c.Root().Find(args)
			if err != nil || len(rest) > 0 {
				return &failure{contract.Outcome{Errors: []contract.Problem{unknownCommand(p, args[0])}}}
			}
			return target.Help()
		},
	}
}

// failure carries the outcome of a command that failed.
type failure struct{ outcome contract.Outcome }

func (f *failure) Error() string { return string(f.outcome.Errors[0].Code) }

func isBuiltIn(root *cobra.Command, name string) bool {
	if name == cobra.ShellCompRequestCmd || name == cobra.ShellCompNoDescRequestCmd {
		return true
	}
	return slices.ContainsFunc(root.Commands(), func(c *cobra.Command) bool { return c.Name() == name })
}

// isExtension reports whether an installed extension adds the domain name.
// A reserved name never resolves to an extension.
func isExtension(name string) bool {
	return !slices.Contains(reservedNames, name) && slices.Contains(extensionNames, name)
}

func unknownCommand(p *i18n.Printer, name string) contract.Problem {
	return contract.Problem{
		Code:    contract.UsageInvalid,
		Message: p.T("usage.unknown_command", map[string]any{"Name": name}),
		Hint:    "ncly --help",
	}
}

// report prints an outcome and returns its exit code. An answer that
// cannot be written is a runtime failure, whatever the verdict: when stderr
// itself refuses it, the exit code is all that remains to tell.
func report(p *i18n.Printer, g globals, o contract.Outcome, stdout, stderr io.Writer) int {
	a := contract.Finish(o, summarizer(p))
	var err error
	if g.json {
		err = a.WriteJSON(stdout, stderr)
	} else {
		_, err = fmt.Fprint(tui.Output(stderr, g.noColor), tui.Problems(a.Errors(), a.Warnings(), p.T("error.try")))
	}
	if err != nil {
		return int(contract.ExitRuntime)
	}
	return int(a.Exit)
}

// summaries hold the catalog IDs of the leading error of an outcome that no
// single failure explains, such as an interrupted run or an incomplete batch.
var summaries = map[contract.Code]struct{ message, hint string }{
	contract.Interrupted: {"outcome.interrupted", "outcome.check_hint"},
	contract.Terminated:  {"outcome.terminated", "outcome.check_hint"},
	contract.Temporary:   {"outcome.temporary", "outcome.temporary_hint"},
	contract.Runtime:     {"outcome.runtime", "outcome.check_hint"},
}

func summarizer(p *i18n.Printer) func(contract.Code) contract.Problem {
	return func(code contract.Code) contract.Problem {
		s := summaries[code]
		return contract.Problem{Code: code, Message: p.T(s.message), Hint: p.T(s.hint)}
	}
}

// configLang reads lang from the config files. A broken or missing file
// gives no language, so the help still works with the defaults.
func configLang() string {
	home, _ := os.UserHomeDir()
	cfg, _, err := config.Load(config.Locate(home, os.LookupEnv))
	if err != nil {
		return ""
	}
	return cfg.Lang
}

// releaseVersion prefers the linked version, then the version that Go stamps
// from the module or the VCS, and finally a development marker.
func releaseVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}
	return "v0.0.0-dev"
}
