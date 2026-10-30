package contract

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2/unstable"
)

// CodeCoverage scopes a declared error or warning to commands and execution modes.
// Nil Commands selects every command that runs. Nil Modes selects each command's supported modes.
// Untested gives a reason and an issue URL to defer scenario coverage.
// Line and Column identify the declaration's TOML key or inline table, starting at one.
type CodeCoverage struct {
	Code     Code
	Warning  bool
	Commands []string
	Modes    []string
	Untested string
	Line     int
	Column   int
}

type declaredCodeFile struct {
	Code     string   `toml:"code"`
	Exit     *int     `toml:"exit"`
	Commands []string `toml:"commands"`
	Modes    []string `toml:"modes"`
	Untested *string  `toml:"untested"`
}

// parseDeclaredCodes joins decoded declarations to positions from the TOML syntax tree.
// The manifest decoder has already checked the document's syntax and field types.
func parseDeclaredCodes(data []byte, file manifestFile) ([]CodeCoverage, error) {
	var declarations []CodeCoverage
	var parser unstable.Parser
	parser.Reset(data)
	counts := map[string]int{}
	add := func(kind string, raw unstable.Range) error {
		entries := file.Errors
		if kind == "warnings" {
			entries = file.Warnings
		}
		if counts[kind] >= len(entries) {
			return broken("extension.file_invalid", nil)
		}
		entry := entries[counts[kind]]
		counts[kind]++
		position := parser.Shape(raw).Start
		d := CodeCoverage{
			Code: Code(entry.Code), Warning: kind == "warnings", Commands: entry.Commands, Modes: entry.Modes,
			Line: position.Line, Column: position.Column,
		}
		if d.Warning && entry.Exit != nil {
			return d.broken("extension.warning_exit", nil)
		}
		if entry.Untested != nil {
			if !hasIssueReason(*entry.Untested) {
				return d.broken("extension.code_untested", nil)
			}
			d.Untested = *entry.Untested
		}
		declarations = append(declarations, d)
		return nil
	}
	root := true
	for parser.NextExpression() {
		node := parser.Expression()
		if node.Kind == unstable.Table || node.Kind == unstable.ArrayTable {
			root = false
		}
		key := node.Key()
		if !key.Next() || !key.IsLast() {
			continue
		}
		kind := strings.ToLower(string(key.Node().Data))
		if kind != "errors" && kind != "warnings" {
			continue
		}
		switch {
		case node.Kind == unstable.ArrayTable:
			if err := add(kind, key.Node().Raw); err != nil {
				return nil, err
			}
		case root && node.Kind == unstable.KeyValue:
			for entries := node.Value().Children(); entries.Next(); {
				if err := add(kind, entries.Node().Raw); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := parser.Error(); err != nil {
		return nil, broken("extension.file_invalid", nil)
	}
	if counts["errors"] != len(file.Errors) || counts["warnings"] != len(file.Warnings) {
		return nil, broken("extension.file_invalid", nil)
	}
	return declarations, nil
}

func (d CodeCoverage) broken(rule string, values map[string]any) *RuleError {
	if values == nil {
		values = map[string]any{}
	}
	values["Code"], values["Line"], values["Column"] = string(d.Code), d.Line, d.Column
	return broken(rule, values)
}

func (m Manifest) checkCodeScopes() error {
	seen := map[Code]bool{}
	for _, d := range m.DeclaredCodes {
		if seen[d.Code] {
			return d.broken("extension.declared_code_twice", nil)
		}
		seen[d.Code] = true
		if d.Warning {
			if !strings.HasPrefix(string(d.Code), m.codePrefix()) || !codeForm.MatchString(string(d.Code)) {
				return d.broken("extension.warning_prefix", map[string]any{"Prefix": m.codePrefix()})
			}
			if _, core := Core.Exit(d.Code); core || d.Code.IsWarning() {
				return d.broken("extension.warning_core", nil)
			}
		}
		if d.Commands != nil && len(d.Commands) == 0 {
			return d.broken("extension.code_commands_empty", nil)
		}
		for i, path := range d.Commands {
			c, exists := m.Command(path)
			if !exists || !m.Runs(c) {
				return d.broken("extension.code_command", map[string]any{"Path": path})
			}
			if slices.Index(d.Commands, path) != i {
				return d.broken("extension.code_command_twice", map[string]any{"Path": path})
			}
		}
		if d.Modes != nil && len(d.Modes) == 0 {
			return d.broken("extension.code_modes_empty", nil)
		}
		for i, mode := range d.Modes {
			if mode != "run" && mode != "dry_run" {
				return d.broken("extension.code_mode", map[string]any{"Mode": mode})
			}
			if slices.Index(d.Modes, mode) != i {
				return d.broken("extension.code_mode_twice", map[string]any{"Mode": mode})
			}
			for _, c := range m.Commands {
				path := strings.Join(c.Path, " ")
				if mode == "dry_run" && m.Runs(c) && !c.Modes.DryRun && (d.Commands == nil || slices.Contains(d.Commands, path)) {
					return d.broken("extension.code_dry_run", map[string]any{"Path": path})
				}
			}
		}
	}
	return nil
}

var issuePath = regexp.MustCompile(`/issues/[1-9][0-9]*$`)

func hasIssueReason(text string) bool {
	for _, word := range strings.Fields(text) {
		link := strings.Trim(word, "()[]<>,.;\"")
		parsed, err := url.Parse(link)
		if err != nil || parsed.Host == "" || parsed.Scheme != "https" && parsed.Scheme != "http" || !issuePath.MatchString(parsed.Path) {
			continue
		}
		if strings.Trim(strings.Replace(text, word, "", 1), " \t\r\n:,-.()[]<>") != "" {
			return true
		}
	}
	return false
}
