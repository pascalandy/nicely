package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// TestCoverageRequiresCheckedCalls exercises the kit through an independent extension repository.
func TestCoverageRequiresCheckedCalls(t *testing.T) {
	cases := []struct {
		name, script string
		covered      bool
	}{
		{"host success", "ncly video convert clip.mov --json", true},
		{"direct success", "ncly-video-convert clip.mov --json", true},
		{"host exits success", "exits 0 ncly video convert clip.mov --json", true},
		{"direct exits success", "exits 0 ncly-video-convert clip.mov --json", true},
		{"negated failure", "! ncly-video-convert --json", true},
		{"exact failure", "exits 2 ncly video convert --json", true},
		{"JSON environment", "env NCLY_JSON=1\nncly video convert clip.mov", true},
		{"help", "ncly video convert --help --json", false},
		{"human text", "ncly video convert clip.mov", false},
		{"skipped condition", "[!linux] [!darwin] ncly video convert clip.mov --json", false},
		{"stopped script", "stop\nncly video convert clip.mov --json", false},
		{"shell exec", "exec sh -c 'ncly video convert clip.mov --json'", false},
		{"background exec", "exec ncly video convert clip.mov --json &\nwait", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := edit{file: "testdata/script/convert.txtar", new: "grant example\n" + c.script + "\n"}
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), script)
			if c.covered {
				if err != nil {
					t.Fatalf("a checked call failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, "no scenario runs `ncly video convert`") {
				t.Fatalf("want missing coverage, got %v\n%s", err, out)
			}
		})
	}
}

// TestTextResultsNeedNoJSON preserves coverage for the contract's text-only commands.
func TestTextResultsNeedNoJSON(t *testing.T) {
	edits := []edit{
		{"manifest.toml", "[commands.result.schema]\ntype = \"object\"\nrequired = [\"file\", \"sha256\"]\nproperties = { file = { type = \"string\" }, sha256 = { type = \"string\" } }", "[commands.result]\nkind = \"text\""},
		{file: "testdata/script/convert.txtar", new: "ncly-video-convert clip.mov --json\nstdout 'Checked clip.mov'\n"},
	}
	if out, err := goTest(t, "standalone-named", os.Getenv("PATH"), edits...); err != nil {
		t.Fatalf("a declared text result failed: %v\n%s", err, out)
	}
}

// TestFailedReportsAreCompleteAnswers checks the full stdout object, not a matching line or field.
func TestFailedReportsAreCompleteAnswers(t *testing.T) {
	const report = `{"ok":false,"contract_version":1,"file":"clip.mov","sha256":"digest"}`
	cases := []struct {
		name, answer, problem string
	}{
		{"valid", report, ""},
		{"extra line", "prefix\n" + report, "want one line of answer on stdout"},
		{"second object", report + "\n" + report, "want one line of answer on stdout"},
		{"wrong verdict", strings.Replace(report, "false", "true", 1), "breaks the answer contract"},
		{"missing result", `{"ok":false,"contract_version":1}`, "breaks the schema"},
		{"wrong result type", strings.Replace(report, `"clip.mov"`, "3", 1), "breaks the schema"},
		{"errors on stdout", strings.TrimSuffix(report, "}") + `,"errors":[{"code":"RUNTIME","message":"Failed.","hint":"Check the input."}]}`, "breaks the answer contract"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edits := []edit{
				{"manifest.toml", "[commands.result.schema]", "[commands.result]\nkind = \"report\"\n\n[commands.result.schema]"},
				{file: "testdata/script/convert.txtar", new: "ncly skill view video-convert\nstdout 'ncly video convert clip.mov'\ncp report.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nchmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nexits 78 ncly-video-convert clip.mov --json\n\n-- report.sh --\n#!/bin/sh\ncat <<'ANSWER'\n" + c.answer + "\nANSWER\nexit 78\n"},
			}
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), edits...)
			if c.problem == "" {
				if err != nil {
					t.Fatalf("a valid failed report failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, c.problem) || !strings.Contains(out, "no scenario runs `ncly video convert`") {
				t.Fatalf("want an invalid, uncredited report with %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
