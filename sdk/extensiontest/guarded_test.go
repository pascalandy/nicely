package extensiontest_test

import (
	"runtime"
	"strings"
	"testing"
)

// closedHost fails closed when the destination refuses the rename of its forced run.
var closedHost = strings.Replace(publicationHost, `mv -f "$stage" "$out"`, `mv -f "$stage" "$out" 2>/dev/null || { rm -f "$stage"; fail RUNTIME "The destination refused the rename." 1; }`, 1)

// TestWithout checks the fixture that refuses an operation for one call, and keeps, through the fixture of TestPredicts.
func TestWithout(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("only Linux refuses an operation for one call")
	}
	noReasons := strings.Split(declared, "\n| Command | Mode | Obligation")[0]
	// Each operation of declared is proved, so no reason remains.
	allProved := proved + "rm out/clip.txt\n" +
		"protect clip.mov out\nwithout link keeps 1 ncly video convert clip.mov\n" +
		"cp clip.mov out/clip.txt\nprotect out\nwithout rename keeps 1 ncly video convert clip.mov --force\n"
	cases := []struct{ name, host, spec, script, problem string }{
		{"each operation proved", closedHost, noReasons, allProved, ""},
		{"stale reasons", closedHost, declared, allProved, "a scenario proves `link` of the create mode of `ncly video convert`, so its reason is stale"},
		{"no try of the refused operation", closedHost, declared, "without rename keeps 0 ncly video convert clip.mov\n", "keeps: no call tried `rename` in the scenario folder"},
		{"a refused rename that the run ignores", publicationHost, declared, "cp clip.mov out/clip.txt\nwithout rename keeps 0 ncly video convert clip.mov --force\n", "keeps: the run published out/clip.txt with no link and no rename"},
		{"a dry run that never checks the hard link", publicationHost, declared, "without link refuses CAPABILITY_UNSUPPORTED ncly video convert clip.mov\n", "refuses: dry run exited with 0, want a refusal with CAPABILITY_UNSUPPORTED"},
		{"unknown operation", closedHost, declared, "without copy keeps 0 ncly video convert clip.mov\n", "usage: without link|rename-noreplace|rename"},
		{"signal is no check of without", closedHost, declared, "without link signal TERM started ncly video convert clip.mov\n", "usage: without"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := publicationFixture(t, c.host, c.spec, c.script)
			switch {
			case c.problem == "" && err != nil:
				t.Fatalf("valid scenario failed: %v\n%s", err, out)
			case c.problem != "" && (err == nil || !strings.Contains(out, c.problem)):
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
