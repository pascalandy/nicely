package platform

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// WriteFile replaces the file at path with data durably, as Record format
// in contract.md writes a record: it writes a temporary file of its own in
// the same folder, flushes it, renames it over path, then flushes the
// folder. It creates the folder first, as MkdirAll does. When it fails
// before the rename, path keeps what it held, and the temporary file is
// gone.
func WriteFile(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	if err := MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	err = f.Chmod(perm)
	if err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = flush(f)
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return flushDir(dir)
}

// MkdirAll creates dir and each missing parent, and flushes the parent of
// each folder that it creates, so a new folder survives a crash, as Record
// format in contract.md asks of a new state folder.
func MkdirAll(dir string, perm fs.FileMode) error {
	info, err := os.Stat(dir)
	switch {
	case err == nil && info.IsDir():
		return nil
	case err == nil:
		return &fs.PathError{Op: "mkdir", Path: dir, Err: unix.ENOTDIR}
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}
	parent := filepath.Dir(dir)
	if parent != dir {
		if err := MkdirAll(parent, perm); err != nil {
			return err
		}
	}
	if err := os.Mkdir(dir, perm); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return flushDir(parent)
}

// flushDir flushes a folder, so that a rename or a new entry in it
// survives a crash.
func flushDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = flush(d)
	if closeErr := d.Close(); err == nil {
		err = closeErr
	}
	return err
}

// ErrLocked means that another process held a lock for longer than the
// wait.
var ErrLocked = errors.New("another process holds the lock")

// Lock takes the lock file at path, as Record format in contract.md
// defines: a file of its own, which Lock creates when it is missing and
// nobody deletes, because a deleted lock file lets two processes own the
// lock. It waits at most wait while another process holds the lock, then
// returns ErrLocked, and it returns the error of ctx, holding nothing, when
// ctx ends before it takes the lock, even a free one. The returned function
// releases the lock, and the operating system releases it when the process
// ends.
func Lock(ctx context.Context, path string, wait time.Duration) (func(), error) {
	if err := MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(wait)
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		switch {
		case err == nil && ctx.Err() != nil:
			_ = f.Close()
			return nil, ctx.Err()
		case err == nil:
			return func() { _ = f.Close() }, nil
		case errors.Is(err, unix.EINTR):
			continue
		case !errors.Is(err, unix.EWOULDBLOCK):
			_ = f.Close()
			return nil, &fs.PathError{Op: "flock", Path: path, Err: err}
		case !time.Now().Before(deadline):
			_ = f.Close()
			return nil, ErrLocked
		}
		select {
		case <-ctx.Done():
			_ = f.Close()
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Held reports whether another process holds the lock file at path. It
// tries the lock without waiting and without creating the file, and
// releases it at once, as inspection and dry run in contract.md ask, so it
// never makes an execution see the lock as held. A missing file is a free
// lock, because nobody deletes a lock file that someone holds.
func Held(path string) (bool, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		switch {
		case err == nil:
			return false, nil
		case errors.Is(err, unix.EINTR):
			continue
		case errors.Is(err, unix.EWOULDBLOCK):
			return true, nil
		}
		return false, &fs.PathError{Op: "flock", Path: path, Err: err}
	}
}
