package extensiontest_test

import (
	"strings"
	"testing"
)

// appearsHost copies its input to a temporary file and writes the marker.
// It waits for the output to appear, then runs commit.
func appearsHost(beforeMarker, commit string) string {
	return `#!/bin/sh
[ "$1" = copy ] && shift
input=$1 output=$2 tmp=$(dirname "$2")/.copy.$$
cat "$input" > "$tmp"
` + beforeMarker + `echo ready > started
i=0
until [ -e "$output" ] || [ "$i" -ge 500 ]; do i=$((i+1)); sleep 0.01; done
` + commit
}

// safeCommit links the temporary file to the output, which fails when the output exists.
const safeCommit = `if ln "$tmp" "$output" 2>/dev/null; then
  rm -f "$tmp"
  echo '{"ok":true,"contract_version":1,"output":"'"$output"'"}'
  exit 0
fi
rm -f "$tmp"
` + copyRefusal

// TestAppears proves that appears accepts a run that keeps a new file, and rejects each unsafe run.
func TestAppears(t *testing.T) {
	safe := appearsHost("", safeCommit)
	call := "appears started out/notes.txt ncly-standalone copy notes.md out/notes.txt\n"
	cases := []struct{ name, host, script, problem string }{
		{"keeps the new file", safe, "protect notes.md out\n" + call + "stderr '\"code\":\"USAGE_INVALID\"'\n", ""},
		{"through ncly", safe, "appears started out/notes.txt ncly standalone copy notes.md out/notes.txt\n", ""},
		{"rename over the new file", appearsHost("", "mv -f \"$tmp\" \"$output\"\n"+copyRefusal), call, "the run replaced out/notes.txt"},
		{"rename with the same bytes", appearsHost("", "cat \"$output\" > \"$tmp\"\nmv -f \"$tmp\" \"$output\"\n"+copyRefusal), call, "the run replaced out/notes.txt"},
		{"write through the new file", appearsHost("", "cat \"$tmp\" > \"$output\"\nrm -f \"$tmp\"\n"+copyRefusal), call, "the run changed out/notes.txt"},
		{"link replaces the new file", appearsHost("", "rm -f \"$output\"\nln -s ../notes.md \"$output\"\nrm -f \"$tmp\"\n"+copyRefusal), call, "out/notes.txt is no longer the regular file"},
		{"success over the new file", appearsHost("", "rm -f \"$tmp\"\necho '{\"ok\":true,\"contract_version\":1,\"output\":\"out/notes.txt\"}'\nexit 0\n"), call, "exited with 0, want a failure"},
		{"temporary file left", appearsHost("", copyRefusal), "protect notes.md out\n" + call, "changed protected paths"},
		{"input changed", appearsHost("", "rm -f \"$tmp\"\necho more >> \"$input\"\n"+copyRefusal), "protect notes.md\n" + call, "changed protected paths"},
		{"output before the marker", appearsHost("cp \"$tmp\" \"$output\"\n", "sleep 1\n"+safeCommit), call, "exists at the marker"},
		{"output before the call", safe, "cp notes.md out/notes.txt\n" + call, "out/notes.txt must not exist before the call"},
		{"stale marker", safe, "cp notes.md started\n" + call, "started must not exist before the call"},
		{"exit before the marker", "#!/bin/sh\n" + copyRefusal, call, "host exited before marker"},
		{"force", safe, "appears started out/notes.txt ncly-standalone copy notes.md out/notes.txt --force\n", "remove --force"},
		{"answer schema", appearsHost("", "rm -f \"$tmp\"\n"+strings.Replace(copyRefusal, `"ok":false`, `"ok":"false"`, 1)), call, "breaks the schema"},
		{"path bypass", safe, "appears started out/notes.txt $XDG_DATA_HOME/nicely/extensions/ncly-standalone copy notes.md out/notes.txt\n", "use the registered command"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := copyFixture(t, c.host, c.script)
			if c.problem == "" {
				if err != nil {
					t.Fatalf("valid appears scenario failed: %v\n%s", err, out)
				}
			} else if err == nil || !strings.Contains(out, c.problem) {
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
