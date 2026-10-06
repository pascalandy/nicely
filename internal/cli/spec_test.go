package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/pascalandy/nicely/internal/specdoc"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
)

func TestGlobalFlagsAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables("../../docs/north-star/cli-spec.md", "Global flags")
	if err != nil {
		t.Fatal(err)
	}
	var spec []contract.Flag
	for _, row := range tables[0] {
		var f contract.Flag
		for _, span := range specdoc.Backticked(row[0]) {
			name, value, _ := strings.Cut(span, " ")
			if long, isLong := strings.CutPrefix(name, "--"); isLong {
				f.Name, f.Value = long, strings.Trim(value, "<>")
			} else {
				f.Shorthand = strings.TrimPrefix(name, "-")
			}
		}
		if vars := specdoc.Backticked(row[1]); len(vars) > 0 {
			f.Env, _, _ = strings.Cut(vars[0], "=")
		}
		spec = append(spec, f)
	}

	key := func(f contract.Flag) string { return f.Name + " " + f.Shorthand + " " + f.Value + " " + f.Env }
	got, want := sortedKeys(globalFlags, key), sortedKeys(spec, key)
	if !slices.Equal(got, want) {
		t.Errorf("cli-spec.md lists the global flags\n%q\nncly declares\n%q", want, got)
	}
}

func TestReservedNamesAgreeWithTheGuide(t *testing.T) {
	tables, err := specdoc.Tables("../../docs/north-star/guide.md", "Domains and commands")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"help", "version", "config"}
	for _, row := range tables[0] {
		want = append(want, specdoc.Backticked(row[0])...)
	}
	slices.Sort(want)
	got := slices.Sorted(slices.Values(reservedNames))
	if !slices.Equal(got, want) {
		t.Errorf("the guide reserves %v, ncly reserves %v", want, got)
	}
}

func sortedKeys(flags []contract.Flag, key func(contract.Flag) string) []string {
	var keys []string
	for _, f := range flags {
		keys = append(keys, key(f))
	}
	slices.Sort(keys)
	return keys
}

// TestCommandsAgreeWithTheSpec compares the declared commands with the M0
// entries of the command tree, such as "completion zsh|bash|fish".
func TestCommandsAgreeWithTheSpec(t *testing.T) {
	tree, err := specdoc.Block("../../docs/north-star/cli-spec.md", "Command tree")
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, line := range tree {
		fields := strings.Fields(strings.TrimLeft(line, "│├└─ "))
		if len(fields) < 2 || fields[len(fields)-1] != "M0" {
			continue
		}
		path := fields[0]
		want = append(want, path)
		for _, words := range fields[1 : len(fields)-1] {
			for word := range strings.SplitSeq(words, "|") {
				want = append(want, path+" "+word)
			}
		}
	}
	var got []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if sub.Name() != "help" && !sub.Hidden {
				got = append(got, strings.TrimPrefix(sub.CommandPath(), "ncly "))
				walk(sub)
			}
		}
	}
	walk(newRoot(i18n.New(language.English)))
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("cli-spec.md lists the M0 commands %q, ncly declares %q", want, got)
	}
}

// TestEveryHelpPageEndsWithExamples checks the rule of the Usage section.
func TestEveryHelpPageEndsWithExamples(t *testing.T) {
	declarations := append([]contract.Command{rootCommand, helpDeclaration, completionCommand}, shellCommands()...)
	for _, d := range declarations {
		if n := len(d.Examples); n < 2 || n > 5 {
			t.Errorf("ncly %s has %d examples, want 2 to 5", strings.Join(d.Path, " "), n)
		}
	}
}

func shellCommands() []contract.Command {
	var commands []contract.Command
	for _, s := range shells {
		commands = append(commands, s.command)
	}
	return commands
}
