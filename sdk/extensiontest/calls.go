package extensiontest

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// Coverage checks answers and records calls across parallel scenarios.
// Other scenario commands use Observe after execution and Missing after all scenarios finish.
type Coverage struct {
	manifest contract.Manifest
	printer  *i18n.Printer
	mu       sync.Mutex
	seen     map[string]bool
	codes    map[codeObservation]bool
}

// NewCoverage prepares the manifest and answer checks that Run uses.
func NewCoverage(e extension.Extension) (*Coverage, error) {
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return nil, err
	}
	locales, err := fs.Sub(e.Locales, "locales")
	if err != nil {
		return nil, err
	}
	catalogs, err := i18n.Load(i18n.Locales(), locales)
	if err != nil {
		return nil, err
	}
	return &Coverage{manifest: m, printer: catalogs.Printer("en"), seen: map[string]bool{}}, nil
}

// record credits a command only after its answer passed validation or its declared text result completed successfully.
func (c *Coverage) record(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen[path] = true
}

// Missing reports commands and declared codes without validated observations, in manifest order.
func (c *Coverage) Missing() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var problems []string
	for _, command := range c.manifest.Commands {
		path := strings.Join(command.Path, " ")
		if c.manifest.Runs(command) && !c.seen[path] {
			problems = append(problems, fmt.Sprintf("%s: no scenario runs `ncly %s`. Fix: add a scenario that runs it with a checked answer, such as %s", scenarios, path, filepath.Join(scenarios, strings.Join(command.Path[1:], "_")+".txtar")))
		}
	}
	return append(problems, c.missingCodes()...)
}
