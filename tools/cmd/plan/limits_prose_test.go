package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLimitsProse(t *testing.T) {
	words := func(n int) string { return strings.TrimSpace(strings.Repeat("word ", n)) + "." }
	long := words(26) + "\n"
	half := strings.Repeat("word ", 15)
	cases := []struct {
		name, file, old, new string
		want                 []string
	}{
		{"26 words", "notes.md", "", "# Title\n\n" + long, []string{"notes.md:3 holds a sentence of 26 words, more than 25", `"word word word word word word word word word word word word ..."`}},
		{"25 words", "notes.md", "", words(25), nil},
		{"untouched", "notes.md", long + "\nBefore.\n", long + "\nAfter.\n", nil},
		{"typo outside sentence", "notes.md", long + "Typo.\n", long + "Fixed.\n", nil},
		{"typo on same line", "notes.md", words(26) + " Typo.", words(26) + " Fixed.", nil},
		{"changed sentence", "notes.md", long, strings.Replace(long, "word", "changed", 1), []string{"notes.md:1 holds a sentence of 26 words, more than 25"}},
		{"wrapped paragraph", "notes.md", "", "# Title\n\n" + half + "\n" + words(15), []string{"notes.md:3 holds a sentence of 30 words, more than 25"}},
		{"wrapped sentences", "notes.md", "", words(25) + "\n" + words(25), nil},
		{"sentence on later line", "notes.md", "", "Short.\n" + long, []string{"notes.md:2 holds a sentence of 26 words, more than 25"}},
		{"list continuation", "notes.md", "", "Intro.\n- " + half + "\n  " + words(15), []string{"notes.md:2 holds a sentence of 30 words, more than 25"}},
		{"list paragraphs", "notes.md", "", "- " + strings.TrimSuffix(words(25), ".") + "\n* " + words(25) + "\n+ " + words(25) + "\n1. " + words(25), nil},
		{"end list", "notes.md", "", "- " + half + "\n" + words(15), nil},
		{"metadata", "notes.md", "", "Status: done\nVersion: v0.0.1\n", nil},
		{"move within file", "notes.md", long + "\nShort.\n", "Short.\n\n" + half + "\n" + words(11) + "\n", nil},
		{"link target", "notes.md", strings.TrimSuffix(words(26), ".") + " [link](old.md).\n", strings.TrimSuffix(words(26), ".") + " [link](new.md).\n", nil},
		{"code span", "notes.md", "", strings.TrimSuffix(words(24), ".") + " `a complete command here`.", nil},
		{"code span overflow", "notes.md", "", strings.TrimSuffix(words(25), ".") + " `a complete command here`.", []string{"notes.md:1 holds a sentence of 26 words, more than 25"}},
		{"fence", "notes.md", "", "````markdown\n```\n" + long + "```\n````\nDone.\n", nil},
		{"paragraph breaks", "notes.md", "", half + "\n\n" + half + "\n# Title\n" + half + "\n```\ncode\n```\n" + half + "\n<!-- comment -->\n" + half + "\n| Cell |\n|---|\n" + half, nil},
		{"table cells", "notes.md", "", "| First | Second |\n| :--- | ---: |\n| " + words(25) + " | " + words(25) + " |\n", nil},
		{"table marker", "notes.md", "", "| First |\n|---|\n| 1. " + words(25) + " |", nil},
		{"long cell", "notes.md", "", "| First | Second |\n|---|---|\n| " + words(26) + " | Fine. |\n", []string{"notes.md:3 holds a sentence of 26 words, more than 25"}},
		{"markup", "notes.md", "", "---\nsummary: " + long + "---\n# " + long + "<!-- " + long + "-->\n> 1. " + words(25), nil},
		{"quoted paragraph", "notes.md", "", "> " + half + "\n> " + words(15), []string{"notes.md:1 holds a sentence of 30 words, more than 25"}},
		{"HTML comment", "notes.md", "", "<!--\n" + long + "-->\n", nil},
		{"HTML fence", "notes.md", "", "<!--\n```\n-->\n" + long, []string{"notes.md:4 holds a sentence of 26 words, more than 25"}},
		{"code sentence start", "notes.md", "", words(25) + " `Command` " + words(24), nil},
		{"abbreviation", "notes.md", "", strings.TrimSuffix(words(24), ".") + " e.g. foo.", nil},
		{"lowercase sentence", "notes.md", "", words(25) + " identifier " + words(24), nil},
		{"quoted boundary", "notes.md", "", words(25) + "\" " + words(25), nil},
		{"emphasis boundary", "notes.md", "", words(25) + "* " + words(25), nil},
		{"closing punctuation", "notes.md", "", strings.TrimSuffix(words(25), ".") + "!')_ " + words(25), nil},
		{"uppercase sentence", "notes.md", "", words(25) + " Another " + words(24), nil},
		{"comment paragraph", "notes.go", "package notes\n// " + half + "\n", "package notes\n// " + half + "\n// " + words(15) + "\n", []string{"notes.go:2 holds a sentence of 30 words, more than 25"}},
		{"comment list", "notes.go", "", "package notes\n// Options:\n//   - " + words(30) + "\n", []string{"notes.go:3 holds a sentence of 30 words, more than 25"}},
		{"comment list continuation", "notes.go", "", "package notes\n//   - " + half + "\n//     " + words(15) + "\n", []string{"notes.go:2 holds a sentence of 30 words, more than 25"}},
		{"comment list paragraphs", "notes.go", "", "package notes\n// " + half + "\n//   - " + words(25) + "\n//   - " + words(25) + "\n", nil},
		{"comment code", "notes.go", "", "package notes\n// " + half + "\n//  " + long + "//\t" + long + "// " + half + "\n//\n// " + half + "\n//go:" + long + "//nolint:" + long + "//line " + long, nil},
		{"raw string", "notes.go", "", "package notes\nvar text = `\n// " + long + "`\n", nil},
		{"line directive", "notes.go", "", "package notes\n//line example.go:100\n// " + long, []string{"notes.go:3 holds a sentence of 26 words, more than 25"}},
		{"comment sentences", "notes.go", "package notes\n// " + words(26) + "\n", "package notes\n// " + words(26) + "\n// Another sentence.\n", nil},
		{"inline comment", "notes.go", "", "package notes\nvar x = 1 // " + long, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := map[string]string{}
			if tc.old != "" {
				before[tc.file] = tc.old
			}
			root := limitsRepository(t, before)
			write(t, filepath.Join(root, tc.file), tc.new)
			limitsTestGit(t, root, "add", ".")
			assertLimits(t, root, tc.want)
		})
	}
}
