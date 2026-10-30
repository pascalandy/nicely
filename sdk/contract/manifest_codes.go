package contract

import (
	"slices"
	"strings"
)

// OwnCode is an error code that an extension declares.
type OwnCode struct {
	Code Code
	Exit ExitCode
}

func (m Manifest) checkCodes() error {
	prefix := m.codePrefix()
	seen := map[Code]bool{}
	for _, e := range m.Errors {
		_, core := e.Code.Exit()
		switch {
		case !strings.HasPrefix(string(e.Code), prefix) || !codeForm.MatchString(string(e.Code)):
			return broken("extension.code_prefix", map[string]any{"Code": string(e.Code), "Prefix": prefix})
		case core || e.Code.IsWarning():
			return broken("extension.code_core", map[string]any{"Code": string(e.Code)})
		case seen[e.Code]:
			return broken("extension.code_twice", map[string]any{"Code": string(e.Code)})
		case !slices.Contains([]ExitCode{ExitRuntime, ExitUsage, ExitHuman}, e.Exit):
			return broken("extension.code_exit", map[string]any{"Code": string(e.Code), "Value": int(e.Exit)})
		}
		seen[e.Code] = true
	}
	return nil
}
