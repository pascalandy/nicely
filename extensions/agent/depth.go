package agent

import (
	"os"
	"strconv"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// defaultMaxDepth applies when [agent] max_depth sets no depth.
const defaultMaxDepth = 2

// depthVariable holds how many agents run above this call, 0 when unset.
// Codex receives one more, as Depth limit in spec.md defines.
const depthVariable = "NCLY_AGENT_DEPTH"

// childDepth returns the depth of Codex: one more than NCLY_AGENT_DEPTH, 0
// when unset. A call that runs at max_depth or deeper answers
// AGENT_DEPTH_LIMIT before any other check, so a chain of agents stops
// before it prepares another one. A value that is no whole number from 0
// answers USAGE_INVALID.
func childDepth(p *i18n.Printer, s settings) (int, *contract.Problem) {
	value := os.Getenv(depthVariable)
	depth := 0
	if value != "" {
		var err error
		if depth, err = strconv.Atoi(value); err != nil || depth < 0 {
			return 0, usage(p, "agent.run.depth_invalid", "agent.run.depth_invalid_hint", map[string]any{"Value": value})
		}
	}
	if int64(depth) >= s.maxDepth {
		data := map[string]any{"Max": p.Number(float64(s.maxDepth), 0), "File": s.file}
		if s.depthFile != config.Default {
			data["File"] = s.depthFile
		}
		return 0, &contract.Problem{Code: codeDepthLimit, Message: p.N("agent.run.depth_limit", depth, data), Hint: p.T("agent.run.depth_limit_hint", data)}
	}
	return depth + 1, nil
}

// codeDepthLimit is the extension's own code for a call at max_depth, as
// its manifest declares.
const codeDepthLimit contract.Code = "AGENT_DEPTH_LIMIT"
