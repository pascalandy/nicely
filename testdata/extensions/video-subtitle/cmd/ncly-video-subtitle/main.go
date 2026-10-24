// Command ncly-video-subtitle runs the named completion provider fixture.
package main

import (
	"github.com/pascalandy/nicely/sdk/extension"
	video "github.com/pascalandy/nicely/testdata/extensions/video-subtitle"
)

func main() { extension.Main(video.Extension) }
