package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// publicationHost stands for ncly-video-convert, which publishes out/<name>.txt.
// A run without --force publishes with a hard link, and a run with --force with a rename.
// The file out/.no-hard-link stands for a destination whose file system has no hard link.
// Its dry run checks the input and an existing output, but not the hard link, as in the builds of #425.
const publicationHost = `#!/bin/sh
dry=false force=false input=
for arg in "$@"; do
  case "$arg" in
    --dry-run|--dry-run=true) dry=true ;;
    --force|--force=true) force=true ;;
    -*) ;;
    *) input=$arg ;;
  esac
done
fail() {
  echo "{\"ok\":false,\"contract_version\":1,\"retry_safe\":true,\"errors\":[{\"code\":\"$1\",\"message\":\"$2\",\"hint\":\"Fix it.\"}]}" >&2
  exit "$3"
}
out=out/${input%.mov}.txt
[ -f "$input" ] || fail USAGE_INVALID "No input." 2
overwrite=false
if [ -e "$out" ]; then
  [ "$force" = true ] || fail USAGE_INVALID "The output exists." 2
  overwrite=true
fi
# CHECK
if [ "$dry" = true ]; then
  echo "{\"ok\":true,\"contract_version\":1,\"dry_run\":true,\"output\":\"$out\",\"overwrite\":$overwrite,\"pending_checks\":[]}"
  exit 0
fi
stage=out/.stage.$$
cat "$input" > "$stage"
if [ "$force" = true ]; then
  mv -f "$stage" "$out"
else
  if [ -n "$APPEARS" ]; then
    echo ready > started
    until [ -e "$out" ]; do sleep 0.01; done
  fi
  if [ -e out/.no-hard-link ]; then rm -f "$stage"; fail RUNTIME "The destination has no hard link." 1; fi
  ln "$stage" "$out" 2>/dev/null || { rm -f "$stage"; fail RUNTIME "The output appeared." 1; }
  rm -f "$stage"
fi
echo "{\"ok\":true,\"contract_version\":1,\"input\":\"$input\",\"output\":\"$out\"}"
`

// repairedHost checks the hard link in its dry run too, and refuses as its run does.
var repairedHost = strings.Replace(publicationHost, "# CHECK", `[ "$force" = true ] || [ ! -e out/.no-hard-link ] || fail CAPABILITY_UNSUPPORTED "The destination has no hard link." 78`, 1)

// declared is ## Publication of spec.md for publicationHost. No check of the kit removes a hard link or a rename yet, so both have a reason.
const declared = "\n## Publication\n\n" +
	"| Command | Mode | Destination | Overwrite | Protected | Operation | Dry run |\n|---|---|---|---|---|---|---|\n" +
	"| `ncly video convert` | create | `output` | `overwrite` | `file` | `link` | checks |\n" +
	"| `ncly video convert` | replace | `output` | `overwrite` | `file` | `rename` | checks |\n\n" +
	"| Command | Mode | Obligation | Reason |\n|---|---|---|---|\n" +
	"| `ncly video convert` | create | `link` | The stub refuses at out/.no-hard-link, and refuses links proves it |\n" +
	"| `ncly video convert` | replace | `rename` | The stub renames with mv -f |\n"

// proved runs every obligation of declared that a scenario can prove.
const proved = `predicts ncly video convert clip.mov
exists out/clip.txt
predicts ncly-video-convert clip.mov --force
rm out/clip.txt
env APPEARS=1
appears started out/clip.txt ncly video convert clip.mov
env APPEARS=
`

// publicationFixture gives standalone-named a command that publishes a file, the publication of spec, and host as its program.
func publicationFixture(t *testing.T, host, spec, script string) (string, error) {
	t.Helper()
	setup := "grant example\nenv NCLY_JSON=1\nmkdir out\n" +
		"rm $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\ncp host.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nchmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\n"
	// The kit runs the manifest example as a dry run with the real handler.
	return goTest(t, "standalone-named", os.Getenv("PATH"),
		edit{"extension.go", "key, problem :=", "if call.Flags[\"dry-run\"] == \"true\" {\n\t\treturn contract.Outcome{Data: map[string]any{\"dry_run\": true, \"output\": \"out/clip.txt\", \"overwrite\": false, \"pending_checks\": []string{}}}\n\t}\n\tkey, problem :="},
		edit{"manifest.toml", "effects = {}", "effects = { user_writes = true }\nmodes = { dry_run = true }"},
		edit{"manifest.toml", "[[commands.args]]", "[[commands.flags]]\nname = \"force\"\nshorthand = \"f\"\nsummary = \"video-convert.file\"\n\n[[commands.args]]"},
		edit{
			"manifest.toml", "[commands.result.schema]\ntype = \"object\"\nrequired = [\"file\", \"sha256\"]\nproperties = { file = { type = \"string\" }, sha256 = { type = \"string\" } }",
			"[commands.result.dry_run_schema]\ntype = \"object\"\nrequired = [\"dry_run\", \"output\", \"overwrite\", \"pending_checks\"]\n" +
				"properties = { dry_run = { const = true }, output = { type = \"string\" }, overwrite = { type = \"boolean\" }, pending_checks = { type = \"array\", items = { type = \"string\" } } }\n\n" +
				"[commands.result.schema]\ntype = \"object\"\nrequired = [\"input\", \"output\"]\nproperties = { input = { type = \"string\" }, output = { type = \"string\" } }",
		},
		edit{"spec.md", "The direct executable uses that sentinel without a host grant.", "The direct executable uses that sentinel without a host grant.\n" + spec},
		edit{file: "testdata/script/convert.txtar", new: setup + script + "\n-- clip.mov --\nfake media\n-- host.sh --\n" + host})
}

// TestPublicationRejectsIssue425 proves the past mistake of #425: a dry run passed, and the run failed on a destination with no hard link.
// Before COR-104-T6, the scenario of ordinary calls passed. The declaration now asks for each obligation, and predicts rejects the run.
func TestPublicationRejectsIssue425(t *testing.T) {
	legacy := "cp clip.mov out/.no-hard-link\nncly video convert clip.mov --dry-run\nexits 1 ncly video convert clip.mov\nstderr '\"code\":\"RUNTIME\"'\n"
	steps := []struct{ name, host, spec, script, want string }{
		{"no declaration", publicationHost, "", legacy, "## Publication has no row for it"},
		{"no check of the hard link", publicationHost, strings.Split(declared, "\n| Command | Mode | Obligation")[0], proved, "has no scenario that proves `link`, and no reason"},
		{
			"dry run predicts a failed run", publicationHost, declared, "cp clip.mov out/.no-hard-link\npredicts ncly video convert clip.mov\n",
			"predicts: the dry run predicted a file at out/clip.txt, and the run failed with RUNTIME and exit 1",
		},
	}
	for _, s := range steps {
		t.Run(s.name, func(t *testing.T) {
			out, err := publicationFixture(t, s.host, s.spec, s.script)
			t.Logf("go test: %v\n%s", err, out)
			if err == nil || !strings.Contains(out, s.want) {
				t.Fatalf("want rejection with %q, got %v\n%s", s.want, err, out)
			}
		})
	}
	t.Run("repaired", func(t *testing.T) {
		script := proved + "rm out/clip.txt\ncp clip.mov out/.no-hard-link\nprotect clip.mov out\nrefuses CAPABILITY_UNSUPPORTED ncly video convert clip.mov\n"
		if out, err := publicationFixture(t, repairedHost, declared, script); err != nil {
			t.Fatalf("repaired fixture failed: %v\n%s", err, out)
		}
	})
}
