// Command ncly-video-convert runs the named test kit fixture.
package main

import (
	video "example.com/ncly-video-convert"
	"github.com/pascalandy/nicely/sdk/extension"
)

func main() { extension.Main(video.Extension) }
