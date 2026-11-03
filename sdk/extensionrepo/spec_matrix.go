package extensionrepo

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// matrixColumns are the columns of the table under ## Case matrix of spec.md.
// Each row is one family of cases. Accepted names the spec IDs of the inputs
// that pass, and Refused those of the inputs that the command refuses, or
// None (#423).
var matrixColumns = []string{"Family", "Source", "Accepted", "Refused"}

var matrixID = regexp.MustCompile("`([a-z][a-z0-9]*(?:-[a-z0-9]+)*)`")

// matrixProblems reports a family of ## Case matrix without a test, and a
// line of Failures or Acceptance in no family. Each spec ID of a family must
// be a line of Failures or Acceptance, or a limit that the user accepted. The
// test kit requires a test that claims each such line. Each family names a
// spec ID that passes, so each refusal has its pair.
func matrixProblems(spec string, limits []limit) []string {
	const fix = "Fix: give spec.md a table under ## Case matrix with the columns Family, Source, Accepted, and Refused, one row for each family of cases"
	proven := map[string]int{}
	for _, l := range SpecLines(spec) {
		if l.ID != "" && (l.Part == "Failures" || l.Part == "Acceptance") {
			proven[l.ID] = l.Number
		}
	}
	var problems []string
	named, header := map[string]bool{}, false
	for _, l := range SpecLines(spec) {
		if l.Part != "Case matrix" {
			continue
		}
		switch {
		case l.Kind == SpecHeader:
			header = slices.Equal(l.Cells, matrixColumns)
			if !header {
				problems = append(problems, fmt.Sprintf("spec.md:%d: the table of ## Case matrix has other columns. %s", l.Number, fix))
			}
		case l.Kind == SpecRow && header && len(l.Cells) == len(matrixColumns):
			accepted, refused := matrixIDs(l.Cells[2]), matrixIDs(l.Cells[3])
			at := fmt.Sprintf("spec.md:%d: the family %q ", l.Number, l.Cells[0])
			if len(accepted) == 0 {
				problems = append(problems, at+"names no spec ID of an input that passes. "+
					"Fix: name in Accepted the line whose test gives an input that passes, so each refusal has its pair")
			}
			if len(refused) == 0 && strings.Trim(l.Cells[3], "`*_.") != "None" {
				problems = append(problems, at+"names no spec ID of a refused input, and not None. Fix: name in Refused the line whose test refuses an input, or write None")
			}
			for _, id := range append(accepted, refused...) {
				named[id] = true
				_, found := proven[id]
				if !found && !slices.ContainsFunc(limits, func(l limit) bool { return l.id == id && l.accepted }) {
					problems = append(problems, fmt.Sprintf("%snames `%s`, which is no line of Failures or Acceptance and no accepted limit, so no test proves it. "+
						"Fix: prove it with a test in a round, or ask the user to accept it as a limit", at, id))
				}
			}
		case l.Kind == SpecRow:
			problems = append(problems, fmt.Sprintf("spec.md:%d: this row of ## Case matrix has no cell for each column. %s", l.Number, fix))
		}
	}
	if !header && len(problems) == 0 {
		return []string{"spec.md: it holds no table under ## Case matrix, so no list of the families of cases shows that each has a test. " + fix}
	}
	ids := slices.SortedFunc(maps.Keys(proven), func(a, b string) int { return proven[a] - proven[b] })
	for _, id := range ids {
		if !named[id] {
			problems = append(problems, fmt.Sprintf("spec.md:%d: the line `%s` is in no family of ## Case matrix. "+
				"Fix: name it in the row of its family, beside a spec ID of the other outcome", proven[id], id))
		}
	}
	return problems
}

// matrixIDs returns the spec IDs in backticks of a cell.
func matrixIDs(cell string) []string {
	var ids []string
	for _, m := range matrixID.FindAllStringSubmatch(cell, -1) {
		ids = append(ids, m[1])
	}
	return ids
}

// A family is one row of ## Case matrix: the spec IDs of its inputs that
// pass, and of those that the command refuses.
type family struct {
	accepted, refused []string
}

// matrixFamilies returns the families of ## Case matrix of spec, by the text
// of their Family cell.
func matrixFamilies(spec string) map[string]family {
	families, header := map[string]family{}, false
	for _, l := range SpecLines(spec) {
		switch {
		case l.Part != "Case matrix":
		case l.Kind == SpecHeader:
			header = slices.Equal(l.Cells, matrixColumns)
		case l.Kind == SpecRow && header && len(l.Cells) == len(matrixColumns):
			families[l.Cells[0]] = family{accepted: matrixIDs(l.Cells[2]), refused: matrixIDs(l.Cells[3])}
		}
	}
	return families
}
