package cases

import (
	"os"

	"github.com/rogpeppe/go-internal/testscript"
)

// lifecycle repeats checkInstallLifecycle of sdk/extensiontest/install.go at
// the parent of 09809608. Its scratch install replaced HOME and the XDG
// folders. It kept each other setting of the caller, such as an invalid
// NCLY_CONFIG, which failed ncly describe.
func lifecycle(work string) *testscript.Env {
	return &testscript.Env{WorkDir: work, Vars: os.Environ()} // want `os.Environ: build the environment with extensionrepo.NewOwnedRun`
}
