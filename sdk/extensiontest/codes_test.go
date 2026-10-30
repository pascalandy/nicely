package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// TestDeclaredCodesRequireObservations reproduces P-H4's comment-only and missing dry-run coverage gaps through the installed extension.
func TestDeclaredCodesRequireObservations(t *testing.T) {
	const warning = `{"code":"VIDEO_CONVERT_WARNING","message":"The input needs attention.","hint":"Check the input."}`
	const success = `{"ok":true,"contract_version":1,"file":"clip.mov","sha256":"digest"}`
	withWarning := strings.TrimSuffix(success, "}") + `,"warnings":[` + warning + `]}`
	cases := []struct {
		name, run, preview, declaration string
		want, absent                    []string
	}{
		{
			name: "comment only", run: success, preview: success,
			want: []string{
				"warning VIDEO_CONVERT_WARNING never observed for `ncly video convert` in run",
				"warning VIDEO_CONVERT_WARNING never observed for `ncly video convert` in dry_run",
				"Fix: add a JSON scenario",
			},
		},
		{
			name: "run without dry run", run: withWarning, preview: success,
			want:   []string{"warning VIDEO_CONVERT_WARNING never observed for `ncly video convert` in dry_run"},
			absent: []string{"never observed for `ncly video convert` in run"},
		},
		{
			name: "invalid answer", run: strings.Replace(withWarning, `"file":"clip.mov"`, `"file":3`, 1), preview: withWarning,
			want: []string{"breaks the schema", "warning VIDEO_CONVERT_WARNING never observed for `ncly video convert` in run"},
		},
		{
			name: "scoped run", run: withWarning, preview: success, declaration: "modes = [\"run\"]\n",
		},
		{
			name: "both modes", run: withWarning, preview: withWarning,
		},
		{
			name: "untested reason", run: success, preview: success,
			declaration: "untested = \"Requires a controlled runtime; https://example.com/issues/1\"\n",
			want:        []string{"manifest.toml:", "warning VIDEO_CONVERT_WARNING untested", "commands=video convert", "modes=run,dry_run", "Requires a controlled runtime; https://example.com/issues/1"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := "# VIDEO_CONVERT_WARNING appears here even when no answer emits it.\n" +
				"cp answer.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\n" +
				"chmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\n" +
				"ncly-video-convert clip.mov --json\n" +
				"ncly-video-convert clip.mov --json --dry-run\n\n" +
				"-- answer.sh --\n#!/bin/sh\ncase \"$*\" in\n*--dry-run*) cat <<'ANSWER'\n" + c.preview +
				"\nANSWER\n;;\n*) cat <<'ANSWER'\n" + c.run + "\nANSWER\n;;\nesac\n"
			edits := []edit{
				{"manifest.toml", `requires = { keys = ["example"] }`, "requires = { keys = [\"example\"] }\nmodes = { dry_run = true }"},
				{"manifest.toml", "domain = \"video\"", "warnings = [{ code = \"VIDEO_CONVERT_WARNING\"" + inlineDeclaration(c.declaration) + " }]\ndomain = \"video\""},
				{file: "testdata/script/convert.txtar", new: script},
			}
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), edits...)
			t.Logf("fixture go test: %v\n%s", err, out)
			wantFailure := c.name == "comment only" || c.name == "run without dry run" || c.name == "invalid answer"
			if (err != nil) != wantFailure {
				t.Fatalf("want failure %t, got %v\n%s", wantFailure, err, out)
			}
			for _, want := range c.want {
				if !strings.Contains(out, want) {
					t.Errorf("fixture output lacks %q:\n%s", want, out)
				}
			}
			for _, absent := range c.absent {
				if strings.Contains(out, absent) {
					t.Errorf("fixture output includes %q:\n%s", absent, out)
				}
			}
		})
	}
}

func inlineDeclaration(declaration string) string {
	if declaration == "" {
		return ""
	}
	return ", " + strings.TrimSpace(declaration)
}

// TestDeclaredCodeCommandScopes prevents an observation in one command from covering another command.
func TestDeclaredCodeCommandScopes(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		name, scope := "default", ""
		if scoped {
			name, scope = "explicit", `, commands = ["standalone tally"]`
		}
		t.Run(name, func(t *testing.T) {
			edits := []edit{
				{"manifest.toml", "format_version = 1", "warnings = [{ code = \"STANDALONE_WARNING\"" + scope + " }]\nformat_version = 1"},
				{"handlers.go", `return contract.Outcome{Data: map[string]int{"count": count}}`, `return contract.Outcome{Data: map[string]int{"count": count}, Warnings: []contract.Problem{{Code: "STANDALONE_WARNING", Message: "The input needs attention.", Hint: "Check the input."}}}`},
			}
			out, err := goTest(t, "standalone", os.Getenv("PATH"), edits...)
			if scoped {
				if err != nil {
					t.Fatalf("a scoped code failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, "warning STANDALONE_WARNING never observed for `ncly standalone fetch` in run") {
				t.Fatalf("a default scope must require the other command: %v\n%s", err, out)
			}
			if strings.Contains(out, "never observed for `ncly standalone tally`") || strings.Contains(out, "never observed for `ncly standalone`") {
				t.Fatalf("an observed command or grouping command added an obligation:\n%s", out)
			}
		})
	}
}

// TestDeclaredErrorsObserveProblemArrays covers top-level errors and batch item errors through schema-valid failures.
func TestDeclaredErrorsObserveProblemArrays(t *testing.T) {
	const problem = `{"code":"VIDEO_CONVERT_FAILED","message":"The input failed.","hint":"Check the input."}`
	for _, batch := range []bool{false, true} {
		name := "top level"
		answer := `{"ok":false,"contract_version":1,"errors":[` + problem + `]}`
		var edits []edit
		if batch {
			name = "batch item"
			answer = `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"The batch failed.","hint":"Check the input."}],"results":[{"ok":false,"errors":[` + problem + `]}]}`
			edits = append(edits, edit{"manifest.toml", "[commands.result.schema]", "[commands.result]\nkind = \"batch\"\n\n[commands.result.schema]"})
		}
		t.Run(name, func(t *testing.T) {
			edits = append(edits,
				edit{"manifest.toml", "format_version = 1", "errors = [{ code = \"VIDEO_CONVERT_FAILED\", exit = 1 }]\nformat_version = 1"},
				edit{file: "testdata/script/convert.txtar", new: "cp failure.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nchmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nexits 1 ncly-video-convert clip.mov --json\n\n-- failure.sh --\n#!/bin/sh\ncat >&2 <<'ANSWER'\n" + answer + "\nANSWER\nexit 1\n"},
			)
			if out, err := goTest(t, "standalone-named", os.Getenv("PATH"), edits...); err != nil {
				t.Fatalf("a validated error did not cover its declaration: %v\n%s", err, out)
			}
		})
	}
}
