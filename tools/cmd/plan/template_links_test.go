package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Source links cannot silently opt a pinned extension into main's contract.
func TestPinnedDocumentationSources(t *testing.T) {
	for _, rel := range []string{"template", "extensions/skill/nicely-extensions"} {
		err := filepath.WalkDir(filepath.Join("../../..", rel), func(file string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			for n, line := range strings.Split(string(data), "\n") {
				if strings.Contains(line, "https://github.com/pascalandy/nicely/blob/main/") {
					t.Errorf("%s:%d: link Nicely documentation to the selected revision instead of main", file, n+1)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
