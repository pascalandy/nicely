package transcript

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// engine holds the Python engine: the modules of python/, without its tests
// and its notes.
//
//go:embed python/*.py
var engine embed.FS

// script is the file of the engine that uv runs, whose PEP 723 header holds
// the runtime pins.
const script = "transcript.py"

// Test changes how the engine runs, and only the test binary sets it, as
// Test transport in spec/protocol.md defines. A release leaves it empty.
var Test struct {
	// UVCache is the uv cache of the test binary, outside $WORK, which
	// replaces the one in Nicely's cache.
	UVCache string
	// Offline adds --offline, so uv runs from that cache without the
	// network.
	Offline bool
}

var engineHash = sync.OnceValue(func() string {
	h := sha256.New()
	_ = fs.WalkDir(engine, "python", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := engine.ReadFile(name)
		if err != nil {
			return err
		}
		_, _ = h.Write([]byte(path.Base(name) + "\x00"))
		_, _ = h.Write(data)
		_, _ = h.Write([]byte{0})
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))
})

// extract writes the engine into python/<hash>/ in the cache, where <hash>
// is the SHA-256 of the embedded files, and returns that folder. It writes
// a temporary folder beside it, then renames that folder, so a folder named
// after a hash always holds every file. A folder that exists already, from
// an earlier call, stays as it is.
func extract(cache string) (string, error) {
	dir := filepath.Join(cache, "python", engineHash())
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, nil
	}
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(parent, ".extract-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	entries, err := engine.ReadDir("python")
	for _, e := range entries {
		var data []byte
		if data, err = engine.ReadFile(path.Join("python", e.Name())); err == nil {
			err = os.WriteFile(filepath.Join(tmp, e.Name()), data, 0o644)
		}
		if err != nil {
			break
		}
	}
	if err == nil {
		err = os.Rename(tmp, dir)
	}
	if err != nil {
		// Another call may have renamed its own copy first.
		if info, statErr := os.Stat(dir); statErr == nil && info.IsDir() {
			return dir, nil
		}
		return "", err
	}
	return dir, nil
}

// uvArgs returns the words after uv that run the engine of dir, as Output
// and cookies in spec/output-and-cookies.md defines them. uv keeps its cache
// in Nicely's cache, ignores an inherited UV_CACHE_DIR, upgrades yt-dlp to
// its newest release, and never downloads Python.
func uvArgs(dir, cache string) []string {
	uvCache := filepath.Join(cache, "uv")
	if Test.UVCache != "" {
		uvCache = Test.UVCache
	}
	args := []string{"run", "--script", "--upgrade-package", "yt-dlp", "--no-python-downloads", "--cache-dir", uvCache}
	if Test.Offline {
		args = append(args, "--offline")
	}
	return append(args, filepath.Join(dir, script))
}
