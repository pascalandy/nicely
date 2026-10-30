package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// TestManifestExamplesFailAndNameTheirFix rejects examples that never prepare or execute.
func TestManifestExamplesFailAndNameTheirFix(t *testing.T) {
	cases := []struct {
		name  string
		edits []edit
		want  []string
	}{
		{
			name: "an example that cannot prepare",
			edits: []edit{
				{"manifest.toml", `examples = ["ncly video convert clip.mov --json"]`, `examples = ["ncly video convert --json"]`},
			},
			want: []string{"manifest.toml: example `ncly video convert --json`", "exit 2", "USAGE_INVALID"},
		},
		{
			name:  "setup stops before execution",
			edits: []edit{{file: "testdata/examples.txtar", new: "stop\n"}},
			want:  []string{"manifest.toml: example `ncly video convert clip.mov --json` did not run"},
		},
		{
			name: "setup cannot replace scenario coverage",
			edits: []edit{
				{file: "testdata/examples.txtar", new: "grant example\nncly video convert clip.mov --json\n"},
				{file: "testdata/script/convert.txtar", new: "stop\n"},
			},
			want: []string{"no scenario runs `ncly video convert`"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), c.edits...)
			if err == nil {
				t.Fatalf("go test passed on the broken example:\n%s", out)
			}
			for _, want := range c.want {
				if !strings.Contains(out, want) {
					t.Errorf("go test failed without %q:\n%s", want, out)
				}
			}
		})
	}
}

// TestExamplesChooseSafeModes exercises mode selection through the independent fixture's kit.
func TestExamplesChooseSafeModes(t *testing.T) {
	const example = `examples = ["ncly video convert clip.mov --json"]`
	cases := []struct {
		name, want string
		edits      []edit
	}{
		{
			name: "paid command without a dry run",
			edits: []edit{
				{"manifest.toml", example, `examples = ["ncly video convert"]`},
				{"manifest.toml", "effects = {}", "effects = { paid = true }"},
			},
			want: "skipped: the command declares paid effects and has no dry run",
		},
		{
			name: "undeclared effects",
			edits: []edit{
				{"manifest.toml", example, `examples = ["ncly video convert"]`},
				{"manifest.toml", "effects = {}\n", ""},
			},
			want: "skipped: paid effects are undeclared and the command has no dry run",
		},
		{
			name: "dry run overrides false switches",
			edits: []edit{
				{"manifest.toml", example, `examples = ["ncly video convert example.mov --dry-run=false --json=false"]`},
				{"manifest.toml", "effects = {}", "effects = { paid = true }\nmodes = { dry_run = true }"},
				{"extension.go", "key, problem :=", `if call.Args[0] == "example.mov" {
		if call.Flags["dry-run"] != "true" {
			return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: "Paid work started.", Hint: "Use a dry run."}}}
		}
		return contract.Outcome{Data: map[string]string{"file": call.Args[0], "sha256": "dry-run"}}
	}
	key, problem :=`},
			},
			want: `manifest.toml: example ` + "`ncly video convert example.mov --dry-run=false --json=false`: run",
		},
		{
			name:  "quoted argument after the flag terminator",
			edits: []edit{{"manifest.toml", example, `examples = ["ncly video convert -- '--json=false quoted.mov'"]`}},
			want:  `"file":"--json=false quoted.mov"`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), c.edits...)
			if err != nil || !strings.Contains(out, c.want) {
				t.Fatalf("want %q, got %v\n%s", c.want, err, out)
			}
		})
	}
}
