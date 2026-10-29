package skill

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
)

const (
	codeNoSource     contract.Code = "SKILL_NO_SOURCE"
	codeNameMismatch contract.Code = "SKILL_NAME_MISMATCH"
)

// check is one finding of ncly skill status, the check object of ncly
// doctor in core-spec.md.
type check struct {
	ID        string        `json:"id"`
	Component string        `json:"component"`
	Status    string        `json:"status"`
	Code      contract.Code `json:"code,omitempty"`
	Message   string        `json:"message"`
	Hint      string        `json:"hint,omitempty"`
}

// report is the answer of ncly skill status.
type report struct {
	Checks []check `json:"checks"`
}

// kinds are the ids of the checks of status, each with the catalog ID of
// its pass, in the order of the report.
var kinds = []struct{ id, pass string }{
	{"skill.sources", "skill.sources_pass"},
	{"skill.names", "skill.names_pass"},
	{"skill.files", "skill.files_pass"},
}

// status checks the skill sources, as ncly skill status in spec.md defines
// it: one warning per problem, and one pass per kind that finds none. It
// fails only for a broken config, as list does.
func status(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	home, _ := os.UserHomeDir()
	paths := config.Locate(home, os.LookupEnv)
	folders, file, problem := configured(p, home, paths)
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	installed, err := described(ctx)
	if err != nil {
		return contract.Outcome{Errors: []contract.Problem{describeFailed(p)}}
	}
	found := map[string][]check{}
	warn := func(id string, code contract.Code, message, hint string, data map[string]any) {
		found[id] = append(found[id], check{ID: id, Component: "skill", Status: "warn", Code: code, Message: p.T(message, data), Hint: p.T(hint, data)})
	}
	// name reads the name of the skill in a folder, or warns why its
	// SKILL.md is no skill. A folder without SKILL.md has no skill.
	name := func(folder string) (string, bool) {
		path := filepath.Join(folder, "SKILL.md")
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			return "", false
		}
		fill := map[string]any{"File": path}
		var missing *extension.MissingKeyError
		fm, parseErr := extension.ParseSkill(data)
		switch {
		case err != nil:
			warn("skill.files", codeInvalid, "skill.invalid_unreadable", "skill.invalid_hint", fill)
		case errors.As(parseErr, &missing):
			fill["Key"] = missing.Key
			warn("skill.files", codeInvalid, "skill.invalid_missing", "skill.invalid_hint", fill)
		case parseErr != nil:
			warn("skill.files", codeInvalid, "skill.invalid_frontmatter", "skill.invalid_hint", fill)
		default:
			return fm.Name, true
		}
		return "", false
	}
	if len(folders) == 0 {
		warn("skill.sources", codeNoSource, "skill.no_source", "skill.no_source_hint", map[string]any{"File": paths.Shared})
	}
	for i, folder := range folders {
		// A folder that [skill] paths lists twice holds the same problems.
		if slices.Index(folders, folder) != i {
			continue
		}
		entries, err := os.ReadDir(folder)
		if err != nil {
			warn("skill.sources", codeSourceMissing, "skill.source_missing", "skill.source_missing_hint", map[string]any{"Folder": folder, "File": file})
			continue
		}
		for _, entry := range entries {
			path := filepath.Join(folder, entry.Name())
			if info, err := os.Stat(path); err != nil || !info.IsDir() {
				continue
			}
			if n, ok := name(path); ok && n != entry.Name() {
				data := map[string]any{"Path": path, "Name": n, "Folder": entry.Name(), "File": filepath.Join(path, "SKILL.md")}
				warn("skill.names", codeNameMismatch, "skill.name_mismatch", "skill.name_mismatch_hint", data)
			}
		}
	}
	for _, s := range installed {
		if n, ok := name(s.folder); ok && n != s.slug {
			data := map[string]any{"Path": s.folder, "Name": n, "Domain": s.slug, "File": filepath.Join(s.folder, "SKILL.md")}
			warn("skill.names", codeNameMismatch, "skill.domain_mismatch", "skill.domain_mismatch_hint", data)
		}
	}
	r := report{Checks: []check{}}
	for _, k := range kinds {
		if len(found[k.id]) == 0 {
			found[k.id] = []check{{ID: k.id, Component: "skill", Status: "pass", Message: p.T(k.pass)}}
		}
		r.Checks = append(r.Checks, found[k.id]...)
	}
	if !call.JSON {
		writeChecks(call.Stdout, p, r.Checks)
	}
	return contract.Outcome{Data: r}
}

// writeChecks prints the report as ncly doctor does, in the look of sdk/tui.
func writeChecks(w io.Writer, p *i18n.Printer, checks []check) {
	labels := map[string]string{"pass": p.T("skill.status_pass"), "warn": p.T("skill.status_warn")}
	rows := make([]tui.Check, len(checks))
	for i, c := range checks {
		rows[i] = tui.Check{Status: c.Status, Label: labels[c.Status], ID: c.ID, Message: c.Message, Code: string(c.Code), Hint: c.Hint}
	}
	_, _ = io.WriteString(tui.Output(w, noColor()), tui.Checks(rows, p.T("error.hint")))
}

// noColor reports whether the human text goes without styles: core passes
// the resolved --no-color as NO_COLOR, as Global flags in contract.md says.
func noColor() bool { return os.Getenv("NO_COLOR") != "" }
