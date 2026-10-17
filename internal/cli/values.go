package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/spf13/cobra"
)

// valuesLimit bounds a list command that completion runs, the reading of
// its output included, as Completion values in contract.md defines.
const valuesLimit = 2 * time.Second

// addValuesCompletion makes the arguments and the flags of one command that
// declare values_from complete from their list command.
func addValuesCompletion(root *cobra.Command, d contract.Command, values func(contract.ValuesFrom, string) []cobra.Completion) {
	c, _, err := root.Find(d.Path)
	if err != nil {
		return
	}
	if slices.ContainsFunc(d.Args, func(a contract.Arg) bool { return a.ValuesFrom != nil }) {
		declared := c.ValidArgsFunction
		c.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			i := len(args)
			if last := len(d.Args) - 1; i > last && d.Args[last].Variadic {
				i = last
			}
			if i < len(d.Args) && d.Args[i].ValuesFrom != nil {
				return values(*d.Args[i].ValuesFrom, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return declared(c, args, toComplete)
		}
	}
	for _, f := range d.Flags {
		if f.ValuesFrom == nil {
			continue
		}
		from := *f.ValuesFrom
		_ = c.RegisterFlagCompletionFunc(f.Name, func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return values(from, toComplete), cobra.ShellCompDirectiveNoFileComp
		})
	}
}

// completeValues suggests the values of a values_from that start with
// prefix. Any failure suggests nothing and prints nothing, so a Tab never
// shows an error: the reason goes to Cobra's completion debug log, which
// BASH_COMP_DEBUG_FILE turns on. Ctrl-C or SIGTERM ends the list command and
// its descendants at once, then ncly with the signal's code.
func completeValues(extensions []*extension, g command.Globals, from contract.ValuesFrom, prefix string) []cobra.Completion {
	values, err := listValues(extensions, g, from)
	var stopped *stoppedError
	if errors.As(err, &stopped) {
		os.Exit(int(contract.Core.Decide(contract.Outcome{Signal: contract.StopSignal(stopped.signal)}).Exit))
	}
	if err != nil {
		cobra.CompDebugln(fmt.Sprintf("values_from %q: %v", strings.Join(from.Command, " "), err), false)
		return nil
	}
	var suggested []cobra.Completion
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			suggested = append(suggested, v)
		}
	}
	return suggested
}

// stoppedError reports a list command that a signal to ncly ended.
type stoppedError struct{ signal os.Signal }

func (e *stoppedError) Error() string { return "stopped by " + e.signal.String() }

// listValues dispatches the list command that from names through the
// extension that discovery found, once its own declaration qualifies, and
// returns the field of each object under the key of its answer. The command
// runs in its own process group with the global flags of g and NCLY_JSON=1,
// never asks, and gets valuesLimit, after which SIGKILL ends it and its
// descendants at once.
func listValues(extensions []*extension, g command.Globals, from contract.ValuesFrom) ([]string, error) {
	path := strings.Join(from.Command, " ")
	i := slices.IndexFunc(extensions, func(e *extension) bool { return e.domain == from.Command[0] })
	if i < 0 {
		return nil, errors.New("no installed extension has this command")
	}
	e := extensions[i]
	if e.broken != nil {
		return nil, errors.New("its extension cannot run")
	}
	d, found := e.manifest.Command(path)
	if !found || !e.manifest.Qualifies(d) {
		return nil, errors.New("it is no list command that qualifies")
	}
	var stdout bytes.Buffer
	stderr := &heldLines{dest: io.Discard}
	// A Tab never waits on the keychain, so core reads no key for the
	// command, as Completion values in contract.md defines.
	vars := extensionVars(withJSON(g))
	missing, _ := keyVars(context.Background(), nil, e, d.Requires.Keys, false)
	maps.Copy(vars, missing)
	program := runner.Program{
		Path:   e.path,
		Args:   append(slices.Clone(d.Path[1:]), "--"),
		Env:    runner.Environ(os.Environ(), vars),
		Stdout: &stdout,
		Stderr: stderr,
		Limit:  valuesLimit,
	}
	if e.bundled {
		program.Name = executablePrefix + e.domain
	}
	result, err := runner.Run(runner.OnSignal(), program)
	switch {
	case err != nil:
		return nil, err
	case result.Signal != nil:
		return nil, &stoppedError{result.Signal}
	case result.TimedOut:
		return nil, fmt.Errorf("it did not answer within the timeout of %s", valuesLimit)
	}
	answer, err := contract.CheckAnswer(e.manifest, d.Result.Kind, contract.Reply{Exit: result.Exit, Stdout: stdout.Bytes(), Last: stderr.finish()})
	if err != nil {
		return nil, err
	}
	if result.Exit != 0 {
		return nil, fmt.Errorf("it failed with exit %d", result.Exit)
	}
	var keys map[string]json.RawMessage
	var items []map[string]json.RawMessage
	if json.Unmarshal(answer.Line, &keys) != nil || json.Unmarshal(keys[from.Key], &items) != nil || items == nil {
		return nil, fmt.Errorf("its answer has no array of objects under %q", from.Key)
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		var value string
		if raw := item[from.Field]; len(raw) == 0 || raw[0] != '"' || json.Unmarshal(raw, &value) != nil {
			return nil, fmt.Errorf("an object under %q has no string %q", from.Key, from.Field)
		}
		values = append(values, value)
	}
	return values, nil
}

// withJSON returns g with machine mode on, as a list command that completion
// runs gets it.
func withJSON(g command.Globals) command.Globals {
	g.JSON = true
	return g
}
