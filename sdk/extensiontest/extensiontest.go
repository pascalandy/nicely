// Package extensiontest is the test kit of an extension repository, as Test
// kit in contract.md defines. One test of the repository calls Run. Run
// checks the links between the manifest, the skill, the scenarios, the
// spec, and the Nicely versions. It installs the real ncly-<slug> and
// the ncly of go tool ncly in scratch folders.
// It runs manifest examples there, then the scenarios of testdata/script/.
// Coverage requires a checked JSON answer or a completed success with a declared text result.
//
// A scenario runs the extension through the commands that the kit adds, so
// that the kit checks each answer against the schema of its command:
//
//	ncly <arg>...                 runs ncly, which exits 0, or not 0 after !
//	ncly-<slug> <arg>...          runs the extension directly, the same way
//	exits <code> <command> <arg>  runs a command, these two included, and checks its exact exit code
//	signal INT|TERM <marker> <command> <arg>  checks the host's signal exit and answer
//	survives <file>...           requires finished files before and after the next signal
//	appears <marker> <path> <command> <arg>  requires a failure that keeps a new file
//	protect <path>...            selects paths that refuses and the next signal or appears keep
//	refuses <code> <command> <arg>  checks the same refusal in a dry run and a run
//	predicts <command> <arg>...  checks that a dry run predicts the publication of its run
//	grant <service>               grants the key of a declared service to the extension
//	fake <name>                   puts the fake child in $WORK/bin as name
//
// The rules of its code:
//   - No case reads the user's keychain, HOME, or XDG folders, or finds the
//     user's extensions. Each scenario gets an owned scratch run of
//     extensionrepo.NewOwnedRun. It has a scratch HOME, TMPDIR, and XDG
//     folders, and a PATH without the folders that hold an ncly-* file. Each service that the manifest declares gets the sentinel key
//     sentinel-<service> in its variable, so ncly never reads the keychain
//     for it. On Linux, D-Bus names no bus, so a read that a scenario forces
//     finds no keychain
//   - No code of the kit reads the environment of the caller as a whole.
//     The analyzer callerenv refuses os.Environ here
//   - The kit imports no package of core's internal/, which an extension
//     repository cannot import either
//   - The kit builds every program that it runs from Go source: the
//     extension, ncly, and the fake child of the fake/ folder
//   - The kit copies each program to where it runs, and never makes a hard
//     link. macOS checks a new program at its first start through one name
//     of its file. A scenario deletes its names when it ends. A check through
//     such a name fails, and macOS kills the start with exit 137. See #371
//   - Each failure names its file, the missing link, and its fix
package extensiontest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/extensionrepo"
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
	examples := *k
	examples.Coverage = k.fresh()
	examples.RunExamples(t, filepath.Join(k.root, "testdata", "examples.txtar"), testscript.Params{
		RequireUniqueNames: true,
		Setup:              examples.setup,
		Cmds:               examples.commands(),
	})
	// testscript starts parallel subtests. Cleanup runs after all scenarios finish.
	t.Cleanup(func() { k.reportCoverage(t) })
	if err := k.checkInstallOrder(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if err := k.checkInstallLifecycle(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir:                filepath.Join(k.root, "testdata", "script"),
		RequireUniqueNames: true,
		Setup:              k.setup,
		Cmds:               k.commands(),
	})
}

// kit is what Run knows of one extension repository.
type kit struct {
	*Coverage
	ext extension.Extension
	// root is the folder of the repository's go.mod.
	root string
	// bin holds ncly, which go tool ncly built.
	bin string
	// built is the ncly-<slug> that go build made, and fake the fake
	// child.
	built, fake string
	// publications and reasons are the tables of ## Publication of spec.md.
	publications []publication
	reasons      []reason
}

func load(e extension.Extension) (*kit, error) {
	coverage, err := NewCoverage(e)
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
	return &kit{Coverage: coverage, ext: e, root: filepath.Dir(gomod)}, nil
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
		err = copyProgram(host, filepath.Join(k.bin, "ncly"))
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

// setup installs the extension in a scenario's scratch folders and registers cleanup.
func (k *kit) setup(env *testscript.Env) error {
	if err := k.installScratch(env); err != nil {
		return err
	}
	extensions := filepath.Dir(k.installed(env.Getenv))
	env.Defer(func() {
		if err := k.uninstall(extensions); err != nil {
			env.T().Fatal(err)
		}
	})
	return nil
}

// installScratch gives an install its owned scratch run, sentinel keys, and
// extension. The run keeps only the coverage settings of the caller.
// It publishes the executable last, after the sidecars that Export writes.
func (k *kit) installScratch(env *testscript.Env) error {
	work := env.WorkDir
	run, err := extensionrepo.NewOwnedRun(work, extensionrepo.Profile{
		Scratch: true,
		Path:    []string{filepath.Join(work, "bin"), k.bin},
		Keep:    []string{"GOCOVERDIR", "GORACE"},
	})
	if err != nil {
		return err
	}
	for _, entry := range run.Env {
		name, value, _ := strings.Cut(entry, "=")
		env.Setenv(name, value)
	}
	for _, service := range k.services() {
		env.Setenv(strings.ToUpper(service)+"_API_KEY", "sentinel-"+service)
	}
	extensions := filepath.Dir(k.installed(env.Getenv))
	if err := os.MkdirAll(filepath.Join(work, "bin"), 0o755); err != nil {
		return err
	}
	return k.install(extensions)
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

// copyProgram copies the program from to the new file to.
func copyProgram(from, to string) error {
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
