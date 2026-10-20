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
// of its time. skippedNames are folders of dependencies and builds, wherever
// they sit. checkDocs also skips each hidden folder but checkedHidden, which
// holds the skills of this repository's agents. See #237.
var (
	skippedDirs   = []string{"docs/archived"}
	skippedNames  = []string{"node_modules", "dist"}
	checkedHidden = []string{".agents"}
)

// scriptReadme is the one Markdown file of a testdata folder that checkDocs
// reads: the rules of the scenarios beside it. The other files are fixtures,
// which can keep broken links on purpose.
const scriptReadme = "testdata/script/README.md"

// docLink is a link target, or a card ID, with the line that names it.
type docLink struct {
	target string
	line   int
}

// checkDocs reads each Markdown file outside the folders that it skips and
// outside the fixtures. It reports each broken link, each card that is not in
// known, and what checkProse finds with codes.
func checkDocs(root string, known, codes map[string]bool) ([]string, error) {
	var problems []string
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
				return filepath.SkipDir
			}
			return nil
		}
		// check reads the milestones, and their links, itself.
		if !strings.HasSuffix(rel, ".md") || path.Dir(rel) == milestoneDir && path.Base(rel) != "README.md" || fixture(rel) {
			return nil
		}
		data, err := os.ReadFile(name)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		for _, l := range links(string(data)) {
			if why := brokenLink(root, rel, l.target); why != "" {
				problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, l.line, why))
			}
		}
		for _, r := range cardRefs(string(data)) {
			if !known[r.target] {
				problems = append(problems, fmt.Sprintf("%s:%d: no milestone defines the card %s", rel, r.line, r.target))
			}
		}
		problems = append(problems, checkProse(rel, string(data), codes)...)
		return nil
	})
	return problems, err
}

// fixture reports whether rel, a file, lies in a testdata folder and is not
// the scriptReadme of one.
func fixture(rel string) bool {
	if rel == scriptReadme || strings.HasSuffix(rel, "/"+scriptReadme) {
		return false
	}
	return slices.Contains(strings.Split(path.Dir(rel), "/"), "testdata")
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
