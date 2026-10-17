// Command ncly-agent runs the agent extension as an external extension.
package main

import (
	"github.com/pascalandy/nicely/extensions/agent"
	"github.com/pascalandy/nicely/sdk/extension"
)

func main() { extension.Main(agent.Extension) }
