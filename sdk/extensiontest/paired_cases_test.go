package extensiontest_test

import (
	"strings"
	"testing"
)

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
