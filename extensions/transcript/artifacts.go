package transcript

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/pascalandy/nicely/sdk/record"
)

// artifactNames are the required files of a result folder, in the order
// that the record lists them, as Output and cookies in
// spec/output-and-cookies.md names them. meta.txt is advisory, so it is no
// artifact.
var artifactNames = []string{"raw_transcript.txt", "raw_sentences.txt", "raw_transcript.json"}

// errUnusable is a required file that is missing, no regular file, or
// whose content the transcript cannot use.
var errUnusable = errors.New("the file is no usable transcript file")

// collect applies Verified artifacts in spec/records-and-replay.md to the
// result folder dir once its step completed. Each required name is a
// regular file directly inside dir, not a symlink, whose content is usable.
// It syncs each file, then dir, and returns them as artifacts, with their
// size and SHA-256.
func collect(dir string) ([]record.Artifact, error) {
	artifacts := make([]record.Artifact, 0, len(artifactNames))
	for _, name := range artifactNames {
		path := filepath.Join(dir, name)
		data, err := readSynced(path)
		if err != nil {
			return nil, err
		}
		if !usable(name, data) {
			return nil, errUnusable
		}
		sum := sha256.Sum256(data)
		artifacts = append(artifacts, record.Artifact{Path: path, SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(sum[:])})
	}
	return artifacts, syncPath(dir)
}

// readSynced reads a regular file that is no symlink, then syncs it, so its
// content survives a crash once the record lists it.
func readSynced(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errUnusable
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return data, err
}

// syncPath syncs a file or a folder. On macOS, File.Sync asks the drive to
// flush its cache too, with F_FULLFSYNC.
func syncPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = f.Sync()
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

// usable reports whether the transcript can use the content of a required
// file. Each file is valid UTF-8, the transcript is not blank, and the JSON
// file holds one array with nothing after it. An empty sentence list stays
// valid, as in the source program.
func usable(name string, data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	switch name {
	case "raw_transcript.txt":
		return strings.TrimSpace(string(data)) != ""
	case "raw_transcript.json":
		// null decodes into a slice too, but it is no array.
		var paragraphs []json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(data))
		if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) || dec.Decode(&paragraphs) != nil {
			return false
		}
		_, err := dec.Token()
		return errors.Is(err, io.EOF)
	}
	return true
}

// intact reports whether a recorded artifact of a completed run still
// holds what its record says, as replay asks. It is the regular file of its
// name directly inside dir, with its size, SHA-256, and usable content.
func intact(a record.Artifact, dir, name string) bool {
	if a.Path != filepath.Join(dir, name) {
		return false
	}
	info, err := os.Lstat(a.Path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := a.Read()
	return err == nil && usable(name, data)
}
