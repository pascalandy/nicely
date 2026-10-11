package cli

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	sdkextension "github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

// extension is a bundled extension, or an installed one that discovery
// found. It stays its domain's extension when it cannot run, so that a
// broken one is reported instead of hidden by a later one.
type extension struct {
	domain string
	// path is where discovery found the executable, links unresolved. A
	// bundled extension runs from the binary of ncly itself.
	path     string
	bundled  bool
	manifest contract.Manifest
	catalogs *i18n.Catalogs
	// broken names the rule that keeps the extension from running, about
	// file. It is nil for a valid extension.
	broken *contract.RuleError
	file   string
}

const executablePrefix = "ncly-"

// host finds the extensions of one run: the bundled ones, which win over
// any other of the same domain, then those of the folders of searchPath.
type host struct {
	bundled []*extension
	dirs    []string
}

func newHost(bundled []sdkextension.Extension) host {
	h := host{dirs: searchPath()}
	for _, b := range bundled {
		h.bundled = append(h.bundled, loadBundled(b))
	}
	return h
}

// reserved reports whether no installed extension may take the domain:
// core's names, and those of the bundled extensions.
func (h host) reserved(domain string) bool {
	return !contract.IsDomain(domain) || slices.Contains(reservedNames, domain) ||
		slices.ContainsFunc(h.bundled, func(e *extension) bool { return e.domain == domain })
}

// searchPath returns the folders where extensions live, in the order that
// discovery reads them: the extensions folder of the data folder, then each
// absolute folder of PATH.
func searchPath() []string {
	home, _ := os.UserHomeDir()
	var dirs []string
	if data := config.Locate(home, os.LookupEnv).Data; data != "" {
		dirs = append(dirs, filepath.Join(data, "extensions"))
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.IsAbs(dir) {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// find returns the extension of one domain, or nil when none has it.
func (h host) find(domain string) *extension {
	for _, e := range h.bundled {
		if e.domain == domain {
			return e
		}
	}
	if h.reserved(domain) {
		return nil
	}
	for _, dir := range h.dirs {
		path := filepath.Join(dir, executablePrefix+domain)
		if isExecutable(path) {
			return loadExtension(domain, path)
		}
	}
	return nil
}

// all returns every extension, the bundled ones and the first one found
// for each other domain, sorted by domain.
func (h host) all() []*extension {
	found := map[string]*extension{}
	for _, e := range h.bundled {
		found[e.domain] = e
	}
	for _, dir := range h.dirs {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			domain, ok := strings.CutPrefix(entry.Name(), executablePrefix)
			if !ok || found[domain] != nil || h.reserved(domain) {
				continue
			}
			if path := filepath.Join(dir, entry.Name()); isExecutable(path) {
				found[domain] = loadExtension(domain, path)
			}
		}
	}
	var all []*extension
	for _, domain := range slices.Sorted(maps.Keys(found)) {
		all = append(all, found[domain])
	}
	return all
}

// isExecutable reports whether path is a regular file, or a link to one,
// that the user may execute.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && unix.Access(path, unix.X_OK) == nil
}

// loadExtension reads the manifest and the catalogs beside the executable.
// It never starts the extension.
func loadExtension(domain, path string) *extension {
	e := &extension{domain: domain, path: path, file: path + ".toml"}
	data, err := os.ReadFile(e.file)
	if err != nil {
		e.broken = reason(err)
		return e
	}
	locales := path + ".locales"
	e.load(data, os.DirFS(locales), filepath.Join(locales, "en.toml"))
	return e
}

// loadBundled reads the manifest and the catalogs that a bundled extension
// embeds. It runs from the binary of ncly, under its own name.
func loadBundled(b sdkextension.Extension) *extension {
	self, _ := os.Executable()
	e := &extension{path: self, bundled: true, file: "manifest.toml"}
	locales, err := fs.Sub(b.Locales, "locales")
	if err != nil {
		e.broken = reason(err)
		return e
	}
	e.load(b.Manifest, locales, "locales/en.toml")
	e.domain = e.manifest.Domain
	return e
}

// load checks the manifest, then the catalogs, and keeps the first rule that
// they break, about the file that breaks it.
func (e *extension) load(manifest []byte, locales fs.FS, catalog string) {
	var err error
	if e.manifest, err = contract.ParseManifest(manifest, e.domain); err != nil {
		e.broken = reason(err)
		return
	}
	e.file = catalog
	if e.catalogs, err = i18n.Load(locales); err != nil {
		e.broken = reason(err)
		return
	}
	if err := e.manifest.CheckCatalog(e.catalogs.Has); err != nil {
		e.broken = reason(err)
	}
}

// reason turns the error of a manifest or a catalog into the rule it breaks.
func reason(err error) *contract.RuleError {
	var (
		rule    *contract.RuleError
		catalog *i18n.CatalogError
	)
	switch {
	case errors.As(err, &rule):
		return rule
	case errors.As(err, &catalog) && catalog.Line > 0:
		return &contract.RuleError{Rule: "extension.file_syntax", Data: map[string]any{"Line": catalog.Line}}
	case errors.Is(err, fs.ErrNotExist):
		return &contract.RuleError{Rule: "extension.file_missing"}
	case errors.Is(err, fs.ErrPermission):
		return &contract.RuleError{Rule: "extension.file_unreadable"}
	}
	return &contract.RuleError{Rule: "extension.file_invalid"}
}

// invalid reports why an extension cannot run, as EXTENSION_INVALID.
func (e *extension) invalid(p *i18n.Printer) contract.Problem {
	file := e.file
	if filepath.IsAbs(file) {
		file, _ = filepath.Rel(filepath.Dir(e.path), file)
	}
	data := map[string]any{"File": file}
	maps.Copy(data, e.broken.Data)
	return contract.Problem{
		Code:    contract.ExtensionInvalid,
		Message: p.T("extension.invalid", map[string]any{"Domain": e.domain, "Reason": p.T(e.broken.Rule, data)}),
		Hint:    p.T("extension.invalid_hint", map[string]any{"Path": e.path}),
	}
}

// extensionGroup is the group of the root's help that lists extensions.
const extensionGroup = "extensions"

// addExtensions adds the commands of each valid extension under the root,
// with the extension's text in the language that lang requests.
func addExtensions(root *cobra.Command, p *i18n.Printer, lang string, extensions []*extension) {
	root.AddGroup(&cobra.Group{ID: extensionGroup, Title: p.T("help.extensions")})
	for _, e := range extensions {
		if e.broken != nil {
			continue
		}
		domain := command.AddManifest(root, p, e.catalogs.Printer(lang), e.manifest, func(d contract.Command) func(*cobra.Command, []string) error {
			return func(c *cobra.Command, args []string) error {
				return &call{extension: e, kind: d.Result.Kind, args: canonicalArgs(d, c, args)}
			}
		})
		domain.GroupID = extensionGroup
	}
}
