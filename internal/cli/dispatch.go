package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
)

// call is a call to an extension that the parser accepted. The command
// returns it as its error, so that Main starts the extension with what it
// resolved.
type call struct {
	extension *extension
	kind      contract.ResultKind
	args      []string
	// services are the services that the command declares in
	// requires.keys, and readKeys reports whether core reads their keys.
	services []string
	readKeys bool
}

func (c *call) Error() string { return "call to " + c.extension.slug }

// canonicalArgs passes a validated call as Dispatch in contract.md defines.
// It removes the extension root, then adds caller-set flags in manifest order.
// A variadic flag passes one word per value.
// Caller-set --dry-run and --run-id follow, then -- and the positional arguments.
func canonicalArgs(m contract.Manifest, d contract.Command, c *cobra.Command, args []string) []string {
	canonical := slices.Clone(d.Path[len(m.Root()):])
	for _, f := range append(slices.Clone(d.Flags), d.ModeFlags()...) {
		if flag := c.Flags().Lookup(f.Name); flag.Changed {
			for _, value := range command.FlagValues(flag) {
				canonical = append(canonical, "--"+f.Name+"="+value)
			}
		}
	}
	return append(append(canonical, "--"), args...)
}

// run starts the extension with the environment of Programs that ncly runs
// in contract.md, with the keys that a human granted to it, which it reads
// through b, and returns the exit code of ncly. Without --json, and for a
// text result, the extension writes to the streams of ncly, and core
// checks only its exit code. Otherwise core checks its answer first. When
// Ctrl-C or SIGTERM arrives before the extension exits, the first signal
// decides the verdict.
func (c *call) run(p *i18n.Printer, g command.Globals, b keychain.Backend, stdout, stderr *os.File) int {
	ctx := runner.OnSignal()
	vars := extensionVars(g)
	keys, err := keyVars(ctx, keychain.New(b, interactive(g, stdout)), c.extension, c.services, c.readKeys)
	if err != nil {
		// A signal ended a keychain read before the extension started, which
		// a read leaves running.
		signal := runner.SignalOf(ctx)
		runner.Stop(signal)
		return command.Report(p, g, contract.Core, contract.Outcome{Signal: contract.StopSignal(signal)}, stdout, stderr)
	}
	maps.Copy(vars, keys)
	program := runner.Program{
		Path:     c.extension.path,
		Args:     c.args,
		Env:      runner.Environ(os.Environ(), vars),
		Stdin:    os.Stdin,
		Stdout:   stdout,
		Stderr:   stderr,
		Terminal: true,
	}
	if c.extension.bundled {
		// Started under the extension's name, the binary of ncly runs the
		// extension instead of core.
		program.Name = executablePrefix + c.extension.slug
	}
	if g.JSON && c.kind != contract.ResultText {
		return c.runChecked(ctx, p, g, program, stdout, stderr)
	}
	result, err := runner.Run(ctx, program)
	data := map[string]any{"Domain": c.extension.slug, "Exit": result.Exit}
	switch {
	case err != nil:
		return command.Report(p, g, contract.Core, protocolFailure(p, "extension.start_failed", data), stdout, stderr)
	case result.Signal != nil:
		// The extension's own text already told the human, unless its exit
		// code says otherwise.
		stopped := contract.Outcome{Signal: contract.StopSignal(result.Signal)}
		if verdict := contract.Core.Decide(stopped); result.Exit == int(verdict.Exit) {
			return int(verdict.Exit)
		}
		return command.Report(p, g, contract.Core, stopped, stdout, stderr)
	case !contract.ExitCode(result.Exit).Known():
		return command.Report(p, g, contract.Core, protocolFailure(p, "extension.answer_exit", data), stdout, stderr)
	}
	return result.Exit
}

// runChecked holds the extension's stdout and the last line of its stderr
// until it exits, then passes its answer on only when Answer check in
// contract.md accepts it. Otherwise it drops what it holds and answers
// RUNTIME, or the signal's code when a signal stopped the extension.
func (c *call) runChecked(ctx context.Context, p *i18n.Printer, g command.Globals, program runner.Program, stdout, stderr *os.File) int {
	var held bytes.Buffer
	lines := &heldLines{dest: stderr}
	program.Stdout, program.Stderr = &held, lines
	result, err := runner.Run(ctx, program)
	if err != nil {
		data := map[string]any{"Domain": c.extension.slug}
		return command.Report(p, g, contract.Core, protocolFailure(p, "extension.start_failed", data), stdout, stderr)
	}
	reply := contract.Reply{Exit: result.Exit, Stdout: held.Bytes(), Last: lines.finish()}
	answer, err := contract.CheckAnswer(c.extension.manifest, c.kind, reply)
	answerOut := io.Writer(stderr)
	if reply.OnStdout(c.kind) {
		// The answer belongs on stdout, so the last line of stderr is a
		// diagnostic like the others. It ends with a newline, so that what
		// core writes next starts a line of its own.
		last := reply.Last
		if len(last) > 0 && !bytes.HasSuffix(last, []byte("\n")) {
			last = append(slices.Clip(last), '\n')
		}
		_, _ = stderr.Write(last)
		answerOut = stdout
	}
	var rule *contract.RuleError
	switch {
	case result.Signal != nil && err == nil:
		// The first signal decides the verdict, and a valid answer keeps its
		// keys, as Answer check in contract.md asks.
		return reportKept(p, g, contract.Outcome{Signal: contract.StopSignal(result.Signal)}, answer, stdout, stderr)
	case result.Signal != nil:
		return command.Report(p, g, contract.Core, contract.Outcome{Signal: contract.StopSignal(result.Signal)}, stdout, stderr)
	case errors.As(err, &rule):
		return command.Report(p, g, contract.Core, protocolFailure(p, rule.Rule, rule.Data), stdout, stderr)
	}
	if _, err := answerOut.Write(answer.Line); err != nil {
		if answerOut == stdout {
			return reportKept(p, g, command.WriteFailure(p), answer, stdout, stderr)
		}
		return int(contract.ExitRuntime)
	}
	return result.Exit
}

// reportKept answers the failure that o decides with the other keys of an
// accepted answer, because the work that the answer describes happened, as
// Output in contract.md asks. It returns the exit code of o, or 1 when
// stderr refuses the answer too.
func reportKept(p *i18n.Printer, g command.Globals, o contract.Outcome, answer contract.ExtensionAnswer, stdout, stderr io.Writer) int {
	a := contract.Core.Finish(o, command.Summarizer(p))
	line, err := answer.Failed(a.Errors()[0])
	if err != nil {
		return command.Report(p, g, contract.Core, o, stdout, stderr)
	}
	if _, err := stderr.Write(append(line, '\n')); err != nil {
		return int(contract.ExitRuntime)
	}
	return int(a.Exit)
}

func protocolFailure(p *i18n.Printer, rule string, data map[string]any) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: p.T(rule, data), Hint: p.T("outcome.check_hint")}}}
}

// heldLines forwards each line of stderr to dest, except the latest one,
// which it holds until the next line arrives, because it may be the answer.
type heldLines struct {
	dest    io.Writer
	held    []byte
	partial []byte
}

func (h *heldLines) Write(b []byte) (int, error) {
	h.partial = append(h.partial, b...)
	for {
		end := bytes.IndexByte(h.partial, '\n')
		if end < 0 {
			return len(b), nil
		}
		h.hold(h.partial[:end+1])
		h.partial = h.partial[end+1:]
	}
}

// hold forwards the line held so far and holds line instead. A failed
// forward loses that line only, as a direct write to stderr would.
func (h *heldLines) hold(line []byte) {
	if h.held != nil {
		_, _ = h.dest.Write(h.held)
	}
	h.held = slices.Clone(line)
}

// finish returns the held line once the extension has exited. A last line
// without a newline counts as a line.
func (h *heldLines) finish() []byte {
	if len(h.partial) > 0 {
		h.hold(h.partial)
		h.partial = nil
	}
	return h.held
}

// extensionVars carries the resolved global flags and the identity of ncly
// to an extension. An empty value removes the variable.
func extensionVars(g command.Globals) map[string]string {
	self, _ := os.Executable()
	vars := g.Vars()
	vars["NCLY_VERSION"] = releaseVersion()
	vars["NCLY_CONTRACT_VERSION"] = strconv.Itoa(contract.Version)
	vars["NCLY_BIN"] = self
	return vars
}
