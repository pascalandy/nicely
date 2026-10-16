package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/internal/specdoc"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/spf13/cobra"
)

func TestGlobalFlagsAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables("../../docs/north-star/contract.md", "Global flags")
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
	got, want := sortedKeys(contract.GlobalFlags, key), sortedKeys(spec, key)
	if !slices.Equal(got, want) {
		t.Errorf("contract.md lists the global flags\n%q\nncly declares\n%q", want, got)
	}
}

// TestReservedNamesAgreeWithTheDomainTable compares the reserved names with the
// core and bundled rows of the domain table in architecture.md.
func TestReservedNamesAgreeWithTheDomainTable(t *testing.T) {
	tables, err := specdoc.Tables("../../docs/north-star/architecture.md", "Domains and commands")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"help", "version", "config", "core", "extensions"}
	for _, row := range tables[0] {
		if row[2] == "core" || row[2] == "bundled" {
			want = append(want, specdoc.Backticked(row[0])...)
		}
	}
	slices.Sort(want)
	got := slices.Sorted(slices.Values(reservedNames))
	if !slices.Equal(got, want) {
		t.Errorf("architecture.md reserves %v, ncly reserves %v", want, got)
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

// shipped lists the milestones whose commands core declares, and pending
// the commands of a shipped milestone whose card is still open: login in
// M06-T3 and logout in M06-T4.
var (
	shipped = []string{"M00", "M03", "M05", "M06"}
	pending = []string{"auth login", "auth logout"}
)

// TestCommandsAgreeWithTheSpec compares the declared commands with the
// entries of the command tree that shipped, such as
// "completion zsh|bash|fish". A word in brackets, such as [command...], is
// an argument, not a command. A bundled extension, the bundled rows of the
// domain table in architecture.md, declares its own commands.
func TestCommandsAgreeWithTheSpec(t *testing.T) {
	spec, err := specdoc.Block("../../docs/north-star/core-spec.md", "Command tree")
	if err != nil {
		t.Fatal(err)
	}
	tables, err := specdoc.Tables("../../docs/north-star/architecture.md", "Domains and commands")
	if err != nil {
		t.Fatal(err)
	}
	var bundled []string
	for _, row := range tables[0] {
		if row[2] == "bundled" {
			bundled = append(bundled, specdoc.Backticked(row[0])...)
		}
	}
	var want []string
	for _, line := range spec {
		fields := strings.Fields(strings.TrimLeft(line, "│├└─ "))
		if len(fields) < 2 || !slices.Contains(shipped, fields[len(fields)-1]) || slices.Contains(bundled, fields[0]) {
			continue
		}
		path := fields[0]
		want = append(want, path)
		for _, words := range fields[1 : len(fields)-1] {
			if strings.HasPrefix(words, "[") {
				continue
			}
			for word := range strings.SplitSeq(words, "|") {
				if !slices.Contains(pending, path+" "+word) {
					want = append(want, path+" "+word)
				}
			}
		}
	}
	var got []string
	for _, c := range tree()[1:] {
		if c.Name() != "help" {
			got = append(got, strings.TrimPrefix(c.CommandPath(), "ncly "))
		}
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("core-spec.md lists the commands %q up to %s, ncly declares %q", want, shipped[len(shipped)-1], got)
	}
}

// TestEveryHelpPageHasTwoToFiveExamples checks the rule of the Usage section
// on every command that ncly builds.
func TestEveryHelpPageHasTwoToFiveExamples(t *testing.T) {
	for _, c := range tree() {
		if n := len(slices.Collect(strings.Lines(c.Example))); n < 2 || n > 5 {
			t.Errorf("%s has %d examples, want 2 to 5", c.CommandPath(), n)
		}
	}
}

// tree returns every command that ncly builds, the root first.
func tree() []*cobra.Command {
	var all []*cobra.Command
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		all = append(all, c)
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(newRoot(catalogs.Printer("en"), command.Globals{}, nil, nil))
	return all
}
