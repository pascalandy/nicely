package skill

import (
	"cmp"
	"embed"
	"io/fs"
	"os"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/extension"
	"golang.org/x/mod/module"
)

// bundled holds the skills that teach an agent to use Nicely: nicely, and
// one skill per topic that it names. They are the data of the bundled
// source, not skills of the skill domain, which ships none. A bundled skill
// has no folder for a relative link, so each topic is a skill of its own.
//
//go:embed nicely*/SKILL.md
var bundled embed.FS

// bundledSkills returns the skills of the bundled source, in the order of
// their folders, without a path.
func bundledSkills() []Skill {
	revision := cmp.Or(os.Getenv("NCLY_VERSION"), command.ModuleVersion())
	if revision == "v0.0.0-dev" || module.IsPseudoVersion(revision) {
		revision = "main"
	}
	files, _ := fs.Glob(bundled, "*/SKILL.md")
	skills := make([]Skill, 0, len(files))
	for _, file := range files {
		data, err := fs.ReadFile(bundled, file)
		if err != nil {
			continue
		}
		if fm, err := extension.ParseSkill(data); err == nil {
			content := strings.ReplaceAll(string(data), "@NCLY_REVISION@", revision)
			skills = append(skills, Skill{Name: fm.Name, Description: fm.Description, Source: sourceBundled, content: content})
		}
	}
	return skills
}
