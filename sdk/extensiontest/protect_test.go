package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// copyCommand adds standalone copy, a command that writes a text output, to the standalone fixture.
// Each scenario replaces the installed extension with a shell host, as pairedFixture does.
const copyCommand = `[[commands]]
path = "standalone copy"
summary = "standalone.tally.summary"
effects = { user_writes = true }
modes = { dry_run = true }

[[commands.args]]
name = "input"
summary = "standalone.tally.summary"
required = true

[[commands.args]]
name = "output"
summary = "standalone.tally.summary"
required = true

[commands.result.schema]
type = "object"
required = ["output"]
properties = { output = { type = "string" } }

[commands.result.dry_run_schema]
type = "object"
required = ["output"]
properties = { output = { type = "string" } }

[programs.tally-tool]`

const copyRefusal = `echo '{"ok":false,"contract_version":1,"retry_safe":true,"errors":[{"code":"USAGE_INVALID","message":"notes.txt exists.","hint":"Add --force."}]}' >&2
exit 2
`

func copyFixture(t *testing.T, host, script string) (string, error) {
	t.Helper()
	ext := "$XDG_DATA_HOME/nicely/extensions/ncly-standalone"
	setup := "mkdir out\nrm " + ext + "\ncp host.sh " + ext + "\nchmod 755 " + ext + "\n"
	return goTest(t, "standalone", os.Getenv("PATH"),
		edit{"manifest.toml", "[programs.tally-tool]", copyCommand},
		edit{"extension.go", `"standalone fetch": fetch,`, "\"standalone fetch\": fetch,\n\t\t\"standalone copy\":  tally,"},
		edit{file: "testdata/script/copy.txtar", new: setup + script + "\n-- notes.md --\nFirst line\nLast line\n-- host.sh --\n" + host})
}

// signalCopyHost copies its input to a temporary file, writes the marker, and waits.
// On SIGTERM, it runs cleanup and answers TERMINATED.
func signalCopyHost(cleanup string) string {
	return `#!/bin/sh
[ "$1" = copy ] && shift
input=$1 output=$2 tmp=$(dirname "$2")/.copy.$$
cat "$input" > "$tmp"
trap '` + cleanup + `; echo '\''{"ok":false,"contract_version":1,"retry_safe":true,"errors":[{"code":"TERMINATED","message":"Stopped.","hint":"Run the command again."}]}'\'' >&2; exit 143' TERM
echo ready > started
while :; do sleep 0.01; done
`
}

// TestProtectSignal proves that a protect selection applies to the next signal call.
// The interrupted run keeps the input, leaves an absent output absent, and keeps an old output.
func TestProtectSignal(t *testing.T) {
	clean := signalCopyHost(`rm -f "$tmp"`)
	call := "signal TERM started ncly-standalone copy notes.md out/notes.txt\n"
	again := "rm started\n" + call
	old := "cp notes.md out/notes.txt\n"
	cases := []struct{ name, host, script, problem string }{
		{"clean interruption", clean, "protect notes.md out\n" + call, ""},
		{"marker in a protected folder", clean, "protect notes.md .\n" + call, ""},
		{"temporary file left", signalCopyHost(":"), "protect notes.md out\n" + call, "signal: TERM changed protected paths"},
		{"partial output", signalCopyHost(`mv "$tmp" "$output"`), "protect notes.md out/notes.txt\n" + call, "signal: TERM changed protected paths"},
		{"old output changed", signalCopyHost(`rm -f "$tmp"; echo partial > "$output"`), old + "protect notes.md out/notes.txt\n" + call, "signal: TERM changed protected paths"},
		{"input changed", signalCopyHost(`rm -f "$tmp"; echo more >> "$input"`), "protect notes.md\n" + call, "signal: TERM changed protected paths"},
		{"selection ends after one signal", signalCopyHost(`rm -f "$tmp"; [ -z "$DIRTY" ] || echo partial > "$output"`), "protect notes.md out\n" + call + "env DIRTY=1\n" + again, ""},
		{"new selection for each signal", signalCopyHost(`rm -f "$tmp"; [ -z "$DIRTY" ] || echo partial > "$output"`), "protect notes.md out\n" + call + "env DIRTY=1\nprotect notes.md out\n" + again, "signal: TERM changed protected paths"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := copyFixture(t, c.host, c.script)
			if c.problem == "" {
				if err != nil {
					t.Fatalf("valid signal scenario failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, c.problem) {
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}

// linkHost refuses an output that is a link. Before the refusal, the run part runs write.
func linkHost(write string) string {
	return `#!/bin/sh
dry_run=false
if [ "$1" = --dry-run=true ]; then dry_run=true; shift; fi
[ "$1" = copy ] && shift
input=$1 output=$2
[ "$dry_run" = true ] || { ` + write + `; }
` + copyRefusal
}

// TestProtectLink proves that protect and refuses catch a run that writes through a link.
// The output is a symbolic link to the input, and the protection records the link without following it.
func TestProtectLink(t *testing.T) {
	ext := "$XDG_DATA_HOME/nicely/extensions/ncly-standalone"
	refusal := "symlink out/notes.txt -> ../notes.md\nprotect notes.md out/notes.txt\nrefuses USAGE_INVALID ncly-standalone copy notes.md out/notes.txt\n"
	cases := []struct{ name, host, script, problem string }{
		{"refusal keeps the link", linkHost(":"), refusal, ""},
		{"write through the link", linkHost(`echo copy >> "$output"`), refusal, "run changed protected paths"},
		{"truncate through the link", linkHost(`: > "$output"`), refusal, "run changed protected paths"},
		{"replace the link", linkHost(`rm "$output"; cp "$input" "$output"`), refusal, "run changed protected paths"},
		{"selection stays after a signal", linkHost(`echo copy >> "$output"`), "symlink out/notes.txt -> ../notes.md\nprotect notes.md out/notes.txt\ncp signal.sh " + ext + "\n" +
			"signal TERM started ncly-standalone copy notes.md out/other.txt\ncp host.sh " + ext + "\nrefuses USAGE_INVALID ncly-standalone copy notes.md out/notes.txt\n" +
			"-- signal.sh --\n" + signalCopyHost(`rm -f "$tmp"`), "run changed protected paths"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := copyFixture(t, c.host, c.script)
			if c.problem == "" {
				if err != nil {
					t.Fatalf("valid link scenario failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, c.problem) {
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
