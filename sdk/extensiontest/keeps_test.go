package extensiontest_test

import (
	"strings"
	"testing"
)

// TestKeeps checks keeps through the fixture of TestPredicts: the exact exit, the protected paths, and the children of the call.
func TestKeeps(t *testing.T) {
	child := strings.Replace(publicationHost, `echo "{\"ok\":true,\"contract_version\":1,\"input\"`, "sleep 30 >/dev/null 2>&1 &\necho \"{\\\"ok\\\":true,\\\"contract_version\\\":1,\\\"input\\\"", 1)
	cases := []struct{ name, host, script, problem string }{
		{"exact exit of a refusal", publicationHost, "protect out\nkeeps 2 ncly video convert missing.mov\n", ""},
		{"exact exit of a success", publicationHost, "keeps 0 ncly-video-convert clip.mov\nexists out/clip.txt\n", ""},
		{"other exit", publicationHost, "keeps 1 ncly video convert missing.mov\n", "keeps: ncly exited with 2, want 1"},
		{"a protected folder changes", publicationHost, "protect out\nkeeps 0 ncly video convert clip.mov\n", "keeps: the call changed protected paths"},
		{"the run changes its input", strings.Replace(publicationHost, `stage=out/.stage.$$`, `echo changed > "$input"; stage=out/.stage.$$`, 1), "keeps 0 ncly video convert clip.mov\n", "keeps: the call changed protected paths: $WORK/clip.mov"},
		{"a child remains", child, "keeps 0 ncly video convert clip.mov\n", "keeps: these processes of the call still ran after it exited"},
		{"answer schema", strings.Replace(publicationHost, `\"ok\":false`, `\"ok\":\"false\"`, 1), "keeps 2 ncly-video-convert missing.mov\n", "breaks the schema"},
		{"path bypass", publicationHost, "keeps 0 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert clip.mov\n", "use the registered command"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := publicationFixture(t, c.host, declared, c.script)
			switch {
			case c.problem == "" && strings.Contains(out, "keeps:"):
				t.Fatalf("valid scenario failed: %v\n%s", err, out)
			case c.problem != "" && (err == nil || !strings.Contains(out, c.problem)):
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}
