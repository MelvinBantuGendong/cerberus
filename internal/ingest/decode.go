package ingest

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	base64Run = regexp.MustCompile(`[A-Za-z0-9+/_-]{16,}={0,2}`)
	hexRun    = regexp.MustCompile(`(?:0x)?[0-9a-fA-F]{16,}`)
)

func decodeChildren(s Segment) []Segment {
	var kids []Segment
	seen := map[string]bool{s.Text: true}
	add := func(dec, kind string) {
		if dec == "" || seen[dec] {
			return
		}
		seen[dec] = true
		kids = append(kids, Segment{
			Text:  dec,
			Role:  s.Role,
			Trust: s.Trust,
			Path:  s.Path + "+" + kind,
			Depth: s.Depth + 1,
		})
	}
	for _, m := range hexRun.FindAllString(s.Text, -1) {
		if dec, ok := tryHex(m); ok {
			add(dec, "hex")
		}
	}
	for _, m := range base64Run.FindAllString(s.Text, -1) {
		if dec, ok := tryBase64(m); ok {
			add(dec, "base64")
		}
	}
	return kids
}

func tryBase64(s string) (string, bool) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil && printable(b) {
			return string(b), true
		}
	}
	return "", false
}

func tryHex(s string) (string, bool) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 != 0 {
		return "", false
	}
	b, err := hex.DecodeString(s)
	if err != nil || !printable(b) {
		return "", false
	}
	return string(b), true
}

func printable(b []byte) bool {
	if len(b) < 4 || !utf8.Valid(b) {
		return false
	}
	var ok, total int
	for _, r := range string(b) {
		if r == utf8.RuneError {
			return false
		}
		total++
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			ok++
		}
	}
	return total > 0 && ok*100/total >= 85
}
