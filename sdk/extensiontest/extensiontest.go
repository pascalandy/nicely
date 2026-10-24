// Package extensiontest is the test kit of an extension repository, as Test
// kit in contract.md defines. One test of the repository calls Run. Run
// checks the links between the manifest, the skill, the scenarios, and the
// Nicely versions. It then installs the real ncly-<slug> and the ncly of
// go tool ncly in scratch folders, and runs the scenarios of testdata/script/.
//
// A scenario runs the extension through the commands that the kit adds, so
// that the kit checks each answer against the schema of its command:
//
//	ncly <arg>...                 runs ncly, which exits 0, or not 0 after !
//	ncly-<slug> <arg>...          runs the extension directly, the same way
//	exits <code> <command> <arg>  runs a command, these two included, and checks its exact exit code
//	grant <service>               grants the key of a declared service to the extension
//	fake <name>                   puts the fake child in $WORK/bin as name
//
// The rules of its code:
//   - No case reads the user's keychain, HOME, or XDG folders, or finds the
//     user's extensions. Each scenario gets a scratch HOME and scratch XDG
//     folders, and a PATH without the folders that hold an ncly-*
//     file. Each service that the manifest declares gets the sentinel key
//     sentinel-<service> in its variable, so ncly never reads the keychain
//     for it. On Linux, D-Bus names no bus, so a read that a scenario forces
//     finds no keychain
//   - The kit imports no package of core's internal/, which an extension
//     repository cannot import either
//   - The kit builds every program that it runs from Go source: the
//     extension, ncly, and the fake child of the fake/ folder
//   - Each failure names its file, the missing link, and its fix
package extensiontest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/rogpeppe/go-internal/testscript"
)

// Run tests the extension repository that holds the current folder, whose
// extension is e. It reports each broken link, then runs each scenario as a
// subtest of t.
//
//	func TestExtension(t *testing.T) { extensiontest.Run(t, Extension) }
func Run(t *testing.T, e extension.Extension) {
	t.Helper()
	if err := extension.Check(e); err != nil {
		t.Fatalf("extension.Check: %v. Fix the extension until extension.Check accepts it", err)
	}
	k, err := load(e)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range k.links() {
		t.Error(problem)
	}
	k.build(t)
	testscript.Run(t, testscript.Params{
		Dir:                filepath.Join(k.root, "testdata", "script"),
		RequireUniqueNames: true,
		Setup:              k.setup,
		Cmds:               k.commands(),
	})
}

// kit is what Run knows of one extension repository.
type kit struct {
	ext      extension.Extension
	manifest contract.Manifest
	printer  *i18n.Printer
	// root is the folder of the repository's go.mod.
	root string
	// bin holds ncly, which go tool ncly built.
	bin string
	// path is the PATH of the caller without its extensions.
	path string
	// built is the ncly-<slug> that go build made, and fake the fake
	// child.
	built, fake string
}

func load(e extension.Extension) (*kit, error) {
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
	gomod, err := goCommand("", "env", "GOMOD")
	if err == nil && (gomod == "" || gomod == os.DevNull) {
		err = errors.New("go test runs outside a Go module. Fix: run it in the extension repository, whose go.mod pins Nicely")
	}
	if err != nil {
		return nil, err
	}
	return &kit{ext: e, manifest: m, printer: catalogs.Printer("en"), root: filepath.Dir(gomod), path: isolate(os.Getenv("PATH"))}, nil
}

// build makes the programs that the scenarios run. ncly is the one that go
// tool ncly builds from the Nicely version of go.mod. ncly-<slug> comes
// from cmd/ncly-<slug>/, and the fake child from fake/.
func (k *kit) build(t *testing.T) {
	dir := t.TempDir()
	k.bin = filepath.Join(dir, "bin")
	if err := os.Mkdir(k.bin, 0o755); err != nil {
		t.Fatal(err)
	}
	host, err := goCommand(k.root, "tool", "-n", "ncly")
	if err == nil {
		err = link(host, filepath.Join(k.bin, "ncly"))
	}
	if err != nil {
		t.Fatalf("go.mod: go tool ncly fails: %v. Fix: run go get -tool github.com/pascalandy/nicely/cmd/ncly", err)
	}
	main := "./cmd/ncly-" + k.manifest.Slug()
	k.built = filepath.Join(dir, "ncly-"+k.manifest.Slug())
	if _, err := goCommand(k.root, "build", "-o", k.built, main); err != nil {
		t.Fatalf("%s: go build fails: %v. Fix: give %s a main.go whose main calls extension.Main", main, err, main)
	}
	k.fake = filepath.Join(dir, "fake")
	if _, err := goCommand(k.root, "build", "-o", k.fake, "github.com/pascalandy/nicely/sdk/extensiontest/fake"); err != nil {
		t.Fatalf("go build of the fake child fails: %v", err)
	}
}

// setup gives a scenario its scratch folders, its sentinel keys, and the
// extension, installed with the sidecars that Export writes, the
// executable last.
func (k *kit) setup(env *testscript.Env) error {
	work := env.WorkDir
	vars := map[string]string{
		"HOME":            work,
		"XDG_CONFIG_HOME": filepath.Join(work, ".config"),
		"XDG_DATA_HOME":   filepath.Join(work, ".local", "share"),
		"XDG_STATE_HOME":  filepath.Join(work, ".local", "state"),
		"XDG_CACHE_HOME":  filepath.Join(work, ".cache"),
		"PATH":            strings.Join([]string{filepath.Join(work, "bin"), k.bin, k.path}, string(filepath.ListSeparator)),
		// The keychain of Linux answers through the D-Bus session bus, which
		// this address cannot reach.
		"DBUS_SESSION_BUS_ADDRESS": "disabled:",
	}
	for _, service := range k.services() {
		vars[strings.ToUpper(service)+"_API_KEY"] = "sentinel-" + service
	}
	for _, name := range slices.Sorted(maps.Keys(vars)) {
		env.Setenv(name, vars[name])
	}
	extensions := filepath.Dir(k.installed(env.Getenv))
	if err := os.MkdirAll(filepath.Join(work, "bin"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(extensions, 0o755); err != nil {
		return err
	}
	if err := extension.Export(k.ext, extensions); err != nil {
		return err
	}
	return link(k.built, k.installed(env.Getenv))
}

// installed returns the path of ncly-<slug> in the extensions folder of
// the scenario's data folder, where discovery finds it first.
func (k *kit) installed(getenv func(string) string) string {
	return filepath.Join(getenv("XDG_DATA_HOME"), "nicely", "extensions", "ncly-"+k.manifest.Slug())
}

// services returns each service whose key a command declares, sorted.
func (k *kit) services() []string {
	var services []string
	for _, c := range k.manifest.Commands {
		services = append(services, c.Requires.Keys...)
	}
	slices.Sort(services)
	return slices.Compact(services)
}

// isolate returns path without each folder that holds an ncly-*
// file, as verify-nicely does. Discovery then finds no extension of the
// caller, only the one that the scenario installs.
func isolate(path string) string {
	var kept []string
	for _, dir := range filepath.SplitList(path) {
		if found, _ := filepath.Glob(filepath.Join(dir, "ncly-*")); len(found) == 0 {
			kept = append(kept, dir)
		}
	}
	return strings.Join(kept, string(filepath.ListSeparator))
}

// goCommand runs the go command in dir and returns its stdout without the
// final newline, or an error that holds its stderr.
func goCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(string(out)), nil
}

// link makes to the same executable as from: a hard link when the file
// system allows one, a copy otherwise.
func link(from, to string) error {
	if os.Link(from, to) == nil {
		return nil
	}
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	target, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(target, source)
	if closeErr := target.Close(); err == nil {
		err = closeErr
	}
	return err
}
