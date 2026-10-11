// Command ncly is Nicely's CLI: core, with its bundled extensions.
package main

import (
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/internal/cli"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// bundled lists the extensions compiled into ncly. skill joins it in M03.
var bundled []extension.Extension

func main() { os.Exit(run(bundled)) }

// run starts core with the bundled extensions. Started under the name of
// one of them, such as ncly-skill, the binary runs that extension instead,
// as busybox picks its applet from its name: core starts a bundled
// extension that way.
func run(bundled []extension.Extension) int {
	name := filepath.Base(os.Args[0])
	for _, e := range bundled {
		if m, err := contract.ParseManifest(e.Manifest, ""); err == nil && name == "ncly-"+m.Domain {
			extension.Main(e)
		}
	}
	return cli.Main(os.Args[1:], bundled, os.Stdout, os.Stderr)
}
