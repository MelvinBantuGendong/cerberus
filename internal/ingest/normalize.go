package ingest

import (
	"strings"
	"unicode"
)

func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Cf, r):
			continue
		case unicode.IsSpace(r):
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(fold(r))
	}
	return strings.TrimSpace(b.String())
}

func fold(r rune) rune {
	if r >= 0xFF01 && r <= 0xFF5E { 
		return r - 0xFEE0
	}
	if c, ok := confusables[r]; ok {
		return c
	}
	return r
}

var confusables = map[rune]rune{
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x',
	'і': 'i', 'ѕ': 's', 'ј': 'j', 'к': 'k', 'т': 't', 'в': 'b', 'н': 'h',
	'м': 'm',


	'А': 'A', 'Е': 'E', 'О': 'O', 'Р': 'P', 'С': 'C', 'У': 'Y', 'Х': 'X',
	'І': 'I', 'Ј': 'J', 'К': 'K', 'Т': 'T', 'В': 'B', 'Н': 'H', 'М': 'M',
	'ο': 'o', 'ρ': 'p', 'ε': 'e', 'ι': 'i', 'ν': 'v', 'α': 'a', 'κ': 'k',
	'Α': 'A', 'Β': 'B', 'Ε': 'E', 'Η': 'H', 'Ι': 'I', 'Κ': 'K', 'Μ': 'M',
	'Ν': 'N', 'Ο': 'O', 'Ρ': 'P', 'Τ': 'T', 'Υ': 'Y', 'Χ': 'X', 'Ζ': 'Z',
}
