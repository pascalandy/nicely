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
//   - a line of Failures or Acceptance in spec.md that no test claims
//   - two Nicely versions
func (k *kit) links() []string {
	var problems []string
	if k.ext.Skill == nil {
		problems = append(problems, fmt.Sprintf("skill/SKILL.md: the extension ships no skill. Fix: add skill/SKILL.md, named %s, and embed it as the Skill of extension.Extension", k.manifest.Slug()))
	}
	if scripts, err := k.scripts(); err != nil {
		problems = append(problems, err.Error())
	} else {
		problems = append(problems, k.scenarioLinks(scripts)...)
		problems = append(problems, k.specLinks(scripts)...)
	}
	problems = append(problems, k.skillLinks()...)
	if problem := k.versionLink(); problem != "" {
		problems = append(problems, problem)
	}
	return problems
}

// A script is the script part of a scenario: the lines before its first
// file, which testscript runs. Its name is its path in the repository.
type script struct{ name, text string }

// scenarios is the folder of the scenarios in the repository.
var scenarios = filepath.Join("testdata", "script")

// scripts reads the script of each scenario of testdata/script/.
func (k *kit) scripts() ([]script, error) {
	files, err := os.ReadDir(filepath.Join(k.root, scenarios))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var scripts []script
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || filepath.Ext(name) != ".txtar" && filepath.Ext(name) != ".txt" {
			continue
		}
		archive, err := txtar.ParseFile(filepath.Join(k.root, scenarios, name))
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, script{name: filepath.Join(scenarios, name), text: string(archive.Comment)})
	}
	return scripts, nil
}

// scenarioLinks reports each command that runs and that no scenario calls
// as ncly <root> .... It also reports each foreground exec of ncly or of
// ncly-<slug>, by name or by path, whose answer the kit cannot check.
func (k *kit) scenarioLinks(scripts []script) []string {
	covered := map[string]bool{}
	var problems []string
	direct := "ncly-" + k.manifest.Slug()
	for _, s := range scripts {
		for i, line := range strings.Split(s.text, "\n") {
			args := words(line)
			for _, path := range k.mentions(strings.Join(args, " ")) {
				covered[path] = true
			}
			if program := foregroundExec(args); program == "ncly" || program == direct {
				problems = append(problems, fmt.Sprintf("%s:%d: exec runs %s, so the kit cannot check its answer. Fix: drop exec, and run the kit's command %s", s.name, i+1, program, program))
			}
		}
	}
	for _, c := range k.manifest.Commands {
		path := strings.Join(c.Path, " ")
		if k.manifest.Runs(c) && !covered[path] {
			problems = append(problems, fmt.Sprintf("%s: no scenario runs `ncly %s`. Fix: add a scenario that runs it, such as %s", scenarios, path, filepath.Join(scenarios, strings.Join(c.Path[1:], "_")+".txtar")))
		}
	}
	return problems
}

// words splits a script line into its words as testscript does, but keeps
// each $variable as written. Single quotes wrap text, two quotes inside them
// stand for one, and a # outside them starts a comment.
func words(line string) []string {
	var args []string
	var word strings.Builder
	inWord, quoted := false, false
	for i := 0; i < len(line); i++ {
		switch c := line[i]; {
		case quoted && c == '\'' && i+1 < len(line) && line[i+1] == '\'':
			word.WriteByte(c)
			i++
		case c == '\'':
			inWord, quoted = true, !quoted
		case !quoted && strings.IndexByte(" \t\r#", c) >= 0:
			if inWord {
				args = append(args, word.String())
				word.Reset()
				inWord = false
			}
			if c == '#' {
				return args
			}
		default:
			word.WriteByte(c)
			inWord = true
		}
	}
	if inWord {
		args = append(args, word.String())
	}
	return args
}

// background matches the last word of an exec that testscript runs in the
// background: & or &name&.
var background = regexp.MustCompile(`^&([a-zA-Z_0-9]+&)?$`)

// foregroundExec returns the base name of the program that the words of a
// script line run with exec in the foreground, after any [condition] and !.
// It returns "" for any other line.
func foregroundExec(args []string) string {
	for len(args) > 0 && strings.HasPrefix(args[0], "[") && strings.HasSuffix(args[0], "]") {
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "!" {
		args = args[1:]
	}
	if len(args) < 2 || args[0] != "exec" || background.MatchString(args[len(args)-1]) {
		return ""
	}
	return filepath.Base(args[1])
}

// skillLinks reports each ncly <root> ... command that a file of the
// skill names and that the manifest lacks. Core's commands and other
// extension roots stay free to mention.
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

// mentions returns the command path of each ncly <root> ... in text, as
// literal words, without parsing shell or Markdown. After a path that only
// groups other commands, the next word is a command, so a mistyped one
// shows. After a path that runs, the next word is a command only when the
// manifest declares it, and an argument otherwise.
func (k *kit) mentions(text string) []string {
	root := strings.Join(k.manifest.Root(), " ")
	declared, groups, runs := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, c := range k.manifest.Commands {
		path := strings.Join(c.Path, " ")
		declared[path], runs[path] = true, k.manifest.Runs(c)
		if len(c.Path) > 1 {
			groups[strings.Join(c.Path[:len(c.Path)-1], " ")] = true
		}
	}
	pattern := regexp.MustCompile(`\bncly ` + regexp.QuoteMeta(root) + `((?: [a-z][a-z0-9-]*)*)(?:[^a-z0-9-]|$)`)
	var paths []string
	for _, m := range pattern.FindAllStringSubmatch(text, -1) {
		path := root
		for _, word := range strings.Fields(m[1]) {
			next := path + " " + word
			if !groups[path] || runs[path] && !declared[next] {
				break
			}
			path = next
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
