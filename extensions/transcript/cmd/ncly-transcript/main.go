// Command ncly-transcript runs the transcript extension as an external
// extension.
package main

import (
	"github.com/pascalandy/nicely/extensions/transcript"
	"github.com/pascalandy/nicely/sdk/extension"
)

func main() { extension.Main(transcript.Extension) }
