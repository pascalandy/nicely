package extensiontest

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/rogpeppe/go-internal/txtar"
)

// RunExamples runs manifest examples with the runner's scenario setup and commands.
// The optional fixture supplies setup and files. Example answers never earn scenario coverage.
func (c *Coverage) RunExamples(t *testing.T, fixture string, params testscript.Params) {
	t.Helper()
	t.Run("examples", func(t *testing.T) {
		archive, err := txtar.ParseFile(fixture)
		if os.IsNotExist(err) {
			archive, err = &txtar.Archive{}, nil
		}
		if err != nil {
			t.Fatal(err)
		}
		// Use a separate recorder so examples cannot satisfy scenario obligations.
		answers := c.fresh()
		params.Dir, params.Files = "", []string{}
		params.Cmds = maps.Clone(params.Cmds)
		if params.Cmds == nil {
			params.Cmds = map[string]func(*testscript.TestScript, bool, []string){}
		}
		dir := t.TempDir()
		for i, command := range c.manifest.Commands {
			for j, example := range command.Examples {
				name := fmt.Sprintf("%s-%d", strings.Join(command.Path, "_"), j+1)
				words, reason, err := c.exampleCall(example)
				if err != nil {
					t.Errorf("manifest.toml: example `%s`: %v. Fix: correct the example", example, err)
					continue
				}
				if reason != "" {
					t.Run(name, func(t *testing.T) { t.Skipf("manifest.toml: example `%s`: skipped: %s", example, reason) })
					continue
				}
				// Each closure holds the original example and its prepared argument list.
				key := fmt.Sprintf("manifest-example-%d-%d", i, j)
				var ran atomic.Bool
				t.Cleanup(func() {
					if !ran.Load() {
						t.Errorf("manifest.toml: example `%s` did not run. Fix: remove stop or skip from %s", example, fixture)
					}
				})
				params.Cmds[key] = func(ts *testscript.TestScript, neg bool, args []string) {
					if neg || len(args) != 0 {
						ts.Fatalf("%s takes no arguments", key)
					}
					ts.Logf("manifest.toml: example `%s`: run %q", example, words)
					ran.Store(true)
					exit, err := exitCode(ts.Exec("ncly", words...))
					if err != nil {
						ts.Fatalf("manifest.toml: example `%s`: %v", example, err)
					}
					stdout, stderr := ts.ReadFile("stdout"), ts.ReadFile("stderr")
					if problem := answers.Observe(words, exit, stdout, stderr, "1"); problem != "" {
						ts.Fatalf("manifest.toml: example `%s`: %s", example, problem)
					}
					if exit != 0 {
						ts.Fatalf("manifest.toml: example `%s`: exit %d. Fix: correct the example or its %s fixtures\n%s%s", example, exit, fixture, stdout, stderr)
					}
				}
				generated := txtar.Archive{Comment: []byte(strings.TrimSpace(string(archive.Comment)) + "\n" + key + "\n"), Files: archive.Files}
				file := filepath.Join(dir, name+".txtar")
				if err := os.WriteFile(file, txtar.Format(&generated), 0o644); err != nil {
					t.Fatal(err)
				}
				params.Files = append(params.Files, file)
			}
		}
		if len(params.Files) != 0 {
			testscript.Run(t, params)
		}
	})
}

// fresh shares the declarations and answer checks without sharing recorded calls.
func (c *Coverage) fresh() *Coverage {
	return &Coverage{manifest: c.manifest, printer: c.printer, seen: map[string]bool{}}
}

// exampleCall resolves the invoked command before choosing its safe mode.
func (c *Coverage) exampleCall(example string) ([]string, string, error) {
	words, err := exampleWords(example)
	if err != nil {
		return nil, "", err
	}
	if len(words) < 2 || words[0] != "ncly" {
		return nil, "", fmt.Errorf("want a command starting with ncly")
	}
	words = words[1:]
	call, found := c.find(words)
	if !found {
		return nil, "command is outside this manifest, so its paid effects are unknown", nil
	}
	if !call.runs {
		return nil, "", fmt.Errorf("the example does not name a runnable command")
	}
	flags := []string{"--json=true", "--no-input=true"}
	switch {
	case call.command.Modes.DryRun:
		flags = append(flags, "--dry-run=true")
	case call.command.Effects == nil:
		return nil, "paid effects are undeclared and the command has no dry run", nil
	case call.command.Effects.Paid:
		return nil, "the command declares paid effects and has no dry run", nil
	}
	// Global flags must precede --, which makes all subsequent words arguments.
	at := slices.Index(words, "--")
	if at < 0 {
		at = len(words)
	}
	return slices.Insert(words, at, flags...), "", nil
}

// exampleWords reads literal shell-style words without expansion or a shell process.
func exampleWords(line string) ([]string, error) {
	var words []string
	var word strings.Builder
	var quote byte
	started := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && quote != '\'':
			i++
			if i == len(line) {
				return nil, fmt.Errorf("unfinished escape")
			}
			word.WriteByte(line[i])
			started = true
		case quote != 0:
			if ch == quote {
				quote = 0
			} else {
				word.WriteByte(ch)
			}
		case ch == '\'' || ch == '"':
			quote, started = ch, true
		case strings.ContainsRune(" \t\r\n", rune(ch)):
			if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
		case strings.ContainsRune(";|&<>`", rune(ch)):
			return nil, fmt.Errorf("shell operators are unsupported: %s", strconv.QuoteRune(rune(ch)))
		default:
			word.WriteByte(ch)
			started = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unfinished quote")
	}
	if started {
		words = append(words, word.String())
	}
	return words, nil
}
