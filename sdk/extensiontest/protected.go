package extensiontest

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type fileState struct {
	mode fs.FileMode
	sum  [sha256.Size]byte
	link string
}

// snapshotFiles records paths without following symbolic links. Missing roots remain protected by their absence.
func snapshotFiles(paths []string, regular bool) (map[string]fileState, error) {
	states := map[string]fileState{}
	for _, path := range paths {
		info, err := os.Lstat(path)
		if !regular && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if regular && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("survives: %s is not a finished regular file", path)
		}
		err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			state := fileState{mode: info.Mode()}
			switch {
			case info.Mode().IsRegular():
				state.sum, err = fileDigest(path)
			case info.Mode()&os.ModeSymlink != 0:
				state.link, err = os.Readlink(path)
			case !info.IsDir():
				err = fmt.Errorf("protect: unsupported file type at %s", path)
			}
			states[path] = state
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	return states, nil
}

func fileDigest(path string) ([sha256.Size]byte, error) {
	var sum [sha256.Size]byte
	file, err := os.Open(path)
	if err != nil {
		return sum, err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	copy(sum[:], hash.Sum(nil))
	return sum, err
}
