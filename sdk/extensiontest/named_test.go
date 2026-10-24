package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

func TestNamedRepositoryChecks(t *testing.T) {
	cases := []struct {
		name  string
		edits []edit
		want  string
	}{
		{
			name:  "a sibling cannot cover the root",
			edits: []edit{{file: "testdata/script/convert.txtar", new: "ncly video subtitle clip.mov --json\n"}},
			want:  "testdata/script: no scenario runs `ncly video convert`. Fix: add a scenario that runs it",
		},
		{
			name: "a direct answer must match the root schema",
			edits: []edit{
				{"manifest.toml", `file = { type = "string" }`, `file = { type = "integer" }`},
				{file: "testdata/script/convert.txtar", new: "ncly-video-convert clip.mov --json\n"},
			},
			want: "`ncly-video-convert clip.mov --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly video convert`",
		},
		{
			name:  "a host answer must match the root schema",
			edits: []edit{{"manifest.toml", `file = { type = "string" }`, `file = { type = "integer" }`}},
			want:  "`ncly video convert clip.mov --json`: the answer on stdout breaks the schema that manifest.toml declares for `ncly video convert`",
		},
		{
			name:  "a direct foreground exec cannot skip the answer check",
			edits: []edit{{file: "testdata/script/convert.txtar", new: "exec $XDG_DATA_HOME/nicely/extensions/ncly-video-convert clip.mov --json\n"}},
			want:  "convert.txtar:1: exec runs ncly-video-convert, so the kit cannot check its answer. Fix: drop exec",
		},
		{
			name:  "a missing skill names the slug",
			edits: []edit{{"extension.go", "\tSkill:    skill,\n", ""}},
			want:  "skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named video-convert",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := goTest(t, "standalone-named", os.Getenv("PATH"), c.edits...)
			if err == nil || !strings.Contains(out, c.want) {
				t.Fatalf("go test error = %v, want failure with %q:\n%s", err, c.want, out)
			}
		})
	}
}
