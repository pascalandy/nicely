package extensiontest

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
)

type codeObservation struct {
	path, mode, kind string
	code             contract.Code
}

// recordCodes reads only problem arrays from an answer that Observe has validated.
func (c *Coverage) recordCodes(call parsed, line string) {
	if !call.runs {
		return
	}
	mode := "run"
	if dry, _ := call.line.Switch("dry-run"); call.command.Modes.DryRun && dry {
		mode = "dry_run"
	}
	var answer struct {
		Errors   []contract.Problem `json:"errors"`
		Warnings []contract.Problem `json:"warnings"`
		Results  json.RawMessage    `json:"results"`
	}
	if json.Unmarshal([]byte(line), &answer) != nil {
		return
	}
	if call.command.Result.Kind == contract.ResultBatch && mode == "run" {
		var items []struct {
			Errors []contract.Problem `json:"errors"`
		}
		if len(answer.Results) > 0 && json.Unmarshal(answer.Results, &items) != nil {
			return
		}
		for _, item := range items {
			answer.Errors = append(answer.Errors, item.Errors...)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.codes == nil {
		c.codes = map[codeObservation]bool{}
	}
	for _, problem := range answer.Errors {
		c.codes[codeObservation{call.path, mode, "error", problem.Code}] = true
	}
	for _, problem := range answer.Warnings {
		c.codes[codeObservation{call.path, mode, "warning", problem.Code}] = true
	}
}

// codeScope expands the declaration's defaults in command and mode order.
func (c *Coverage) codeScope(d contract.CodeCoverage) []codeObservation {
	kind := "error"
	if d.Warning {
		kind = "warning"
	}
	var scope []codeObservation
	for _, command := range c.manifest.Commands {
		path := strings.Join(command.Path, " ")
		if !c.manifest.Runs(command) || d.Commands != nil && !slices.Contains(d.Commands, path) {
			continue
		}
		modes := []string{"run"}
		if command.Modes.DryRun {
			modes = append(modes, "dry_run")
		}
		for _, mode := range modes {
			if d.Modes == nil || slices.Contains(d.Modes, mode) {
				scope = append(scope, codeObservation{path, mode, kind, d.Code})
			}
		}
	}
	return scope
}

// missingCodes runs under Missing's lock after all scenarios finish.
func (c *Coverage) missingCodes() []string {
	var problems []string
	for _, d := range c.manifest.DeclaredCodes {
		if d.Untested != "" {
			continue
		}
		for _, obligation := range c.codeScope(d) {
			if !c.codes[obligation] {
				problems = append(problems, c.printer.T("extensiontest.code_missing", map[string]any{
					"Position": fmt.Sprintf("%d:%d", d.Line, d.Column), "Kind": obligation.kind, "Code": string(d.Code),
					"Path": obligation.path, "Mode": obligation.mode,
				}))
			}
		}
	}
	return problems
}

// MissingCodes checks declared codes without requiring foreground coverage for commands that a runner exercises only in the background.
func (c *Coverage) MissingCodes() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.missingCodes()
}

// Reservations lists each untested declaration and its reason for the UAT report, in manifest order.
func (c *Coverage) Reservations() []string {
	var reservations []string
	for _, d := range c.manifest.DeclaredCodes {
		if d.Untested == "" {
			continue
		}
		var paths, modes []string
		kind := "error"
		if d.Warning {
			kind = "warning"
		}
		scope := c.codeScope(d)
		for _, obligation := range scope {
			if !slices.Contains(paths, obligation.path) {
				paths = append(paths, obligation.path)
			}
			if !slices.Contains(modes, obligation.mode) {
				modes = append(modes, obligation.mode)
			}
		}
		reservations = append(reservations, c.printer.T("extensiontest.code_untested", map[string]any{
			"Position": fmt.Sprintf("%d:%d", d.Line, d.Column), "Kind": kind, "Code": string(d.Code),
			"Commands": strings.Join(paths, ","), "Modes": strings.Join(modes, ","), "Reason": d.Untested,
		}))
	}
	return reservations
}
