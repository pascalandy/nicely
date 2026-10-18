package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"slices"
	"strings"
)

type change struct {
	path     string
	old, new []byte
}

type limitRule func(change) bool

var skipped = []limitRule{
	// The go command writes go.sum and go.mod.
	func(c change) bool { return path.Base(c.path) == "go.sum" || path.Base(c.path) == "go.mod" },
	// Archives keep the text of their time.
	func(c change) bool { return strings.HasPrefix(c.path, "docs/archived/") },
	// Text of an extension, in the contract, asks for one catalog per language.
	func(c change) bool {
		return path.Base(path.Dir(c.path)) == "locales" && path.Ext(c.path) == ".toml"
	},
	// Generators own generated Go files.
	func(c change) bool {
		if path.Ext(c.path) != ".go" {
			return false
		}
		content := c.new
		if content == nil {
			content = c.old
		}
		return generated(content)
	},
	// Binary files hold no agent prose.
	func(c change) bool {
		content := c.new
		if content == nil {
			content = c.old
		}
		return bytes.IndexByte(content[:min(8000, len(content))], 0) >= 0
	},
}

// generated asks go/ast, so the gate exempts exactly what the go command
// treats as generated. A file without a valid package clause is not generated.
// The header comes before the package clause, so the parse stops there. A deep
// or broken body cannot change the answer.
func generated(src []byte) bool {
	file, _ := parser.ParseFile(token.NewFileSet(), "", src, parser.PackageClauseOnly|parser.ParseComments)
	return ast.IsGenerated(file)
}

const agentsSizeFix = `AGENTS.md only routes an agent to the files it must read. Move the text you added to the file that owns it, as the row "Where does a new rule go?" in "Find what you need" says, and leave a link in its place.`

const sizeFix = `Split each file in its own commit, with no change of behavior, then commit your change:
- Move each responsibility to a file of its own beside it. A test file follows the file it tests
- Markdown: move whole sections to files in a folder named after the file, such as contract/ for contract.md. Keep the opening in the file, and link each moved section. ` + "`just test`" + ` lists the links to fix
Do not delete comments or text to make room.`

func runLimits(args []string) error {
	if len(args) != 1 && (len(args) != 3 || args[1] != "--base") {
		return errors.New("usage: plan limits <repository root> [--base <ref>]")
	}
	refs, old, next := []string{"--cached"}, "HEAD", ""
	if len(args) == 3 {
		base, err := limitsGit(args[0], "merge-base", args[2], "HEAD")
		if err != nil {
			return err
		}
		old, next = strings.TrimSpace(string(base)), "HEAD"
		refs = []string{old, next}
	}
	changes, err := readChanges(args[0], refs, old, next)
	if err != nil {
		return err
	}
	var problems []string
	var size, agents, sentences bool
	previous := map[string]bool{}
	for _, c := range changes {
		if slices.ContainsFunc(skipped, func(r limitRule) bool { return r(c) }) {
			continue
		}
		for _, s := range extractSentences(c.path, c.old) {
			previous[s.text] = true
		}
	}
	for _, c := range changes {
		matches := func(r limitRule) bool { return r(c) }
		if slices.ContainsFunc(skipped, matches) {
			continue
		}
		cap := 2500
		if c.path == "AGENTS.md" {
			cap = 1250
		}
		if len(c.new) > cap*4 && (c.old == nil || len(c.new) > len(c.old)) {
			problem := fmt.Sprintf("%s holds %d tokens (bytes/4), more than %d", c.path, (len(c.new)+3)/4, cap)
			if c.old != nil {
				problem += fmt.Sprintf(", and the change grows it by %d bytes", len(c.new)-len(c.old))
			}
			problems = append(problems, problem)
			agents = agents || c.path == "AGENTS.md"
			size = size || c.path != "AGENTS.md"
		}
		for _, s := range extractSentences(c.path, c.new) {
			if s.words > 25 && !previous[s.text] {
				sentences = true
				problems = append(problems, fmt.Sprintf("%s:%d holds a sentence of %d words, more than 25: %q", c.path, s.line, s.words, s.preview+" ..."))
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	for _, fix := range []struct {
		needed bool
		text   string
	}{{agents, agentsSizeFix}, {size, sizeFix}, {sentences, sentenceFix}} {
		if fix.needed {
			problems = append(problems, "\n"+fix.text)
		}
	}
	return errors.New(strings.Join(problems, "\n"))
}
