// Package skill is the bundled extension that lists the skills that teach
// agents, as extensions/skill/spec.md defines. It imports only sdk/, so it
// could leave the binary of ncly without a rewrite.
package skill

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

// Extension is the skill extension, which cmd/ncly bundles.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Handlers: map[string]extension.Handler{
		"skill list": list,
	},
}

// Skill is one skill of the list. Path is nil for a skill without a folder.
type Skill struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Path        *string `json:"path"`
	Source      string  `json:"source"`
}

// sourceConfig is the source of a skill of a folder of [skill] paths.
const sourceConfig = "config"

const (
	codeShadowed      contract.Code = "SKILL_SHADOWED"
	codeNoSource      contract.Code = "SKILL_NO_SOURCE"
	codeInvalid       contract.Code = "SKILL_INVALID"
	codeSourceMissing contract.Code = "SKILL_SOURCE_MISSING"
)

func list(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	home, _ := os.UserHomeDir()
	paths := config.Locate(home, os.LookupEnv)
	folders, file, problem := configured(p, home, paths)
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	s := &scan{p: p, seen: map[string]bool{}, owner: map[string]string{}, skills: []Skill{}}
	if len(folders) == 0 {
		s.warn(codeNoSource, "skill.no_source", "skill.no_source_hint", map[string]any{"File": paths.Shared})
	}
	for _, folder := range folders {
		s.folder(folder, file)
	}
	slices.SortStableFunc(s.skills, func(a, b Skill) int { return cmp.Compare(a.Name, b.Name) })
	if !call.JSON {
		write(call.Stdout, s.skills)
	}
	return contract.Outcome{Data: map[string]any{"skills": s.skills}, Warnings: s.warnings}
}

// configured reads the folders of [skill] paths, each resolved from the
// home folder, with the file that set them. A config file that cannot be
// read, or a paths that is not a list of absolute or ~/ folders, answers
// CONFIG_INVALID, whose hint names the file.
func configured(p *i18n.Printer, home string, paths config.Paths) ([]string, string, *contract.Problem) {
	invalid := func(id string, data map[string]any) *contract.Problem {
		return &contract.Problem{Code: contract.ConfigInvalid, Message: p.T(id, data), Hint: p.T("skill.config_hint", data)}
	}
	values, err := config.LoadValues(paths)
	var bad *config.InvalidError
	switch {
	case errors.As(err, &bad) && bad.Line > 0:
		return nil, "", invalid("skill.config_invalid_line", map[string]any{"File": bad.File, "Line": bad.Line})
	case errors.As(err, &bad):
		return nil, "", invalid("skill.config_invalid", map[string]any{"File": bad.File})
	case err != nil:
		return nil, "", invalid("skill.config_invalid", map[string]any{"File": paths.Shared})
	}
	var listed []string
	set, err := values.Decode("skill.paths", &listed)
	if err != nil {
		return nil, "", invalid("skill.paths_invalid", map[string]any{"File": set.Source})
	}
	folders := make([]string, 0, len(listed))
	for _, folder := range listed {
		resolved, err := config.Path(home, folder)
		if err != nil {
			return nil, "", invalid("skill.path_relative", map[string]any{"File": set.Source, "Folder": folder})
		}
		folders = append(folders, resolved)
	}
	return folders, set.Source, nil
}

// scan collects the skills of the sources in order, with their warnings.
// owner maps each name to the path of the skill that took it first, and
// seen holds each folder already reached, links resolved.
type scan struct {
	p        *i18n.Printer
	seen     map[string]bool
	owner    map[string]string
	skills   []Skill
	warnings []contract.Problem
}

func (s *scan) warn(code contract.Code, message, hint string, data map[string]any) {
	s.warnings = append(s.warnings, contract.Problem{Code: code, Message: s.p.T(message, data), Hint: s.p.T(hint, data)})
}

// folder adds the skills of one folder of [skill] paths, which file lists:
// each subfolder with a SKILL.md, in the order of their names byte by byte,
// so that the first one keeps a name that two of them declare.
func (s *scan) folder(folder, file string) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		s.warn(codeSourceMissing, "skill.source_missing", "skill.source_missing_hint", map[string]any{"Folder": folder, "File": file})
		return
	}
	for _, entry := range entries {
		path := filepath.Join(folder, entry.Name())
		// A link counts as the folder that it reaches, and a broken link is
		// no skill.
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		if real, err := filepath.EvalSymlinks(path); err == nil {
			if s.seen[real] {
				continue
			}
			s.seen[real] = true
		}
		s.add(path)
	}
}

// add adds the skill of one subfolder, or warns why its SKILL.md is no
// skill. A subfolder without SKILL.md is no skill, and warns nothing.
func (s *scan) add(path string) {
	file := filepath.Join(path, "SKILL.md")
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	fill := map[string]any{"File": file}
	if err != nil {
		s.warn(codeInvalid, "skill.invalid_unreadable", "skill.invalid_hint", fill)
		return
	}
	fm, err := extension.ParseSkill(data)
	var missing *extension.MissingKeyError
	switch {
	case errors.As(err, &missing):
		fill["Key"] = missing.Key
		s.warn(codeInvalid, "skill.invalid_missing", "skill.invalid_hint", fill)
		return
	case err != nil:
		s.warn(codeInvalid, "skill.invalid_frontmatter", "skill.invalid_hint", fill)
		return
	}
	if winner, taken := s.owner[fm.Name]; taken {
		s.warn(codeShadowed, "skill.shadowed", "skill.shadowed_hint", map[string]any{"Name": fm.Name, "Path": path, "Winner": winner, "File": file})
		return
	}
	s.owner[fm.Name] = path
	s.skills = append(s.skills, Skill{Name: fm.Name, Description: fm.Description, Path: &path, Source: sourceConfig})
}

// write prints one line per skill: its name, then its description on one
// line. The SDK reports a failed write.
func write(w io.Writer, skills []Skill) {
	width := 0
	for _, s := range skills {
		width = max(width, len(s.Name))
	}
	var b strings.Builder
	for _, s := range skills {
		fmt.Fprintf(&b, "%-*s  %s\n", width, s.Name, strings.Join(strings.Fields(s.Description), " "))
	}
	_, _ = io.WriteString(w, b.String())
}
