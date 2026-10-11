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

var rootCommand = contract.Command{
	Summary: "root.summary",
	Examples: []string{
		`ncly completion zsh > "${fpath[1]}/_ncly"`,
		`ncly --version`,
		`NCLY_JSON=1 ncly help`,
	},
}

// reservedNames are the names of core's commands and of the bundled
// extensions, so an external extension never runs under them, even before
// core ships their command.
var reservedNames = []string{
	"completion", "describe", "doctor", "auth", "run", "tap", "skill",
	"help", "version", "config",
}

// Main runs ncly with args, without the program name, and returns its exit code.
func Main(args []string, stdout, stderr *os.File) int {
	// Go ends the process with SIGPIPE when the reader of stdout is gone,
	// before ncly can report the lost output. Asking for the signal turns
	// that write into an EPIPE error instead. signal.Ignore would do the
	// same, but every program that ncly runs would inherit the ignored signal.
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	// The parser stops at its first error, so read the command line first:
	// the language, help, and machine mode must hold even when parsing fails.
	// The first reading, with core's commands only, finds the extensions that
	// the line needs, whose flags the second reading knows. The help text
	// comes from the catalogs, so the tree is built again once the language
	// is known.
	en := i18n.New(language.English)
	core := newRoot(en, "", nil)
	cmd, _, _ := core.Find(args)
	extensions := neededExtensions(searchPath(), core, scan(cmd, args).words)
	cmd, _, _ = newRoot(en, "", extensions).Find(args)
	line := scan(cmd, args)
	g := resolveGlobals(line)
	p := i18n.New(i18n.Match(g.lang))
	root := newRoot(p, g.lang, extensions)
	out := &output{file: stdout}
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(stderr)
	root.SetHelpFunc(helpFunc(p, g.noColor))
	// A completion request from the shell scripts answers in Cobra's protocol:
	// the words it completes never ask for help or name a command to check.
	// Global flags may come first, as an alias such as ncly='ncly --no-input'
	// puts them there, so the request is the first word after the flags.
	completing := len(line.words) > 0 && isCompletionRequest(line.words[0])
	broken := brokenTarget(extensions, line.words)
	var err error
	switch {
	case completing:
		err = root.Execute()
	case line.version() && !line.help():
		_, err = fmt.Fprintf(out, "ncly %s\n", releaseVersion())
	case broken != nil:
		// An extension that cannot run has no help either.
		err = fail(broken.invalid(p))
	case line.help():
		target, _, _ := root.Find(args)
		err = target.Help()
	case len(line.words) == 0 && len(line.operands) > 0:
		// Operands after -- reach the root, which takes none.
		err = fail(unknownCommand(p, root, line.operands[0]))
	case len(line.words) > 0 && !isCommand(root, line.words[0]):
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

// newRoot builds the command tree, with the commands of each valid
// extension in the language that lang requests.
func newRoot(p *i18n.Printer, lang string, extensions []*extension) *cobra.Command {
	root := &cobra.Command{
		Use:           "ncly",
		Short:         p.T(rootCommand.Summary),
		Example:       strings.Join(rootCommand.Examples, "\n"),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	for _, f := range contract.GlobalFlags {
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
	addExtensions(root, p, lang, extensions)
	return root
}

// failure carries the outcome of a command that failed.
type failure struct{ outcome contract.Outcome }

func (f *failure) Error() string { return string(f.outcome.Errors[0].Code) }

func fail(problems ...contract.Problem) *failure {
	return &failure{contract.Outcome{Errors: problems}}
}

// isCommand reports whether the root has a command of that name: one of
// core's, or the domain of a valid extension.
func isCommand(root *cobra.Command, name string) bool {
	return slices.ContainsFunc(root.Commands(), func(c *cobra.Command) bool { return c.Name() == name })
}

// isCompletionRequest reports whether a word asks for completions in Cobra's
// protocol, as the shell scripts do.
func isCompletionRequest(word string) bool {
	return word == cobra.ShellCompRequestCmd || word == cobra.ShellCompNoDescRequestCmd
}

// target returns the first word of a command line, or the word after help,
// which names the command that the line runs or asks help about.
func target(words []string) string {
	if len(words) > 0 && words[0] == "help" {
		words = words[1:]
	}
	if len(words) == 0 {
		return ""
	}
	return words[0]
}

// neededExtensions finds the extensions that a command line needs: none
// for a command of core, the one whose domain it targets, and every one
// otherwise, for the root's help, for completion, and for a word that names
// no command, whose help is the root's.
func neededExtensions(dirs []string, core *cobra.Command, words []string) []*extension {
	name := target(words)
	if name != "" && !isCompletionRequest(name) {
		if isCommand(core, name) {
			return nil
		}
		if e := findExtension(dirs, name); e != nil {
			return []*extension{e}
		}
	}
	return findExtensions(dirs)
}

// brokenTarget returns the extension that cannot run whose domain the
// command line targets.
func brokenTarget(extensions []*extension, words []string) *extension {
	name := target(words)
	for _, e := range extensions {
		if e.domain == name && e.broken != nil {
			return e
		}
	}
	return nil
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
