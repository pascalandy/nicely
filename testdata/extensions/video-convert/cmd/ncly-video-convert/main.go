// Command ncly-video-convert runs the named conversion fixture.
package main

import (
	"github.com/pascalandy/nicely/sdk/extension"
	video "github.com/pascalandy/nicely/testdata/extensions/video-convert"
)

func main() { extension.Main(video.Extension) }
