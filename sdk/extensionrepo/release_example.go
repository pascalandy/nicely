package extensionrepo

import (
	"fmt"
	"slices"
	"strings"
)

// releaseExample reads literal shell words, without expansion or a shell process.
func releaseExample(line string) ([]string, error) {
	var words []string
	var word strings.Builder
	var quote byte
	started := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && quote != '\'':
			i++
			if i == len(line) {
				return nil, fmt.Errorf("unfinished escape")
			}
			word.WriteByte(line[i])
			started = true
		case quote != 0:
			if ch == quote {
				quote = 0
			} else {
				word.WriteByte(ch)
			}
		case ch == '\'' || ch == '"':
			quote, started = ch, true
		case strings.ContainsRune(" \t\r\n", rune(ch)):
			if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
		case strings.ContainsRune(";|&<>`", rune(ch)):
			return nil, fmt.Errorf("shell operators are unsupported")
		default:
			word.WriteByte(ch)
			started = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unfinished quote")
	}
	if started {
		words = append(words, word.String())
	}
	if len(words) < 2 || words[0] != "ncly" {
		return nil, fmt.Errorf("want a command starting with ncly")
	}
	// Select each output and execution mode ourselves, before the argument delimiter.
	var call []string
	options := true
	for _, word := range words {
		if word == "--" {
			options = false
		}
		if options && (word == "--json" || strings.HasPrefix(word, "--json=") || word == "--dry-run" || strings.HasPrefix(word, "--dry-run=")) {
			continue
		}
		call = append(call, word)
	}
	return call, nil
}

func releaseFlags(call []string, flags ...string) []string {
	at := slices.Index(call, "--")
	if at < 0 {
		at = len(call)
	}
	return slices.Insert(slices.Clone(call), at, flags...)
}
