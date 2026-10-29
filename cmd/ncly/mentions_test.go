package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/rogpeppe/go-internal/testscript"
)

// printCodes prints every code of core's registry, one per line.
func printCodes() {
	for _, code := range contract.Core.Codes() {
		fmt.Println(code)
	}
	os.Exit(0)
}

// mentionedCommand finds a code span of a skill that starts with ncly and a
// lowercase word, such as `ncly skill view nicely`, and captures its words.
// Prose such as "every ncly call" is no command.
var mentionedCommand = regexp.MustCompile("`ncly((?: [a-z][a-z0-9-]*)+)[^`]*`")

// mentionedCode finds a word of the form of a code. A word before "=" or
// one that starts with NCLY_ is an environment variable instead.
var mentionedCode = regexp.MustCompile(`\b[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+\b=?`)

// cmdMentions fails when a skill names an ncly command in a code span that
// describe does not give, or a code that the registry lacks, or holds more
// than 60 lines: mentions <skill-file> <describe-file> <codes-file>. After
// a command path, a word is a command only when that path groups others.
// With -bytes <max> first, it also fails for a skill of more than max bytes.
func cmdMentions(ts *testscript.TestScript, neg bool, args []string) {
	usage := "usage: mentions [-bytes <max>] <skill-file> <describe-file> <codes-file>"
	maxBytes := -1
	if len(args) > 1 && args[0] == "-bytes" {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 0 {
			ts.Fatalf("%s", usage)
		}
		maxBytes, args = n, args[2:]
	}
	if len(args) != 3 {
		ts.Fatalf("%s", usage)
	}
	skill := ts.ReadFile(args[0])
	var listing struct {
		Commands []struct {
			Path string `json:"path"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(ts.ReadFile(args[1])), &listing); err != nil {
		ts.Fatalf("mentions: %s: %v", args[1], err)
	}
	paths := map[string]bool{}
	groups := map[string]bool{}
	for _, c := range listing.Commands {
		paths[c.Path] = true
		if i := strings.LastIndex(c.Path, " "); i > 0 {
			groups[c.Path[:i]] = true
		}
	}
	codes := strings.Fields(ts.ReadFile(args[2]))
	var problems []string
	if lines := strings.Count(skill, "\n"); lines > 60 {
		problems = append(problems, fmt.Sprintf("%d lines, more than 60", lines))
	}
	if maxBytes >= 0 && len(skill) > maxBytes {
		problems = append(problems, fmt.Sprintf("%d bytes, more than %d", len(skill), maxBytes))
	}
	for _, m := range mentionedCommand.FindAllStringSubmatch(skill, -1) {
		words := strings.Fields(m[1])
		path := words[0]
		for _, word := range words[1:] {
			if !groups[path] {
				break
			}
			path += " " + word
		}
		if !paths[path] {
			problems = append(problems, fmt.Sprintf("the command %q, which ncly lacks", "ncly "+path))
		}
	}
	for _, word := range mentionedCode.FindAllString(skill, -1) {
		if strings.HasSuffix(word, "=") || strings.HasPrefix(word, "NCLY_") {
			continue
		}
		if !slices.Contains(codes, word) {
			problems = append(problems, fmt.Sprintf("the code %s, which ncly lacks", word))
		}
	}
	switch {
	case neg && len(problems) == 0:
		ts.Fatalf("mentions: %s names only what ncly has", args[0])
	case !neg && len(problems) > 0:
		ts.Fatalf("mentions: %s names %s", args[0], strings.Join(problems, "; "))
	}
}
