package extensiontest_test

import (
	"cmp"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestPairedCommandsRejectPastMistakes proves that loose scenarios accepted each named defect.
// The replacement scenarios must reject the same fixture with the command's specific diagnostic.
func TestPairedCommandsRejectPastMistakes(t *testing.T) {
	cases := []struct{ name, host, repaired, legacy, guarded, want string }{
		{
			name:     "TERM mapped to INTERRUPTED",
			host:     signalHost("INTERRUPTED", "130", ""),
			repaired: signalHost("TERMINATED", "143", ""),
			legacy:   legacySignal,
			guarded:  "signal TERM started ncly-video-convert clip.mov\n",
			want:     "exited with 130, want 143",
		},
		{
			name:     "D009 cleanup exit becomes the result",
			host:     signalHost("TERMINATED", "1", "kill cleanup || exit 1"),
			repaired: signalHost("TERMINATED", "143", "kill cleanup || :"),
			legacy:   legacySignal,
			guarded:  "signal TERM started ncly-video-convert clip.mov\n",
			want:     "exited with 1, want 143",
		},
		{
			name:     "D096 fake file passes only the dry run",
			host:     refusalHost,
			repaired: strings.Replace(refusalHost, "if [ \"$1\" = fake.mov ]; then", "if false; then", 1),
			legacy:   "exits 2 ncly-video-convert missing.mov --dry-run\nexits 2 ncly-video-convert fake.mov\n",
			guarded:  "protect fake.mov output.mov\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n",
			want:     "dry run exited with 0, want a refusal with USAGE_INVALID",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if out, err := pairedFixture(t, c.host, c.legacy); err != nil {
				t.Fatalf("legacy scenario rejected the past mistake: %v\n%s", err, out)
			}
			out, err := pairedFixture(t, c.host, c.guarded)
			t.Logf("guarded fixture: %v\n%s", err, out)
			if err == nil || !strings.Contains(out, c.want) {
				t.Fatalf("want rejection with %q, got %v\n%s", c.want, err, out)
			}
			if out, err := pairedFixture(t, c.repaired, c.guarded); err != nil {
				t.Fatalf("repaired fixture failed: %v\n%s", err, out)
			}
			t.Log("legacy go test exit 0; guarded defect go test exit 1; guarded repair go test exit 0")
		})
	}
}

const legacySignal = `! exec $XDG_DATA_HOME/nicely/extensions/ncly-video-convert clip.mov --json &
exec sh -c 'i=0; until [ -f started ]; do i=$((i+1)); [ "$i" -lt 200 ] || exit 1; sleep 0.01; done; kill -TERM "$(cat host.pid)"'
wait
stderr '"code":"(INTERRUPTED|TERMINATED)"'
`

func signalHost(code, exit, cleanup string) string {
	return `#!/bin/sh
kill() { return 1; }
echo $$ > host.pid
trap 'echo '\''{"ok":false,"contract_version":1,"retry_safe":true,"errors":[{"code":"` + code + `","message":"Stopped.","hint":"Inspect the files."}]}'\'' >&2; ` + cmp.Or(cleanup, ":") + `; exit ` + exit + `' INT TERM
echo finished > finished.mov
echo ready > started
while :; do sleep 0.01; done
`
}

const refusalHost = `#!/bin/sh
dry_run=false
if [ "$1" = --dry-run=true ]; then dry_run=true; shift; fi
if [ "$1" = fake.mov ]; then
  if [ "$dry_run" = true ]; then
    echo '{"ok":true,"contract_version":1,"file":"fake.mov","sha256":"digest"}'
    exit 0
  fi
  for arg in "$@"; do
    if [ "$arg" = --dry-run ]; then
      echo '{"ok":true,"contract_version":1,"file":"fake.mov","sha256":"digest"}'
      exit 0
    fi
  done
fi
echo '{"ok":false,"contract_version":1,"retry_safe":true,"errors":[{"code":"USAGE_INVALID","message":"No media.","hint":"Select media."}]}' >&2
exit 2
`

func pairedFixture(t *testing.T, host, script string) (string, error) {
	t.Helper()
	setup := "grant example\nenv NCLY_JSON=1\n"
	// The legacy background call earns no coverage. Its unrelated checked call keeps that guard satisfied.
	if script == legacySignal {
		setup += "! ncly-video-convert --json\n"
	}
	setup += "rm $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\ncp host.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nchmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\n"
	// The kit runs the manifest example as a dry run, where ncly passes no key.
	return goTest(t, "standalone-named", os.Getenv("PATH"),
		edit{"extension.go", "key, problem :=", "if call.Flags[\"dry-run\"] == \"true\" {\n\t\treturn contract.Outcome{Data: map[string]string{\"file\": call.Args[0], \"sha256\": \"dry-run\"}}\n\t}\n\tkey, problem :="},
		edit{"manifest.toml", `requires = { keys = ["example"] }`, "requires = { keys = [\"example\"] }\nmodes = { dry_run = true }"},
		edit{"manifest.toml", "[commands.result.schema]", "[commands.result.dry_run_schema]\ntype = \"object\"\nrequired = [\"file\", \"sha256\"]\nproperties = { file = { type = \"string\" }, sha256 = { type = \"string\" } }\n\n[commands.result.schema]"},
		edit{"manifest.toml", "[[commands.args]]", "[[commands.flags]]\nname = \"label\"\nvalue = \"text\"\nsummary = \"video-convert.file\"\n\n[[commands.args]]"},
		edit{file: "testdata/script/convert.txtar", new: setup + script + "\n-- fake.mov --\nfake media\n-- host.sh --\n" + host})
}

// TestPublicationRejectsCheckedRename proves the past mistake of N3 in the builds of 2026-10-30.
// After a hard link failed, the fallback checked the output, then renamed over it, which Generic cases forbid.
// A host with hard links never enters that fallback, so the scenario of ordinary calls passes.
// Under without link, the guard of the kit refuses the hard link for the run only, and predicts sees the rename.
// The repaired stub fails closed, which keeps proves.
func TestPublicationRejectsCheckedRename(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("only Linux refuses an operation for the child only")
	}
	if _, err := exec.LookPath("perl"); err != nil {
		t.Skip("the fallback renames with perl, as os.Rename does")
	}
	link := `ln "$stage" "$out" 2>/dev/null || { rm -f "$stage"; fail RUNTIME "The output appeared." 1; }`
	fallback := strings.Replace(publicationHost, link, `ln "$stage" "$out" 2>/dev/null || {
    [ ! -e "$out" ] || { rm -f "$stage"; fail RUNTIME "The output appeared." 1; }
    perl -e 'rename $ARGV[0], $ARGV[1] or exit 1' "$stage" "$out"
  }`, 1)
	// The repaired stub is publicationHost, which fails closed without a hard link.
	withoutLink := strings.Replace(declared, "| `ncly video convert` | create | `link` | The stub refuses at out/.no-hard-link, and refuses links proves it |\n", "", 1)
	if out, err := publicationFixture(t, fallback, declared, proved); err != nil {
		t.Fatalf("legacy scenario rejected the past mistake: %v\n%s", err, out)
	}
	out, err := publicationFixture(t, fallback, declared, proved+"rm out/clip.txt\n[linux] without link predicts ncly video convert clip.mov\n")
	t.Logf("guarded fixture: %v\n%s", err, out)
	if want := "predicts: the run published out/clip.txt with `rename`, which the create mode of `ncly video convert` does not declare"; err == nil || !strings.Contains(out, want) {
		t.Fatalf("want rejection with %q, got %v\n%s", want, err, out)
	}
	script := proved + "rm out/clip.txt\nprotect clip.mov out\n[linux] without link keeps 1 ncly video convert clip.mov\n"
	if out, err := publicationFixture(t, publicationHost, withoutLink, script); err != nil {
		t.Fatalf("repaired fixture failed: %v\n%s", err, out)
	}
	t.Log("legacy go test exit 0; guarded fallback go test exit 1; guarded repair go test exit 0")
}
