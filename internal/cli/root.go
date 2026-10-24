// Package cli builds the ncly command tree and runs it.
package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	sdkextension "github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
)

// version is set at link time by release builds:
// -ldflags "-X github.com/pascalandy/nicely/internal/cli.version=vX.Y.Z".
var version string

var rootCommand = contract.Command{
	Summary: "root.summary",
	Examples: []string{
		`ncly skill view nicely`,
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
	"help", "version", "config", "core", "extensions",
}

// Main runs ncly with args, without the program name, and returns its exit
// code. bundled lists the extensions compiled into ncly, and keys is the
// keychain's backend.
func Main(args []string, bundled []sdkextension.Extension, keys keychain.Backend, stdout, stderr *os.File) int {
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
	en := catalogs.Printer("en")
	core := newRoot(en, command.Globals{}, nil, keys)
	cmd, _, _ := core.Find(args)
	extensions := neededExtensions(newHost(bundled), core, command.Scan(cmd, args).Words)
	cmd, _, _ = newRoot(en, command.Globals{}, extensions, keys).Find(args)
	line := command.Scan(cmd, args)
	g := command.Resolve(line)
	p := catalogs.Printer(g.Lang)
	root := newRoot(p, g, extensions, keys)
	out := command.NewOutput(stdout)
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(stderr)
	root.SetHelpFunc(command.HelpFunc(p, g.NoColor))
	// A completion request from the shell scripts answers in Cobra's protocol:
	// the words it completes never ask for help or name a command to check.
	// Global flags may come first, as an alias such as ncly='ncly --no-input'
	// puts them there, so the request is the first word after the flags.
	completing := len(line.Words) > 0 && isCompletionRequest(line.Words[0])
	broken := brokenTarget(extensions, line.Words)
	form := command.NewForm(p, g, os.Stdin, stdout, stderr)
	var err error
	switch {
	case completing:
		err = root.Execute()
	case line.Version() && !line.Help():
		_, err = fmt.Fprintf(out, "ncly %s\n", releaseVersion())
	case broken != nil:
		// An extension that cannot run has no help either.
		err = command.Fail(broken.invalid(p))
	case line.Help():
		target, _, _ := root.Find(args)
		err = target.Help()
	case len(line.Words) == 0 && len(line.Operands) > 0:
		// Operands after -- reach the root, which takes none.
		err = command.Fail(command.UnknownCommand(p, root, line.Operands[0]))
	case len(line.Words) > 0 && !isCommand(root, line.Words[0]):
		err = command.Fail(command.UnknownCommand(p, root, line.Words[0]))
	default:
		err = root.ExecuteContext(command.WithForm(context.Background(), form))
	}
	var (
		extensionCall *call
		done          *answered
		code          int
	)
	switch {
	case errors.As(err, &extensionCall):
		code = extensionCall.run(p, g, keys, stdout, stderr)
	case errors.As(err, &done) && out.Err() == nil:
		code = done.print(p, g, out, stderr)
	default:
		code = command.Finish(p, g, contract.Core, out, err, "ncly --help", stdout, stderr)
	}
	if err := form.PrintNext(code); err != nil {
		return command.Report(p, g, contract.Core, command.WriteFailure(p), stdout, stderr)
	}
	return code
}

// newRoot builds the command tree, with the commands of each valid
// extension in the language that g requests. Completion passes g on to the
// list commands that it runs. keys is the keychain's backend.
func newRoot(p *i18n.Printer, g command.Globals, extensions []*extension, keys keychain.Backend) *cobra.Command {
	root := &cobra.Command{
		Use:           "ncly",
		Short:         p.T(rootCommand.Summary),
		Example:       strings.Join(rootCommand.Examples, "\n"),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	command.AddGlobalFlags(root, p)
	help := helpCommand(p)
	root.SetHelpCommand(help)
	root.AddCommand(help)
	addCompletion(p, root)
	addDescribe(p, g.Lang, root, extensions)
	addDoctor(p, g, root, extensions, keys)
	addAuth(p, g, root, extensions, keys)
	addRun(p, root)
	addExtensions(root, p, g, extensions)
	return root
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

// neededExtensions loads declarations for the command line.
// Direct extension calls need only their selected root; a domain group's help needs its siblings.
// Describe, doctor, auth, root help, completion, and unknown commands need the installed inventory.
// Other core commands need no extension declarations.
func neededExtensions(h host, core *cobra.Command, words []string) []*extension {
	name := target(words)
	if name != "" && name != describeCommand.Path[0] && name != doctorCommand.Path[0] && name != authCommand.Path[0] && !isCompletionRequest(name) {
		if isCommand(core, name) {
			return nil
		}
		return h.forPath(targetPath(words))
	}
	return h.all()
}

// brokenTarget returns the extension that cannot run whose root the
// command line targets.
func brokenTarget(extensions []*extension, words []string) *extension {
	return brokenPath(extensions, targetPath(words))
}

// releaseVersion prefers the linked version, then the version that Go stamps
// from the module or the VCS, and finally a development marker.
func releaseVersion() string { return cmp.Or(version, command.ModuleVersion()) }
