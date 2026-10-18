package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"regexp"
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
	// Generators own generated files.
	func(c change) bool {
		content := c.new
		if content == nil {
			content = c.old
		}
		lines := bytes.SplitN(content, []byte("\n"), 11)
		return generatedMarker.Match(bytes.Join(lines[:min(10, len(lines))], []byte("\n")))
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

// generatedMarker is the header line that Go's convention reserves for
// generated files, so a doc that quotes it is still checked.
var generatedMarker = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

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

// limitsGit returns the standard output of git alone, because a warning on
// stderr, such as a skipped rename detection, would corrupt a parsed diff.
func limitsGit(root string, args ...string) ([]byte, error) {
	data, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, exit.Stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", args[0], err)
	}
	return data, nil
}

func readChanges(root string, refs []string, old, next string) ([]change, error) {
	data, err := limitsGit(root, append([]string{"diff", "--name-status", "-z", "-M", "-l0", "--ignore-submodules=all", "--diff-filter=ACMRD"}, refs...)...)
	if err != nil {
		return nil, err
	}
	fields := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
	var changes []change
	for i := 0; i+1 < len(fields); {
		status, before := fields[i], fields[i+1]
		i += 2
		after := before
		if status[0] == 'R' || status[0] == 'C' {
			after = fields[i]
			i++
		}
		c := change{path: after}
		if status != "A" {
			c.old, err = limitsGit(root, "show", old+":"+before)
			if err != nil {
				return nil, err
			}
			if c.old == nil {
				c.old = []byte{}
			}
		}
		if status != "D" {
			c.new, err = limitsGit(root, "show", next+":"+after)
			if err != nil {
				return nil, err
			}
		}
		changes = append(changes, c)
	}
	return changes, nil
}
