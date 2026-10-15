package contract

import (
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/pascalandy/nicely/internal/specdoc"
)

const contractDoc = "../../docs/north-star/contract.md"

// shipped lists the milestones whose codes the registry holds.
var shipped = []string{"M00", "M01", "M03"}

func TestErrorCodesAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables(contractDoc, "Error codes")
	if err != nil {
		t.Fatal(err)
	}
	spec := map[Code]ExitCode{}
	for _, row := range tables[0] {
		if slices.Contains(shipped, row[3]) {
			exit, _ := strconv.Atoi(row[1])
			spec[Code(specdoc.Backticked(row[0])[0])] = ExitCode(exit)
		}
	}
	code := map[Code]ExitCode{}
	for _, e := range errorCodes {
		code[e.code] = e.exit
	}
	if !maps.Equal(spec, code) {
		t.Errorf("contract.md lists the error codes %v up to %s, the registry holds %v", spec, shipped[len(shipped)-1], code)
	}

	var warnings []Code
	for _, row := range tables[1] {
		if slices.Contains(shipped, row[2]) {
			warnings = append(warnings, Code(specdoc.Backticked(row[0])[0]))
		}
	}
	if !slices.Equal(warnings, warningCodes) {
		t.Errorf("contract.md lists the warning codes %v up to %s, the registry holds %v", warnings, shipped[len(shipped)-1], warningCodes)
	}
}

func TestExitCodesAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables(contractDoc, "Exit codes")
	if err != nil {
		t.Fatal(err)
	}
	var spec []ExitCode
	for _, row := range tables[0] {
		exit, _ := strconv.Atoi(specdoc.Backticked(row[0])[0])
		spec = append(spec, ExitCode(exit))
	}
	code := []ExitCode{ExitOK}
	for _, e := range errorCodes {
		if !slices.Contains(code, e.exit) {
			code = append(code, e.exit)
		}
	}
	slices.Sort(spec)
	slices.Sort(code)
	if !slices.Equal(spec, code) {
		t.Errorf("contract.md lists the exit codes %v, the registry uses %v", spec, code)
	}
}
