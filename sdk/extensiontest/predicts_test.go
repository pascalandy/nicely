package extensiontest_test

import (
	"strings"
	"testing"
)

// TestPredicts checks each rule of predicts and of the form through the same fixture.
func TestPredicts(t *testing.T) {
	cases := []struct{ name, host, spec, script, problem string }{
		{"each obligation proved", publicationHost, declared, proved, ""},
		{"dry run fills the cache", strings.Replace(publicationHost, "# CHECK", `[ "$dry" = true ] && mkdir -p "$XDG_CACHE_HOME/fill" && echo cached > "$XDG_CACHE_HOME/fill/entry"`, 1), declared, proved, ""},
		{"dry run writes the destination", strings.Replace(publicationHost, "# CHECK", `[ "$dry" = true ] && echo planned > "$out"`, 1), declared, "predicts ncly video convert clip.mov\n", "predicts: the dry run changed $WORK/out/clip.txt"},
		{"dry run writes a record", strings.Replace(publicationHost, "# CHECK", `[ "$dry" = true ] && mkdir -p "$XDG_STATE_HOME/runs" && echo run > "$XDG_STATE_HOME/runs/record"`, 1), declared, "predicts ncly video convert clip.mov\n", "predicts: the dry run changed $WORK/.local/state/runs"},
		{"run publishes elsewhere", strings.Replace(publicationHost, `\"input\":\"$input\",\"output\":\"$out\"`, `\"input\":\"$input\",\"output\":\"out/other.txt\"`, 1), declared, "predicts ncly video convert clip.mov\n", `predicts: the run published at "out/other.txt", and the dry run predicted out/clip.txt`},
		{"wrong overwrite choice", strings.Replace(publicationHost, "  overwrite=true\n", "", 1), declared, "cp clip.mov out/clip.txt\npredicts ncly video convert clip.mov --force\n", "predicts: the dry run gives overwrite false, and the destination out/clip.txt exists"},
		{"create over an existing file", strings.Replace(publicationHost, `[ "$force" = true ] || fail USAGE_INVALID "The output exists." 2`, ":", 1), declared, "cp clip.mov out/clip.txt\npredicts ncly video convert clip.mov\n", "predicts: the dry run predicts a new file at out/clip.txt, which exists"},
		{"run changes the input", strings.Replace(publicationHost, `stage=out/.stage.$$`, `echo changed > "$input"; stage=out/.stage.$$`, 1), declared, "predicts ncly video convert clip.mov\n", "predicts: the run changed protected paths: $WORK/clip.mov"},
		{"pending without its check", publicationHost, strings.ReplaceAll(declared, "| checks |", "| pending |"), "predicts ncly video convert clip.mov\n", "predicts: the dry run lists no publication.link in pending_checks"},
		{"pending with its check", strings.Replace(publicationHost, `\"pending_checks\":[]`, `\"pending_checks\":[\"publication.link\",\"publication.rename\"]`, 1), strings.ReplaceAll(declared, "| checks |", "| pending |"), proved, ""},
		{"undeclared mode", publicationHost, strings.Replace(declared, "| `ncly video convert` | replace | `output` | `overwrite` | `file` | `rename` | checks |\n", "", 1), "predicts ncly video convert clip.mov --force\n", "predicts: ## Publication of spec.md declares no replace mode of `ncly video convert`"},
		{"replace mode without a row", publicationHost, strings.Replace(declared, "| `ncly video convert` | replace | `output` | `overwrite` | `file` | `rename` | checks |\n", "", 1), "predicts ncly video convert clip.mov\n", "spec.md: `ncly video convert` may write a user file, as its effects set user_writes or are absent, and ## Publication has no row for its replace mode"},
		{"refused dry run", publicationHost, declared, "predicts ncly video convert missing.mov\n", "predicts: the dry run refused with USAGE_INVALID and exit 2"},
		{"stale reason", publicationHost, declared + "| `ncly video convert` | create | `predicts` | Stale |\n", proved, "a scenario proves `predicts` of the create mode of `ncly video convert`, so its reason is stale"},
		{"missing appears", publicationHost, declared, "predicts ncly video convert clip.mov\npredicts ncly video convert clip.mov --force\n", "has no scenario that proves `appears`, and no reason"},
		{"form against the manifest", publicationHost, "\n## Publication\n\n" +
			"| Command | Mode | Destination | Overwrite | Protected | Operation | Dry run |\n|---|---|---|---|---|---|---|\n" +
			"| `ncly video convert` | create | `path` | `overwrite` | `input` | `rename` | maybe |\n" +
			"| `ncly video convert` | append | `output` | `overwrite` | `--input` | `copy` | checks |\n" +
			"| `ncly video subtitle` | create | `output` | `overwrite` | None | `link` | checks |\n" +
			"| `ncly video convert` | replace | `output` | `overwrite` | None | `rename` | checks |\n" +
			"| `ncly video convert` | replace | `output` | `overwrite` | None | `rename` | checks |\n\n" +
			"| Command | Mode | Obligation | Reason |\n|---|---|---|---|\n" +
			"| `ncly video convert` | create | `signal` | No such obligation |\n" +
			"| `ncly video convert` | replace | `publication` | It writes no file |\n", proved, ""},
	}
	formProblems := []string{
		"the key `path` of Destination is no property of a result schema",
		"Protected names `input`, which is no argument",
		"the mode create publishes with `rename`",
		`the dry-run treatment "maybe" is not checks or pending`,
		`the mode "append" is not create or replace`,
		"the operation `copy` is unknown",
		"`ncly video subtitle` is no running command of the manifest",
		"the reason names `signal`, which is no obligation",
		"the replace mode of `ncly video convert` repeats the row of line 21",
		"the obligation `publication` says that the mode replace of `ncly video convert` writes no file, and that mode has a row",
		"Protected names `--input`, which is no argument and no flag with one value",
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := publicationFixture(t, c.host, c.spec, c.script)
			switch {
			case c.name == "form against the manifest":
				for _, want := range formProblems {
					if err == nil || !strings.Contains(out, want) {
						t.Errorf("want %q, got %v", want, err)
					}
				}
				if t.Failed() {
					t.Log(out)
				}
			case c.problem == "" && err != nil:
				t.Fatalf("valid publication scenario failed: %v\n%s", err, out)
			case c.problem != "" && (err == nil || !strings.Contains(out, c.problem)):
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
