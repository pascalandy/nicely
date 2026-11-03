package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The rule indexes of D068. Each row of a table with an Enforced by column is
// a rule. Its enforcer links the files of its checks, or starts with
// proseOnly or onPurpose. proseOnly marks a rule that waits for a check.
// onPurpose comes with the judgment that keeps the rule in prose. The core
// index is each Markdown file of coreRules but its README, which lists the
// support code. The template index is templateIndex and its parts.
const (
	coreRules     = "docs/north-star/rules"
	templateIndex = "template/AGENTS/rule-index.md.jinja"
	templateParts = "template/AGENTS/rule-index"
	proseOnly     = "Prose only"
	onPurpose     = "Prose on purpose"
	enforcerTitle = "Enforced by"
	supportTitle  = "Support code"
	repoBlob      = "https://github.com/pascalandy/nicely/blob/main/"
)

// gateDirs hold the kit and the repository gates. Each Go file of them,
// outside its tests, is the enforcer of a rule in an index, or support code.
// See COR-104-T1, which fails first at 7f3ae60e.
var gateDirs = []string{"sdk/extensionrepo", "sdk/extensiontest"}

var (
	refDefinition = regexp.MustCompile(`^\[([^\]]+)\]:\s+(\S+)`)
	refLink       = regexp.MustCompile(`\]\[([^\]]+)\]`)
)

// ruleIndex counts the rules of one index by their enforcer.
type ruleIndex struct {
	name                       string
	rules, proseOnly, purposed int
}

// checkRuleIndexes returns the problems of the rule indexes of root.
func checkRuleIndexes(root string) ([]string, error) {
	_, problems, err := ruleIndexes(root)
	return problems, err
}

// ruleIndexes reads the core and the template index of root, when they
// exist. It checks each enforcer, and that a row or the support code names
// each gate file.
func ruleIndexes(root string) ([]ruleIndex, []string, error) {
	var indexes []ruleIndex
	var problems []string
	named := map[string]bool{}
	files := map[string][]string{}
	core, err := markdownFiles(root, coreRules, ".md")
	if err != nil {
		return nil, nil, err
	}
	files["core"] = slices.DeleteFunc(core, func(f string) bool { return path.Base(f) == "README.md" })
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(templateIndex))); err == nil {
		parts, err := markdownFiles(root, templateParts, ".md.jinja")
		if err != nil {
			return nil, nil, err
		}
		files["template"] = append([]string{templateIndex}, parts...)
	}
	ids := map[string]string{}
	for _, name := range []string{"core", "template"} {
		if len(files[name]) == 0 {
			continue
		}
		index := ruleIndex{name: name}
		for _, file := range files[name] {
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
			if err != nil {
				return nil, nil, err
			}
			problems = append(problems, readRules(root, file, string(data), &index, named, ids)...)
		}
		if index.rules == 0 {
			problems = append(problems, fmt.Sprintf("%s: the %s index holds no row under an %s column", files[name][0], name, enforcerTitle))
		}
		indexes = append(indexes, index)
	}
	support := supportCode(root)
	for _, dir := range gateDirs {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			file := path.Join(dir, e.Name())
			if e.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") || named[file] || support[file] {
				continue
			}
			problems = append(problems, fmt.Sprintf("%s: no row of a rule index links this file. Link it from the rules that it enforces, or list it under %s in %s/README.md", file, supportTitle, coreRules))
		}
	}
	return indexes, problems, nil
}

// markdownFiles returns the files of dir whose name ends with suffix, as paths
// from root, or none when dir does not exist.
func markdownFiles(root, dir, suffix string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
	if os.IsNotExist(err) {
		return nil, nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			files = append(files, path.Join(dir, e.Name()))
		}
	}
	return files, err
}

// readRules counts the rules of one file of an index, adds the files that
// their enforcers link to named, and returns its problems. ids maps each ID
// of a core rule to the file that holds it.
func readRules(root, file, text string, index *ruleIndex, named map[string]bool, ids map[string]string) []string {
	var problems []string
	refs := definitions(text)
	var header []string
	var f fence
	for i, line := range strings.Split(text, "\n") {
		n := i + 1
		if f.in(line) {
			header = nil
			continue
		}
		cells, isHeader := tableRow(line, &header)
		column := slices.Index(header, enforcerTitle)
		if cells == nil || isHeader || column < 0 {
			continue
		}
		at := fmt.Sprintf("%s:%d", file, n)
		if len(cells) != len(header) {
			problems = append(problems, fmt.Sprintf("%s: the row has %d cells, and its header %d", at, len(cells), len(header)))
			continue
		}
		index.rules++
		if id := slices.Index(header, "ID"); id >= 0 {
			if other, seen := ids[cells[id]]; seen {
				problems = append(problems, fmt.Sprintf("%s: %s also names the rule %s", at, other, cells[id]))
			}
			ids[cells[id]] = at
		}
		// Each link of the row must reach a file, and only those of the
		// enforcer name a check.
		var targets []string
		for i, cell := range cells {
			found, broken := cellLinks(root, file, cell, refs)
			for _, b := range broken {
				problems = append(problems, at+": "+b)
			}
			if i == column {
				targets = found
			}
		}
		enforcer := cells[column]
		switch {
		case strings.HasPrefix(enforcer, proseOnly):
			index.proseOnly++
		case strings.HasPrefix(enforcer, onPurpose):
			index.purposed++
			if judgment, ok := strings.CutPrefix(enforcer, onPurpose+": "); !ok || strings.TrimSpace(judgment) == "" {
				problems = append(problems, fmt.Sprintf("%s: write the judgment that keeps the rule in prose: %s: <judgment>", at, onPurpose))
			}
		case len(targets) == 0:
			problems = append(problems, fmt.Sprintf("%s: the enforcer %q links no file. Link its check, or write %s or %s: <judgment>", at, enforcer, proseOnly, onPurpose))
		}
		for _, t := range targets {
			named[t] = true
		}
	}
	return problems
}

// supportCode returns the files that the Support code section of the core
// README links. checkDocs checks its inline links.
func supportCode(root string) map[string]bool {
	file := path.Join(coreRules, "README.md")
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		return nil
	}
	text := string(data)
	refs := definitions(text)
	support := map[string]bool{}
	var section string
	for line := range strings.Lines(text) {
		if title, ok := strings.CutPrefix(strings.TrimRight(line, "\n"), "## "); ok {
			section = title
			continue
		}
		if section == supportTitle {
			targets, _ := cellLinks(root, file, line, refs)
			for _, t := range targets {
				support[t] = true
			}
		}
	}
	return support
}

// definitions maps each reference of text, outside code, to its target.
func definitions(text string) map[string]string {
	refs := map[string]string{}
	var f fence
	for line := range strings.Lines(text) {
		if f.in(line) {
			continue
		}
		if m := refDefinition.FindStringSubmatch(line); m != nil {
			refs[strings.ToLower(m[1])] = m[2]
		}
	}
	return refs
}

// cellLinks returns the repository files that the links of cell reach, inline
// or by reference, and why each other link reaches none. file is the path of
// the index from root.
func cellLinks(root, file, cell string, refs map[string]string) (targets, broken []string) {
	cell = codeSpan.ReplaceAllString(cell, "")
	var raw []string
	for _, m := range mdLink.FindAllStringSubmatch(cell, -1) {
		raw = append(raw, m[1])
	}
	for _, m := range refLink.FindAllStringSubmatch(cell, -1) {
		target, ok := refs[strings.ToLower(m[1])]
		if !ok {
			broken = append(broken, fmt.Sprintf("the reference [%s] has no definition", m[1]))
			continue
		}
		raw = append(raw, target)
	}
	for _, target := range raw {
		name, _, _ := strings.Cut(target, "#")
		var rel string
		switch {
		case strings.HasPrefix(name, repoBlob):
			rel = strings.TrimPrefix(name, repoBlob)
		case strings.Contains(name, "://"):
			broken = append(broken, fmt.Sprintf("the link %s leaves this repository. Link a file as %s<path>", target, repoBlob))
			continue
		default:
			rel = path.Join(path.Dir(file), name)
		}
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil || info.IsDir() {
			broken = append(broken, fmt.Sprintf("the link %s reaches no file", target))
			continue
		}
		targets = append(targets, rel)
	}
	return targets, broken
}

// ruleStatus writes the rules of each index by enforcer, for just status.
func ruleStatus(out *strings.Builder, indexes []ruleIndex) {
	if len(indexes) == 0 {
		return
	}
	fmt.Fprintln(out, "\nRule indexes")
	rows := [][]string{{"Index", "Rules", proseOnly, onPurpose}}
	for _, index := range indexes {
		rows = append(rows, []string{index.name, fmt.Sprint(index.rules), fmt.Sprint(index.proseOnly), fmt.Sprint(index.purposed)})
	}
	out.WriteString(table("", rows))
}
