package contract

import (
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/pascalandy/nicely/internal/specdoc"
)

const cliSpec = "../../docs/north-star/cli-spec.md"

func TestErrorCodesAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables(cliSpec, "Error codes")
	if err != nil {
		t.Fatal(err)
	}
	spec := map[Code]ExitCode{}
	for _, row := range tables[0] {
		if row[3] == "M00" {
			exit, _ := strconv.Atoi(row[1])
			spec[Code(specdoc.Backticked(row[0])[0])] = ExitCode(exit)
		}
	}
	code := map[Code]ExitCode{}
	for _, e := range errorCodes {
		code[e.code] = e.exit
	}
	if !maps.Equal(spec, code) {
		t.Errorf("cli-spec.md lists the M00 error codes %v, the registry holds %v", spec, code)
	}

	var warnings []Code
	for _, row := range tables[1] {
		if row[2] == "M00" {
			warnings = append(warnings, Code(specdoc.Backticked(row[0])[0]))
		}
	}
	if !slices.Equal(warnings, warningCodes) {
		t.Errorf("cli-spec.md lists the M00 warning codes %v, the registry holds %v", warnings, warningCodes)
	}
}

func TestExitCodesAgreeWithTheSpec(t *testing.T) {
	tables, err := specdoc.Tables(cliSpec, "Exit codes")
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
		t.Errorf("cli-spec.md lists the exit codes %v, the registry uses %v", spec, code)
	}
}
