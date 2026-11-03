package extensionrepo

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// OwnedRun is the environment and the scratch folder of one verification
// run, as Owned runs in the contract defines. The test kit, signoff,
// release, and acceptance each build theirs with NewOwnedRun. The analyzer
// callerenv permits whole caller reads only in the exact functions of
// Analyzers in the contract.
type OwnedRun struct {
	// Dir is the folder of the run. A scratch run holds HOME, TMPDIR, and
	// the XDG folders there. A host run keeps its notices there.
	Dir string
	// Env is the whole environment of each program that the run starts.
	Env []string
}

// Profile says what an owned run keeps of its caller.
type Profile struct {
	// Scratch gives the run its own HOME, TMPDIR, and XDG folders in Dir,
	// and no D-Bus session. It keeps no other variable of the caller.
	// Without Scratch, the run keeps the tools and accounts of the caller,
	// and drops the caller state of dropped.
	Scratch bool
	// Path holds the folders that go first in PATH, before those of the
	// caller that hold no ncly-* file. Only a scratch run reads it.
	Path []string
	// Keep names the variables of the caller that the run keeps anyway,
	// such as GIT_CONFIG_GLOBAL.
	Keep []string
	// Set gives variables their values in the run, after all others.
	Set map[string]string
}

// noticeFile names the file where Notice records the notices of a run.
const noticeFile = "NCLY_TEST_NOTICES"

// dropped reports whether a host run drops the variable name of its caller.
// These settings changed past checks (78963f6c, cd22fd11, 89f6b009, 09809608).
// They are the Nicely settings of a command, the metadata of a recipe or
// a Git hook, and the Go settings that select other source.
func dropped(name string) bool {
	switch name {
	case "NCLY_CONFIG", "NCLY_LANG", "NCLY_JSON", "NCLY_NO_INPUT", "NCLY_VERBOSE", ScenarioEnv, noticeFile, "GOFLAGS", "GOWORK":
		return true
	}
	return strings.HasPrefix(name, "JUST_") || strings.HasPrefix(name, "GIT_")
}

// NewOwnedRun builds the environment of a run in dir, which must be an
// absolute path, or empty for a host run without notices. It is the one
// function of the kit and the repository gates that reads the environment
// of the caller as a whole.
//
// Each run masks GOFLAGS with a blank value, which also masks one that go
// env -w keeps, and sets GOWORK=off, unless Keep names them. A host run with
// a folder records the notices of its programs in that folder.
func NewOwnedRun(dir string, p Profile) (OwnedRun, error) {
	if dir != "" && !filepath.IsAbs(dir) {
		return OwnedRun{}, errors.New("the folder of an owned run must be an absolute path")
	}
	if p.Scratch && dir == "" {
		return OwnedRun{}, errors.New("a scratch run needs a folder")
	}
	caller := map[string]string{}
	var names []string
	for _, entry := range os.Environ() {
		name, value, _ := strings.Cut(entry, "=")
		if _, twice := caller[name]; !twice {
			names = append(names, name)
		}
		caller[name] = value
	}
	vars := map[string]string{}
	order := []string{}
	set := func(name, value string) {
		if _, ok := vars[name]; !ok {
			order = append(order, name)
		}
		vars[name] = value
	}
	if !p.Scratch {
		for _, name := range names {
			if !dropped(name) {
				set(name, caller[name])
			}
		}
	}
	for _, name := range p.Keep {
		if value, ok := caller[name]; ok {
			set(name, value)
		}
	}
	if p.Scratch {
		tmp := filepath.Join(dir, ".tmp")
		if err := os.MkdirAll(tmp, 0o700); err != nil {
			return OwnedRun{}, err
		}
		set("PATH", strings.Join(append(slices.Clone(p.Path), withoutExtensions(caller["PATH"])...), string(filepath.ListSeparator)))
		set("HOME", dir)
		set("TMPDIR", tmp)
		set("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
		set("XDG_DATA_HOME", filepath.Join(dir, ".local", "share"))
		set("XDG_STATE_HOME", filepath.Join(dir, ".local", "state"))
		set("XDG_CACHE_HOME", filepath.Join(dir, ".cache"))
		// The keychain of Linux answers through the D-Bus session bus, which
		// this address cannot reach.
		set("DBUS_SESSION_BUS_ADDRESS", "disabled:")
	} else if dir != "" {
		set(noticeFile, filepath.Join(dir, "notices"))
	}
	if !slices.Contains(p.Keep, "GOFLAGS") {
		set("GOFLAGS", " ")
	}
	if !slices.Contains(p.Keep, "GOWORK") {
		set("GOWORK", "off")
	}
	for _, name := range slices.Sorted(maps.Keys(p.Set)) {
		set(name, p.Set[name])
	}
	env := make([]string, 0, len(order))
	for _, name := range order {
		env = append(env, name+"="+vars[name])
	}
	return OwnedRun{Dir: dir, Env: env}, nil
}

// Notices returns the notices that the programs of a host run recorded.
func (r OwnedRun) Notices() ([]byte, error) {
	if r.Dir == "" {
		return nil, nil
	}
	notices, err := os.ReadFile(filepath.Join(r.Dir, "notices"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return notices, err
}

// withoutExtensions returns the absolute folders of path that hold no ncly-*
// file, as verify-nicely does. Discovery then finds no extension of the
// caller.
func withoutExtensions(path string) []string {
	var kept []string
	for _, folder := range filepath.SplitList(path) {
		if !filepath.IsAbs(folder) {
			continue
		}
		if found, _ := filepath.Glob(filepath.Join(folder, "ncly-*")); len(found) == 0 {
			kept = append(kept, folder)
		}
	}
	return slices.Compact(kept)
}
