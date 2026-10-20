// Command ncly-standalone runs the standalone extension.
package main

import (
	"example.com/standalone"
	"github.com/pascalandy/nicely/sdk/extension"
)

func main() { extension.Main(standalone.Extension) }
