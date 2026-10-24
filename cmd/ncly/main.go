// Command ncly is Nicely's CLI: core, with its bundled extensions.
package main

import (
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/extensions/skill"
	"github.com/pascalandy/nicely/internal/cli"
	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// bundled lists the extensions compiled into ncly: only what an agent needs
// to operate Nicely, which is skill, as Extensions in contract.md says.
var bundled = []extension.Extension{skill.Extension}

// keys is the keychain's backend: the OS keychain, which the test binary
// replaces with its test keychain.
var keys keychain.Backend = keychain.System{}

func main() { os.Exit(run(bundled)) }

// run starts core with the bundled extensions and the keychain's backend.
// Started under the name of one of them, such as ncly-skill, the binary
// runs that extension instead, as busybox picks its applet from its name:
// core starts a bundled extension that way.
func run(bundled []extension.Extension) int {
	name := filepath.Base(os.Args[0])
	for _, e := range bundled {
		if m, err := contract.ParseManifest(e.Manifest, ""); err == nil && name == "ncly-"+m.Slug() {
			extension.Main(e)
		}
	}
	return cli.Main(os.Args[1:], bundled, keys, os.Stdout, os.Stderr)
}
