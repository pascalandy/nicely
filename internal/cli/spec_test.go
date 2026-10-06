package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/internal/contract"
	"github.com/pascalandy/nicely/internal/i18n"
	"github.com/pascalandy/nicely/internal/specdoc"
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

	root := newRoot(i18n.New(""))
	declared := slices.Clone(globalFlags)
	if v := root.Flags().Lookup("version"); v != nil {
		declared = append(declared, contract.Flag{Name: v.Name})
	}
	key := func(f contract.Flag) string { return f.Name + " " + f.Shorthand + " " + f.Value + " " + f.Env }
	got, want := sortedKeys(declared, key), sortedKeys(spec, key)
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
