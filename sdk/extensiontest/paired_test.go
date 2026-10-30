package extensiontest_test

import (
	"cmp"
	"os"
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

// TestPairedCommands checks the helpers through the same independent extension as the negative fixtures.
func TestPairedCommands(t *testing.T) {
	refusal := strings.Replace(refusalHost, "if [ \"$1\" = fake.mov ]; then", "if false; then", 1)
	cases := []struct{ name, host, script, problem string }{
		{"signal host without PATH", strings.ReplaceAll(signalHost("TERMINATED", "143", ""), "sleep 0.01", "/bin/sleep 0.01"), "env PATH=$WORK/bin\nsignal TERM started ncly video convert clip.mov\n", ""},
		{"signal stdin environment and directory", strings.Replace(signalHost("TERMINATED", "143", ""), "kill()", "read -r line\n[ \"$line\" = 'fake media' ] && [ \"$SIGNAL_FIXTURE\" = ready ] || exit 3\nkill()", 1), "env SIGNAL_FIXTURE=ready\nstdin fake.mov\nmkdir sub\ncd sub\nsignal TERM started ncly-video-convert clip.mov\n", ""},
		{"refusal host without PATH", refusal, "env PATH=$WORK/bin\nrefuses USAGE_INVALID ncly video convert fake.mov\n", ""},
		{"INT host", signalHost("INTERRUPTED", "130", ""), "survives finished.mov\nsignal INT started ncly video convert clip.mov\nstderr '\"code\":\"INTERRUPTED\"'\n", ""},
		{"TERM direct", signalHost("TERMINATED", "143", ""), "survives finished.mov\nsignal TERM started ncly-video-convert clip.mov\nstderr '\"code\":\"TERMINATED\"'\n", ""},
		{"finished file removed", signalHost("TERMINATED", "143", "rm finished.mov"), "survives finished.mov\nsignal TERM started ncly-video-convert clip.mov\n", "finished.mov: no such file"},
		{"finished file changed", signalHost("TERMINATED", "143", "echo changed > finished.mov"), "survives finished.mov\nsignal TERM started ncly-video-convert clip.mov\n", "finished files changed"},
		{"unfinished file", signalHost("TERMINATED", "143", ""), "survives missing.mov\nsignal TERM started ncly-video-convert clip.mov\n", "missing.mov: no such file"},
		{"stale marker", signalHost("TERMINATED", "143", ""), "cp fake.mov started\nsignal TERM started ncly-video-convert clip.mov\n", "must not exist before the call"},
		{"signal answer schema", strings.Replace(signalHost("TERMINATED", "143", ""), `"ok":false`, `"ok":"false"`, 1), "signal TERM started ncly-video-convert clip.mov\n", "breaks the schema"},
		{"paired host refusal", refusal, "protect fake.mov output.mov\nrefuses USAGE_INVALID ncly video convert fake.mov\nstderr '\"code\":\"USAGE_INVALID\"'\n", ""},
		{"paired direct refusal", refusal, "protect fake.mov output.mov\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n", ""},
		{"identical stdin", strings.Replace(refusal, "echo '{\"ok\":false", "read -r line\n[ \"$line\" = 'fake media' ] || exit 3\necho '{\"ok\":false", 1), "stdin fake.mov\nprotect fake.mov\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n", ""},
		{"flags before separator", refusal, "refuses USAGE_INVALID ncly-video-convert -- fake.mov\n", ""},
		{"separator as flag value", strings.Replace(refusal, "echo '{\"ok\":false", "[ \"$#\" -eq 3 ] && [ \"$1\" = fake.mov ] && [ \"$2\" = --label ] && [ \"$3\" = -- ] || exit 3\necho '{\"ok\":false", 1), "refuses USAGE_INVALID ncly-video-convert fake.mov --label --\n", ""},
		{"wrong refusal", refusal, "refuses RUNTIME ncly-video-convert fake.mov\n", "returned USAGE_INVALID, want RUNTIME"},
		{"dry run changes input", strings.Replace(refusal, "echo '{\"ok\":false", "echo changed > fake.mov\necho '{\"ok\":false", 1), "protect fake.mov\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n", "dry run changed protected paths"},
		{"run creates output", strings.Replace(refusal, "echo '{\"ok\":false", "[ \"$dry_run\" = true ] || echo output > output.mov\necho '{\"ok\":false", 1), "protect fake.mov output.mov\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n", "run changed protected paths"},
		{"run changes folder", strings.Replace(refusal, "echo '{\"ok\":false", "[ \"$dry_run\" = true ] || echo output > inputs/output.mov\necho '{\"ok\":false", 1), "mkdir inputs\nprotect inputs\nrefuses USAGE_INVALID ncly-video-convert fake.mov\n", "run changed protected paths"},
		{"refusal answer schema", strings.Replace(refusal, `"ok":false`, `"ok":"false"`, 1), "refuses USAGE_INVALID ncly-video-convert fake.mov\n", "breaks the schema"},
		{"explicit JSON off", refusal, "refuses USAGE_INVALID ncly-video-convert fake.mov --json=false\n", "require JSON"},
		{"path bypass", refusal, "refuses USAGE_INVALID $XDG_DATA_HOME/nicely/extensions/ncly-video-convert fake.mov\n", "use the registered command"},
		{"dry run already selected", refusal, "refuses USAGE_INVALID ncly-video-convert fake.mov -n\n", "without a dry-run flag"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := pairedFixture(t, c.host, c.script)
			if c.problem == "" {
				if err != nil {
					t.Fatalf("valid paired scenario failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, c.problem) {
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
