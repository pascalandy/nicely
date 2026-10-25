package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// knownCodes returns the codes that a doc may name. They come from the tables
// under codeSections in the parts of contract.md and in each extension's spec,
// whole or split into spec/.
// They also come from the errors that the fixture manifests declare, and from
// exampleCode.
func knownCodes(root string) (map[string]bool, error) {
	codes := map[string]bool{exampleCode: true}
	parts, err := filepath.Glob(filepath.Join(root, "docs", "north-star", "contract", "*.md"))
	if err != nil {
		return nil, err
	}
	specs, err := filepath.Glob(filepath.Join(root, "extensions", "*", "spec.md"))
	if err != nil {
		return nil, err
	}
	specParts, err := filepath.Glob(filepath.Join(root, "extensions", "*", "spec", "*.md"))
	if err != nil {
		return nil, err
	}
	for _, name := range slices.Concat(parts, specs, specParts) {
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		section := ""
		var header []string
		var f fence
		for line := range strings.Lines(string(data)) {
			// A table in an example, such as a fenced Markdown block, defines no code.
			if f.in(line) {
				header = nil
				continue
			}
			if m := sectionTitle.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
				section = m[1]
			}
			cells, isHeader := tableRow(line, &header)
			if cells == nil || isHeader || !slices.Contains(codeSections, section) {
				continue
			}
			for i, h := range header {
				if (h == "Code" || h == "Warning") && i < len(cells) {
					for _, c := range cellCodes(cells[i]) {
						codes[c] = true
					}
				}
			}
		}
	}
	manifests, err := filepath.Glob(filepath.Join(root, "testdata", "extensions", "*", "manifest.toml"))
	if err != nil {
		return nil, err
	}
	for _, name := range manifests {
		var manifest struct {
			Errors []struct {
				Code string `toml:"code"`
			} `toml:"errors"`
		}
		if _, err := toml.DecodeFile(name, &manifest); err != nil {
			rel, _ := filepath.Rel(root, name)
			return nil, fmt.Errorf("%s: %w", filepath.ToSlash(rel), err)
		}
		for _, e := range manifest.Errors {
			codes[e.Code] = true
		}
	}
	return codes, nil
}
