package callerenv

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOwnedPathBoundaries(t *testing.T) {
	const base = "github.com/pascalandy/nicely/sdk/"
	dir, cleanup, err := analysistest.WriteFiles(map[string]string{
		base + "extensionrepo/child/env.go": `package child
import "os"
func NewOwnedRun() { _ = os.Environ() } // want "os.Environ"
`,
		base + "extensiontestx/env.go": `package extensiontestx
import "os"
func other() { _ = os.Environ() }
`,
		base + "extensiontest/guard/env.go": `package guard
import "os"
func stage() { _ = os.Environ() }
func other() { _ = os.Environ() } // want "os.Environ"
`,
		base + "extensiontest/fake/env.go": `package fake
import "syscall"
func run() { _ = syscall.Environ() }
func other() { _ = syscall.Environ() } // want "syscall.Environ"
`,
		base + "extensiontest/guard/child/env.go": `package child
import "os"
func stage() { _ = os.Environ() } // want "os.Environ"
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, dir, Analyzer, base+"extensionrepo/child", base+"extensiontestx",
		base+"extensiontest/guard", base+"extensiontest/fake", base+"extensiontest/guard/child")
}
