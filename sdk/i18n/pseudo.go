package i18n

import (
	"strings"
	"unicode/utf8"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Pseudo is the pseudo-locale: every catalog string, marked between ⟦ and ⟧,
// so a scenario finds the text that bypasses the catalog.
var Pseudo = language.MustParse("en-XA")

// pseudoMessage marks every plural form of an English message.
func pseudoMessage(m *goi18n.Message) *goi18n.Message {
	out := *m
	for _, form := range []*string{&out.Zero, &out.One, &out.Two, &out.Few, &out.Many, &out.Other} {
		if *form != "" {
			*form = pseudo(*form)
		}
	}
	return &out
}

var accents = strings.NewReplacer(
	"a", "á", "b", "ƀ", "c", "ç", "d", "ð", "e", "é", "f", "ƒ", "g", "ĝ", "h", "ĥ", "i", "î",
	"j", "ĵ", "k", "ķ", "l", "ļ", "m", "ɱ", "n", "ñ", "o", "ö", "p", "þ", "q", "ǫ", "r", "ŕ",
	"s", "š", "t", "ţ", "u", "û", "v", "ṽ", "w", "ŵ", "x", "ẋ", "y", "ý", "z", "ž",
	"A", "Á", "C", "Ç", "E", "É", "I", "Î", "N", "Ñ", "O", "Ö", "S", "Š", "U", "Û", "Y", "Ý",
)

// pseudo accents the text outside template actions, lengthens it by about a
// third as translations often do, and marks it between ⟦ and ⟧.
func pseudo(text string) string {
	var b strings.Builder
	b.WriteString("⟦")
	for rest := text; rest != ""; {
		before, action, found := strings.Cut(rest, "{{")
		b.WriteString(accents.Replace(before))
		if !found {
			break
		}
		action, rest, _ = strings.Cut(action, "}}")
		b.WriteString("{{" + action + "}}")
	}
	b.WriteString(strings.Repeat("·", (utf8.RuneCountInString(text)+2)/3))
	b.WriteString("⟧")
	return b.String()
}
