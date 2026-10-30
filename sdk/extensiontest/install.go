package extensiontest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extensionrepo"
	"github.com/rogpeppe/go-internal/testscript"
)

func (k *kit) install(dir string) error {
	return extensionrepo.Install(k.ext, k.built, dir)
}

// checkInstallOrder blocks manifest publication on a first install.
// A failed sidecar publication must leave no executable for discovery.
func (k *kit) checkInstallOrder(dir string) error {
	name := "ncly-" + k.manifest.Slug()
	if err := os.Mkdir(filepath.Join(dir, name+".toml"), 0o755); err != nil {
		return err
	}
	err := k.install(dir)
	if err == nil {
		return errors.New("install: accepted a directory in place of the manifest. Fix: fail when sidecar publication fails")
	}
	var rename *os.LinkError
	if !errors.As(err, &rename) || rename.Op != "rename" || rename.New != filepath.Join(dir, name+".toml") {
		return fmt.Errorf("install: want failed manifest publication. Fix: fail when sidecar publication fails: %w", err)
	}
	if _, err := os.Lstat(filepath.Join(dir, name)); !errors.Is(err, fs.ErrNotExist) {
		return errors.New("install: published the executable before its sidecars. Fix: publish the executable last")
	}
	return nil
}

// uninstall removes only the extension under test from the scratch install.
func (k *kit) uninstall(dir string) error {
	if _, err := extensionrepo.Uninstall(k.ext, dir); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}
	return nil
}

// checkInstallLifecycle checks discovery after uninstall in its own scratch folders.
// Extensions that a scenario adds cannot keep the root present here.
func (k *kit) checkInstallLifecycle(work string) error {
	tmp := filepath.Join(work, ".tmp")
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}
	// Match testscript's runtime defaults before adding the scratch settings.
	env := &testscript.Env{WorkDir: work, Vars: []string{
		"WORK=" + work,
		"TMPDIR=" + tmp,
		"GOTRACEBACK=system",
	}}
	for _, name := range []string{"GOCOVERDIR", "GORACE"} {
		if value := os.Getenv(name); value != "" {
			env.Setenv(name, value)
		}
	}
	if err := k.installScratch(env); err != nil {
		return err
	}
	if err := k.uninstall(filepath.Dir(k.installed(env.Getenv))); err != nil {
		return err
	}
	args := append([]string{"describe"}, k.manifest.Root()...)
	args = append(args, "--json", "--no-input")
	cmd := exec.Command(filepath.Join(k.bin, "ncly"), args...)
	cmd.Dir, cmd.Env = env.WorkDir, env.Vars
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	status, err := exitCode(err)
	if err != nil {
		return fmt.Errorf("uninstall: ncly describe failed: %w", err)
	}
	var answer contract.Outcome
	if status != 2 || len(stdout) != 0 ||
		json.Unmarshal(stderr.Bytes(), &answer) != nil || len(answer.Errors) != 1 || answer.Errors[0].Code != contract.NotFound {
		return fmt.Errorf("uninstall: ncly describe must exit 2 with NOT_FOUND; got exit %d, stdout %q, stderr %q. Fix: remove the scratch install", status, stdout, stderr.String())
	}
	return nil
}
