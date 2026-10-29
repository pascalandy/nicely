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
	"github.com/pascalandy/nicely/sdk/tui"
	"golang.org/x/term"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

// nicely is the skill that teaches an agent to use Nicely: the data of the
// bundled source, not a skill of the skill domain, which ships none.
//
//go:embed nicely/SKILL.md
var nicely []byte

// Extension is the skill extension, which cmd/ncly bundles.
var Extension = extension.Extension{
	Manifest: manifest,
	Locales:  locales,
	Handlers: map[string]extension.Handler{
		"skill list":   list,
		"skill view":   view,
		"skill status": status,
	},
}

// Skill is one skill of the list. Path is nil for a skill without a folder.
type Skill struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Path        *string `json:"path"`
	Source      string  `json:"source"`
	// content is the SKILL.md, which view prints.
	content string
}

// The sources of skills, in the order that list reads them.
const (
	sourceConfig    = "config"
	sourceBundled   = "bundled"
	sourceExtension = "extension"
)

const (
	codeShadowed      contract.Code = "SKILL_SHADOWED"
	codeInvalid       contract.Code = "SKILL_INVALID"
	codeSourceMissing contract.Code = "SKILL_SOURCE_MISSING"
)

func list(ctx context.Context, call extension.Call) contract.Outcome {
	skills, warnings, problem := gather(ctx, call.Printer)
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	if !call.JSON {
		write(call.Stdout, skills)
	}
	return contract.Outcome{Data: map[string]any{"skills": skills}, Warnings: warnings}
}

// view prints the SKILL.md of one skill that list returns, after a line
// that gives its folder, so that the relative paths inside it resolve. It
// adds none of the warnings of list, which describe other skills.
func view(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	skills, _, problem := gather(ctx, p)
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	name := call.Args[0]
	i := slices.IndexFunc(skills, func(s Skill) bool { return s.Name == name })
	if i < 0 {
		return contract.Outcome{Errors: []contract.Problem{{
			Code:    contract.NotFound,
			Message: p.T("skill.not_found", map[string]any{"Name": name}),
			Hint:    "ncly skill list",
			Next:    []string{"ncly", "skill", "list"},
		}}}
	}
	s := skills[i]
	if !call.JSON && call.Mode == contract.Interactive {
		render(call.Stdout, p, s)
	} else if !call.JSON {
		if s.Path != nil {
			_, _ = fmt.Fprintf(call.Stdout, "<!-- skill-dir: %s -->\n", *s.Path)
		}
		_, _ = io.WriteString(call.Stdout, s.content)
		if !strings.HasSuffix(s.content, "\n") {
			_, _ = io.WriteString(call.Stdout, "\n")
		}
	}
	return contract.Outcome{Data: map[string]any{"name": s.Name, "path": s.Path, "source": s.Source, "content": s.content}}
}

// gather returns the skills of every source, sorted by name byte by byte,
// with the warnings of list, or the problem of a broken config or of a
// failed ncly describe.
func gather(ctx context.Context, p *i18n.Printer) ([]Skill, []contract.Problem, *contract.Problem) {
	home, _ := os.UserHomeDir()
	folders, file, problem := configured(p, home, config.Locate(home, os.LookupEnv))
	if problem != nil {
		return nil, nil, problem
	}
	installed, err := described(ctx)
	if err != nil {
		problem := describeFailed(p)
		return nil, nil, &problem
	}
	s := &scan{p: p, seen: map[string]bool{}, owner: map[string]*string{}, skills: []Skill{}}
	for _, folder := range folders {
		s.folder(folder, file)
	}
	if fm, err := extension.ParseSkill(nicely); err == nil {
		s.keep(Skill{Name: fm.Name, Description: fm.Description, Source: sourceBundled, content: string(nicely)})
	}
	for _, skill := range installed {
		s.add(skill.folder, sourceExtension)
	}
	slices.SortStableFunc(s.skills, func(a, b Skill) int { return cmp.Compare(a.Name, b.Name) })
	return s.skills, s.warnings, nil
}

// describeFailed is the failure of a command that could not read the
// skills of the installed extensions from ncly describe.
func describeFailed(p *i18n.Printer) contract.Problem {
	return contract.Problem{
		Code:    contract.Runtime,
		Message: p.T("skill.describe_failed"),
		Hint:    "ncly describe --json",
		Next:    []string{"ncly", "describe", "--json"},
	}
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
// owner maps each name to the path of the skill that took it first, nil
// for the bundled one, and seen holds each folder already reached, links
// resolved.
type scan struct {
	p        *i18n.Printer
	seen     map[string]bool
	owner    map[string]*string
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
		s.add(path, sourceConfig)
	}
}

// add adds the skill of one folder from a source, or warns why its SKILL.md
// is no skill. A folder without SKILL.md is no skill, and warns nothing. A
// folder already reached, links resolved, is the same skill, and warns
// nothing either.
func (s *scan) add(path, source string) {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		if s.seen[real] {
			return
		}
		s.seen[real] = true
	}
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
	s.keep(Skill{Name: fm.Name, Description: fm.Description, Path: &path, Source: source, content: string(data)})
}

// keep adds a skill whose name no earlier source took, or warns that the
// skill that took it first hides this one.
func (s *scan) keep(k Skill) {
	winner, taken := s.owner[k.Name]
	if !taken {
		s.owner[k.Name] = k.Path
		s.skills = append(s.skills, k)
		return
	}
	data := map[string]any{"Name": k.Name}
	switch {
	case k.Path == nil:
		data["Winner"], data["File"] = *winner, filepath.Join(*winner, "SKILL.md")
		s.warn(codeShadowed, "skill.bundled_shadowed", "skill.bundled_shadowed_hint", data)
	case winner == nil:
		data["Path"], data["File"] = *k.Path, filepath.Join(*k.Path, "SKILL.md")
		s.warn(codeShadowed, "skill.shadowed_by_bundled", "skill.shadowed_hint", data)
	default:
		data["Path"], data["Winner"], data["File"] = *k.Path, *winner, filepath.Join(*k.Path, "SKILL.md")
		s.warn(codeShadowed, "skill.shadowed", "skill.shadowed_hint", data)
	}
}

// write prints one line per skill, as the help prints its commands: its
// name, in bold, and a colon, then its description on one line. The SDK
// reports a failed write.
func write(w io.Writer, skills []Skill) {
	entries := make([]tui.Entry, len(skills))
	for i, s := range skills {
		entries[i] = tui.Entry{Name: s.Name, Summary: strings.Join(strings.Fields(s.Description), " ")}
	}
	_, _ = io.WriteString(tui.Output(w, noColor()), tui.Names(entries))
}

// render prints a skill in interactive mode: its folder as a first line,
// then its Markdown without the frontmatter, which Glamour would draw as a
// rule and a heading, in the look of sdk/tui, wrapped at the width of the
// terminal. The SDK reports a failed write.
func render(w io.Writer, p *i18n.Printer, s Skill) {
	width := 80
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		if columns, _, err := term.GetSize(int(f.Fd())); err == nil && columns > 0 {
			width = columns
		}
	}
	text, err := tui.Markdown(body(s.content), width, noColor())
	if err != nil {
		text = s.content
	}
	out := tui.Output(w, noColor())
	if s.Path != nil {
		_, _ = io.WriteString(out, tui.Muted(p.T("skill.folder"))+"  "+*s.Path+"\n\n")
	}
	_, _ = io.WriteString(out, text)
}

// body returns a SKILL.md without its frontmatter, which gather has checked.
func body(content string) string {
	lines := strings.SplitAfter(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r\n") == "---" {
			return strings.Join(lines[i+1:], "")
		}
	}
	return content
}
