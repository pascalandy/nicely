// Package extensionrepo runs repository tasks through the extension's export program.
// It owns staged install and uninstall, as Extension repository in contract.md defines.
// Its test wrapper prints notices after Go's package results, as Test kit defines.
package extensionrepo

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// Install stages executable and the embedded sidecars of e, then publishes the executable last.
// The kit uses it with its built executable and a scratch extensions folder.
func Install(e extension.Extension, executable, dir string) error {
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return err
	}
	name := "ncly-" + m.Slug()
	return install(dir, name, func(stage string) error {
		if err := copyProgram(executable, filepath.Join(stage, name)); err != nil {
			return err
		}
		return extension.Export(e, stage)
	})
}

// install prepares all files beside dir before replacing any installed file.
// Each publication is one rename on the destination's file system.
func install(dir, name string, prepare func(string) error) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(dir, ".install.")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err := prepare(stage); err != nil {
		return err
	}
	for _, suffix := range []string{".locales", ".skill"} {
		if err := moveIfPresent(filepath.Join(dir, name+suffix), filepath.Join(stage, "previous-"+name+suffix)); err != nil {
			return err
		}
	}
	for _, suffix := range []string{".toml", ".locales", ".skill", ""} {
		if err := moveIfPresent(filepath.Join(stage, name+suffix), filepath.Join(dir, name+suffix)); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall removes e from dir, with the executable first.
// It returns the paths removed, in order, and accepts an absent install.
func Uninstall(e extension.Extension, dir string) ([]string, error) {
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return nil, err
	}
	name := "ncly-" + m.Slug()
	var stage string
	defer func() {
		if stage != "" {
			_ = os.RemoveAll(stage)
		}
	}()
	var removed []string
	for _, suffix := range []string{"", ".toml", ".locales", ".skill"} {
		path := filepath.Join(dir, name+suffix)
		if _, err := os.Lstat(path); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return removed, err
		}
		if stage == "" {
			stage, err = os.MkdirTemp(dir, ".uninstall.")
			if err != nil {
				return removed, err
			}
		}
		if err := os.Rename(path, filepath.Join(stage, name+suffix)); err != nil {
			return removed, err
		}
		removed = append(removed, path)
	}
	return removed, nil
}

func moveIfPresent(from, to string) error {
	if _, err := os.Lstat(from); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Rename(from, to)
}

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
