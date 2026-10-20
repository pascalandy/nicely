package main

import (
	"go/scanner"
	"go/token"
	"path"
	"regexp"
	"strings"
	"unicode"
)

const sentenceFix = `Rewrite each sentence with the asd-ste100 skill, STE_11 and STE_09:
- Split it into sentences of one idea each, at most 25 words, or 20 for an instruction
- Turn a series of items into a list after an introductory sentence
- Keep every fact, number, condition, and "must"
Rule: docs/north-star/dev-preferences.md#write-clear-technical-english
Skill: https://github.com/pascalandy/skills/tree/main/skills/asd-ste100`

var (
	linkTarget  = regexp.MustCompile(`\]\([^)]*\)`)
	listMarker  = regexp.MustCompile(`^(?:[-*+] |\d+\. )`)
	goDirective = regexp.MustCompile(`^//(?:go:|nolint|line(?: |$))`)
	sentenceEnd = regexp.MustCompile(`[.!?]["'”’)*_]*[\s\pZ]+|\z`)
	wordToken   = regexp.MustCompile(`[^\s\pZ\v\x{85}]*[\pL\pN][^\s\pZ\v\x{85}]*`)
)

type sentence struct {
	text, preview string
	line, words   int
}

func extractSentences(file string, content []byte) []sentence {
	ext := path.Ext(strings.TrimSuffix(file, ".jinja"))
	if ext != ".md" && ext != ".go" {
		return nil
	}
	var found []sentence
	var paragraph, header []string
	first, listIndent := 1, -1
	flush := func() {
		found = append(found, splitSentences(strings.Join(paragraph, "\n"), first)...)
		paragraph = nil
		listIndent = -1
	}
	appendLine := func(text string, line int) {
		if len(paragraph) == 0 {
			first = line
		}
		paragraph = append(paragraph, text)
	}
	var f fence
	front, html := false, false
	comments := map[int]bool{}
	if ext == ".go" {
		var s scanner.Scanner
		file := token.NewFileSet().AddFile(file, -1, len(content))
		s.Init(file, content, nil, scanner.ScanComments)
		for {
			pos, kind, _ := s.Scan()
			if kind == token.EOF {
				break
			}
			if kind == token.COMMENT {
				comments[file.PositionFor(pos, false).Line] = true
			}
		}
	}
	for i, line := range strings.Split(string(content), "\n") {
		text := strings.TrimSpace(line)
		if ext == ".go" {
			if !comments[i+1] || !strings.HasPrefix(text, "//") || text == "//" || goDirective.MatchString(text) {
				flush()
				continue
			}
			body := text[2:]
			text = strings.TrimSpace(body)
			indent := len(body) - len(strings.TrimLeft(body, " \t"))
			switch {
			case listMarker.MatchString(text):
				flush()
				listIndent = indent
				text = listMarker.ReplaceAllString(text, "")
			case listIndent >= 0 && indent > listIndent:
			case strings.HasPrefix(body, "\t") || indent >= 2:
				flush()
				continue
			case listIndent >= 0:
				flush()
			}
			appendLine(text, i+1)
			continue
		}
		for {
			quoted := strings.TrimLeft(line, " \t")
			if !strings.HasPrefix(quoted, ">") {
				break
			}
			line = strings.TrimPrefix(quoted[1:], " ")
		}
		text = strings.TrimSpace(line)
		if i == 0 && text == "---" {
			front = true
			continue
		}
		if front {
			front = text != "---"
			continue
		}
		if html {
			html = !strings.Contains(text, "-->")
			continue
		}
		if f.in(line) || text == "" || strings.HasPrefix(text, "#") {
			flush()
			continue
		}
		if strings.HasPrefix(text, "<!--") {
			flush()
			html = !strings.Contains(text, "-->")
			continue
		}
		cells, _ := tableRow(line, &header)
		if strings.HasPrefix(text, "|") {
			flush()
			for _, cell := range cells {
				found = append(found, splitSentences(listMarker.ReplaceAllString(cell, ""), i+1)...)
			}
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		switch {
		case listMarker.MatchString(text):
			flush()
			listIndent = indent
			text = listMarker.ReplaceAllString(text, "")
		case listIndent >= 0 && indent <= listIndent:
			flush()
		}
		appendLine(text, i+1)
	}
	flush()
	return found
}

// maskProse preserves line positions while hiding syntax from the sentence splitter.
func maskProse(text string, re *regexp.Regexp, word bool) string {
	return re.ReplaceAllStringFunc(text, func(s string) string {
		b := []byte(s)
		for i := range b {
			if b[i] != '\n' {
				b[i] = ' '
			}
		}
		if word && len(b) > 0 {
			b[0] = 'X'
		}
		return string(b)
	})
}

func splitSentences(text string, first int) []sentence {
	original := text
	labels := map[int]string{}
	for _, span := range codeSpan.FindAllStringIndex(text, -1) {
		labels[span[0]] = text[span[0]:span[1]]
	}
	text = maskProse(maskProse(text, codeSpan, true), linkTarget, false)
	var found []sentence
	start := 0
	for _, boundary := range sentenceEnd.FindAllStringIndex(text, -1) {
		end := boundary[1]
		start += end - start - len(strings.TrimLeftFunc(text[start:end], unicode.IsSpace))
		tokens := wordToken.FindAllStringIndex(text[start:end], -1)
		if len(tokens) > 0 {
			var preview []string
			for _, word := range tokens[:min(12, len(tokens))] {
				at := start + word[0]
				shown, ok := labels[at]
				if !ok {
					shown = original[at : start+word[1]]
				}
				preview = append(preview, shown)
			}
			found = append(found, sentence{
				text:    strings.Join(strings.Fields(linkTarget.ReplaceAllString(original[start:end], "]")), " "),
				preview: strings.Join(preview, " "),
				line:    first + strings.Count(text[:start], "\n"),
				words:   len(tokens),
			})
		}
		start = end
	}
	return found
}
