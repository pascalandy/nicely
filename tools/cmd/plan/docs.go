package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// skippedDirs hold no links for checkDocs: archived history keeps the links
// of its time. skippedNames are folders of fixtures, dependencies, and
// builds, wherever they sit. checkDocs also skips each hidden folder but
// checkedHidden, which holds the skills of this repository's agents. See #237.
var (
	skippedDirs   = []string{"docs/archived"}
	skippedNames  = []string{"testdata", "node_modules", "dist"}
	checkedHidden = []string{".agents"}
)

// scriptReadme is the one Markdown file of a testdata folder that checkDocs
// reads. It holds the rules of the scenarios, and checkDocs never walks that
// folder. The other files are fixtures, which can keep broken links on purpose.
const scriptReadme = "script/README.md"

// docLink is a link target, or a card ID, with the line that names it.
type docLink struct {
	target string
	line   int
}

// checkDocs reads each Markdown file outside the folders that it skips, and
// the scriptReadme of each testdata folder. It reports each broken link, each
// card that is not in known, and what checkProse finds with codes.
func checkDocs(root string, known, codes map[string]bool) ([]string, error) {
	var problems []string
	checkFile := func(rel string) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", rel, err))
			return
		}
		for _, l := range links(string(data)) {
			if why := brokenLink(root, rel, l.target); why != "" {
				problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, l.line, why))
			}
		}
		if path.Dir(rel) == milestoneDir {
			return
		}
		for _, r := range cardRefs(string(data)) {
			if !known[r.target] {
				problems = append(problems, fmt.Sprintf("%s:%d: no milestone defines the card %s", rel, r.line, r.target))
			}
		}
		problems = append(problems, checkProse(rel, string(data), codes)...)
	}
	err := filepath.WalkDir(root, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			hidden := strings.HasPrefix(d.Name(), ".") && !slices.Contains(checkedHidden, rel)
			if rel != "." && (hidden || slices.Contains(skippedDirs, rel) || slices.Contains(skippedNames, d.Name())) {
				if readme := path.Join(rel, scriptReadme); d.Name() == "testdata" {
					if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(readme))); err == nil {
						checkFile(readme)
					}
				}
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(rel, ".md") {
			checkFile(rel)
		}
		return nil
	})
	return problems, err
}

// links returns the Markdown links of text outside code: fenced blocks,
// indented or not, and code spans.
func links(text string) []docLink {
	var found []docLink
	var f fence
	for i, line := range strings.Split(text, "\n") {
		if f.in(line) {
			continue
		}
		for _, l := range mdLink.FindAllStringSubmatch(codeSpan.ReplaceAllString(line, ""), -1) {
			found = append(found, docLink{l[1], i + 1})
		}
	}
	return found
}

// cardRefs returns the card IDs that text names, in code too, since a
// renumbered card makes a command or an example as stale as a sentence.
func cardRefs(text string) []docLink {
	var found []docLink
	for i, line := range strings.Split(text, "\n") {
		for _, id := range cardRef.FindAllString(line, -1) {
			found = append(found, docLink{id, i + 1})
		}
	}
	return found
}

// brokenLink says why target, a link in file, reaches no file or heading, or
// returns "" when it does. file is a path from root.
func brokenLink(root, file, target string) string {
	if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
		return ""
	}
	name, anchor, _ := strings.Cut(target, "#")
	full := filepath.Join(root, filepath.FromSlash(file))
	if name != "" {
		full = filepath.Join(root, filepath.FromSlash(path.Join(path.Dir(file), name)))
	}
	info, err := os.Stat(full)
	if err != nil {
		return fmt.Sprintf("the link %s points to a missing file", target)
	}
	if anchor == "" || info.IsDir() || !strings.HasSuffix(full, ".md") {
		return ""
	}
	if !slices.Contains(anchors(full), anchor) {
		return fmt.Sprintf("the link %s points to a missing heading", target)
	}
	return ""
}

var anchorCache = map[string][]string{}

// anchors returns the anchors that GitHub gives the headings of a Markdown file.
func anchors(file string) []string {
	if a, ok := anchorCache[file]; ok {
		return a
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var list []string
	seen := map[string]int{}
	var f fence
	for line := range strings.Lines(string(data)) {
		// A heading starts its line, so only a fence of the root can hold one,
		// and a marker indented by four spaces is code there, never a fence.
		code := f.open != ""
		if !strings.HasPrefix(line, "    ") {
			code = f.in(line)
		}
		level := len(line) - len(strings.TrimLeft(line, "#"))
		if code || level == 0 || level > 6 || !strings.HasPrefix(line[level:], " ") {
			continue
		}
		s := slug(strings.TrimSpace(line[level:]))
		if n := seen[s]; n > 0 {
			list = append(list, s+"-"+strconv.Itoa(n))
		} else {
			list = append(list, s)
		}
		seen[s]++
	}
	anchorCache[file] = list
	return list
}

func slug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r), unicode.IsMark(r), r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}
