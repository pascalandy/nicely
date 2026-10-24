// Command export writes the named fixtures' embedded sidecars for verification.
package main

import (
	"log"
	"os"

	"github.com/pascalandy/nicely/sdk/extension"
	convert "github.com/pascalandy/nicely/testdata/extensions/video-convert"
	subtitle "github.com/pascalandy/nicely/testdata/extensions/video-subtitle"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: export DIRECTORY")
	}
	for _, e := range []extension.Extension{convert.Extension, subtitle.Extension} {
		if err := extension.Check(e); err != nil {
			log.Fatal(err)
		}
		if err := extension.Export(e, os.Args[1]); err != nil {
			log.Fatal(err)
		}
	}
}
