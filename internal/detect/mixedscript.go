package detect

import (
	"unicode"

	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type mixedScript struct{}

func (mixedScript) ID() string { return "obfuscation" }

func (mixedScript) Check(s ingest.Segment) verdict.Verdict {
	if !hasMixedScriptToken(s.Text) {
		return verdict.Allowing(verdict.Inbound, s.Trust)
	}
	const score = 0.8
	return verdict.Verdict{
		Action:       decide(score, s.Trust),
		Score:        score,
		Categories:   []string{"obfuscation"},
		MatchedRules: []string{"mixed_script"},
		Direction:    verdict.Inbound,
		TrustLevel:   s.Trust,
	}
}

func hasMixedScriptToken(text string) bool {
	var latin, cyrillic, greek bool
	mixed := func() bool {
		n := 0
		for _, ok := range []bool{latin, cyrillic, greek} {
			if ok {
				n++
			}
		}
		return n > 1
	}
	for _, r := range text {
		if !unicode.IsLetter(r) {
			if mixed() {
				return true
			}
			latin, cyrillic, greek = false, false, false
			continue
		}
		switch {
		case unicode.Is(unicode.Latin, r):
			latin = true
		case unicode.Is(unicode.Cyrillic, r):
			cyrillic = true
		case unicode.Is(unicode.Greek, r):
			greek = true
		}
	}
	return mixed()
}
