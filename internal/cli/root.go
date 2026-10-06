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

var rootCommand = contract.Command{
	Summary: "root.summary",
	Examples: []string{
		`ncly completion zsh > "${fpath[1]}/_ncly"`,
		`ncly --version`,
		`NCLY_JSON=1 ncly help`,
	},
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
	g := resolveGlobals(line)
	p := i18n.New(i18n.Match(g.lang))
	root := newRoot(p)
	out := &output{file: stdout}
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(stderr)
	root.SetHelpFunc(helpFunc(p, g.noColor))
	// A completion request from the shell scripts answers in Cobra's protocol:
	// the words it completes never ask for help or name a command to check.
	// Global flags may come first, as an alias such as ncly='ncly --no-input'
	// puts them there, so the request is the first word after the flags.
	completing := len(line.words) > 0 && (line.words[0] == cobra.ShellCompRequestCmd || line.words[0] == cobra.ShellCompNoDescRequestCmd)
	var err error
	switch {
	case completing:
		err = root.Execute()
	case line.help():
		target, _, _ := root.Find(args)
		err = target.Help()
	case line.version():
		_, err = fmt.Fprintf(out, "ncly %s\n", releaseVersion())
	case len(line.words) == 0 && len(line.operands) > 0:
		// Operands after -- reach the root, which takes none.
		err = fail(unknownCommand(p, root, line.operands[0]))
	case len(line.words) > 0 && !isBuiltIn(root, line.words[0]) && !isExtension(line.words[0]):
		err = fail(unknownCommand(p, root, line.words[0]))
	default:
		err = root.Execute()
	}
	if out.err != nil {
		err = fail(contract.Problem{Code: contract.Runtime, Message: p.T("output.write_failed"), Hint: p.T("output.write_failed_hint")})
	}
	if err == nil {
		return 0
	}
	var failed *failure
	if !errors.As(err, &failed) {
		failed = fail(contract.Problem{Code: contract.UsageInvalid, Message: p.T("usage.invalid"), Hint: "ncly --help"})
	}
	return report(p, g, failed.outcome, stdout, stderr)
}

// output records the first failed write, so lost output never reports
// success. Beyond Write, it keeps only Fd, for terminal detection, so that no
// other method of the file can write around Write.
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

func (o *output) Fd() uintptr { return o.file.Fd() }

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
	// Cobra offers file names for a flag value, and a language tag is never one.
	_ = root.RegisterFlagCompletionFunc("lang", cobra.NoFileCompletions)
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fail(flagProblem(p, c, err))
	})
	help := helpCommand(p)
	root.SetHelpCommand(help)
	root.AddCommand(help)
	addCompletion(p, root)
	return root
}

// failure carries the outcome of a command that failed.
type failure struct{ outcome contract.Outcome }

func (f *failure) Error() string { return string(f.outcome.Errors[0].Code) }

func fail(problems ...contract.Problem) *failure {
	return &failure{contract.Outcome{Errors: problems}}
}

func isBuiltIn(root *cobra.Command, name string) bool {
	return slices.ContainsFunc(root.Commands(), func(c *cobra.Command) bool { return c.Name() == name })
}

// isExtension reports whether an installed extension adds the domain name.
// A reserved name never resolves to an extension.
func isExtension(name string) bool {
	return !slices.Contains(reservedNames, name) && slices.Contains(extensionNames, name)
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
		label := func(code contract.Code, warning bool) string {
			if warning {
				return p.T("warning.label", map[string]any{"Code": code})
			}
			return p.T("error.label", map[string]any{"Code": code})
		}
		_, err = fmt.Fprint(tui.Output(stderr, g.noColor), tui.Problems(a.Errors(), a.Warnings(), label, p.T("error.hint")))
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

// configLang reads lang from the config files. A load error gives no language,
// so the help still works with the defaults.
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
