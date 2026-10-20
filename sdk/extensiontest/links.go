package extensiontest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rogpeppe/go-internal/txtar"
	"go.yaml.in/yaml/v3"
)

// links returns a problem for each broken link that the files show,
// before any scenario runs:
//   - a missing skill
//   - a command that no scenario runs
//   - a skill that names a command that the manifest lacks
//   - a call that skips the answer check
//   - two Nicely versions
func (k *kit) links() []string {
	var problems []string
	if k.ext.Skill == nil {
		problems = append(problems, fmt.Sprintf("skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named %s, and embed it as the Skill of extension.Extension", k.manifest.Domain))
	}
	problems = append(problems, k.scenarioLinks()...)
	problems = append(problems, k.skillLinks()...)
	if problem := k.versionLink(); problem != "" {
		problems = append(problems, problem)
	}
	return problems
}

// scenarioLinks reports each command that runs and that no scenario calls
// as ncly <domain> .... It also reports each call of ncly that exec runs in
// the foreground, whose answer the kit cannot check.
func (k *kit) scenarioLinks() []string {
	dir := filepath.Join("testdata", "script")
	files, err := os.ReadDir(filepath.Join(k.root, dir))
	if err != nil && !os.IsNotExist(err) {
		return []string{err.Error()}
	}
	covered := map[string]bool{}
	var problems []string
	exec := regexp.MustCompile(`^(?:\[[^\]]*\]\s*)*[!?]?\s*exec\s+ncly\s+` + regexp.QuoteMeta(k.manifest.Domain) + `\b`)
	background := regexp.MustCompile(`\s&(?:[a-zA-Z_0-9]+&)?$`)
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || filepath.Ext(name) != ".txtar" && filepath.Ext(name) != ".txt" {
			continue
		}
		archive, err := txtar.ParseFile(filepath.Join(k.root, dir, name))
		if err != nil {
			return []string{err.Error()}
		}
		for i, line := range strings.Split(string(archive.Comment), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			for _, path := range k.mentions(line) {
				covered[path] = true
			}
			if exec.MatchString(line) && !background.MatchString(line) {
				problems = append(problems, fmt.Sprintf("%s:%d: exec runs ncly, so the kit cannot check its answer. Fix: drop exec, and run the kit's ncly command", filepath.Join(dir, name), i+1))
			}
		}
	}
	for _, c := range k.manifest.Commands {
		path := strings.Join(c.Path, " ")
		if k.manifest.Runs(c) && !covered[path] {
			problems = append(problems, fmt.Sprintf("%s: no scenario runs `ncly %s`. Fix: add a scenario that runs it, such as %s", dir, path, filepath.Join(dir, strings.Join(c.Path[1:], "_")+".txtar")))
		}
	}
	return problems
}

// skillLinks reports each ncly <domain> ... command that a file of the
// skill names and that the manifest lacks. Core's commands and other
// domains stay free to mention.
func (k *kit) skillLinks() []string {
	if k.ext.Skill == nil {
		return nil
	}
	var problems []string
	err := fs.WalkDir(k.ext.Skill, "skill", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(k.ext.Skill, path)
		if err != nil {
			return err
		}
		for _, named := range k.mentions(string(data)) {
			if _, declared := k.manifest.Command(named); !declared {
				problems = append(problems, fmt.Sprintf("%s: names `ncly %s`, which the manifest lacks. Fix: declare that command in manifest.toml, or correct the skill", path, named))
			}
		}
		return nil
	})
	if err != nil {
		problems = append(problems, err.Error())
	}
	return problems
}

// mentions returns the command path of each ncly <domain> ... in text, as
// literal words, without parsing shell or Markdown. After a path that
// groups other commands, the next word is a command, so a mistyped one
// shows. After any other path, the next words are its arguments.
func (k *kit) mentions(text string) []string {
	domain := k.manifest.Domain
	groups := map[string]bool{}
	for _, c := range k.manifest.Commands {
		if len(c.Path) > 1 {
			groups[strings.Join(c.Path[:len(c.Path)-1], " ")] = true
		}
	}
	pattern := regexp.MustCompile(`\bncly ` + regexp.QuoteMeta(domain) + `\b((?: [a-z][a-z0-9-]*)*)`)
	var paths []string
	for _, m := range pattern.FindAllStringSubmatch(text, -1) {
		path := domain
		for _, word := range strings.Fields(m[1]) {
			if !groups[path] {
				break
			}
			path += " " + word
		}
		paths = append(paths, path)
	}
	return paths
}

// versionLink reports a .copier-answers.yml whose Nicely version, the
// _commit that the template came from, is not the one that go.mod pins.
func (k *kit) versionLink() string {
	data, err := os.ReadFile(filepath.Join(k.root, ".copier-answers.yml"))
	if os.IsNotExist(err) {
		return ""
	}
	var answers struct {
		Commit string `yaml:"_commit"`
	}
	if err == nil {
		err = yaml.Unmarshal(data, &answers)
	}
	pinned, goErr := goCommand(k.root, "list", "-m", "-f", "{{.Version}}", "github.com/pascalandy/nicely")
	switch {
	case err != nil:
		return ".copier-answers.yml: " + err.Error()
	case goErr != nil:
		return "go.mod: " + goErr.Error()
	case answers.Commit != pinned:
		return fmt.Sprintf(".copier-answers.yml: the template comes from Nicely %q, and go.mod pins Nicely %s. Fix: run just update %s", answers.Commit, pinned, pinned)
	}
	return ""
}
