package extension

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/contract"
)

// Export writes the sidecars of e into dir, as SDK in contract.md defines.
// They are the manifest as ncly-<slug>.toml, the catalogs as
// ncly-<slug>.locales/, and the skill as ncly-<slug>.skill/. An install
// that calls it takes them from the build under test, never from the source
// tree. Each sidecar replaces the one of the same name. A file that e no
// longer holds goes, and so does the skill folder when e has no skill.
// Export writes no executable.
func Export(e Extension, dir string) error {
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return err
	}
	base := filepath.Join(dir, "ncly-"+m.Slug())
	locales, err := fs.Sub(e.Locales, "locales")
	if err != nil {
		return err
	}
	if err := replaceFolder(base+".locales", locales); err != nil {
		return err
	}
	var skill fs.FS
	if e.Skill != nil {
		if skill, err = fs.Sub(e.Skill, "skill"); err != nil {
			return err
		}
	}
	if err := replaceFolder(base+".skill", skill); err != nil {
		return err
	}
	return replaceFile(base+".toml", e.Manifest)
}

// replaceFolder writes files as the folder path, or removes path when files
// is nil. The new folder is written beside path first, so a failed copy
// keeps the previous one.
func replaceFolder(path string, files fs.FS) error {
	if files == nil {
		return os.RemoveAll(path)
	}
	stage, err := os.MkdirTemp(filepath.Dir(path), "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	staged := filepath.Join(stage, "new")
	if err := os.CopyFS(staged, files); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.Rename(staged, path)
}

// replaceFile writes data as the file path through a rename, so a reader
// sees the previous file or the new one, never a part.
func replaceFile(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	_, err = file.Write(data)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Chmod(file.Name(), 0o644)
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
