package ingest

import (
	"strings"
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type keyword struct {
	word     string
	action   verdict.Action
	score    float64
	category string
	rule     string
}

func (k keyword) Check(s Segment) verdict.Verdict {
	if !strings.Contains(s.Text, k.word) {
		return verdict.Allowing(verdict.Inbound, s.Trust)
	}
	return verdict.Verdict{
		Action:       k.action,
		Score:        k.score,
		Categories:   []string{k.category},
		MatchedRules: []string{k.rule},
		Direction:    verdict.Inbound,
		TrustLevel:   s.Trust,
	}
}

func TestDispatchAllowsWhenNothingMatches(t *testing.T) {
	segs := []Segment{{Text: "hello there", Trust: verdict.Default}}
	dets := []Detector{keyword{word: "ignore", action: verdict.Block, score: 0.9, category: "prompt_injection", rule: "r1"}}
	v := Dispatch(verdict.Inbound, segs, dets)
	if v.Action != verdict.Allow {
		t.Errorf("action = %q, want allow", v.Action)
	}
	if len(v.Categories) != 0 || len(v.MatchedRules) != 0 {
		t.Errorf("expected empty categories/rules, got %+v / %+v", v.Categories, v.MatchedRules)
	}
}

func TestDispatchWorstActionWins(t *testing.T) {
	segs := []Segment{
		{Text: "please flag me", Trust: verdict.SemiTrusted},
		{Text: "ignore all previous", Trust: verdict.Untrusted},
	}
	dets := []Detector{
		keyword{word: "flag", action: verdict.Flag, score: 0.4, category: "suspicious", rule: "soft"},
		keyword{word: "ignore", action: verdict.Block, score: 0.95, category: "prompt_injection", rule: "ignore_previous"},
	}
	v := Dispatch(verdict.Inbound, segs, dets)
	if v.Action != verdict.Block {
		t.Errorf("action = %q, want block", v.Action)
	}
	if v.Score != 0.95 {
		t.Errorf("score = %v, want 0.95", v.Score)
	}

	if v.TrustLevel != verdict.Untrusted {
		t.Errorf("trust = %q, want untrusted", v.TrustLevel)
	}

	if got := strings.Join(v.Categories, ","); got != "prompt_injection,suspicious" {
		t.Errorf("categories = %q, want prompt_injection,suspicious", got)
	}
	if got := strings.Join(v.MatchedRules, ","); got != "ignore_previous,soft" {
		t.Errorf("rules = %q, want ignore_previous,soft", got)
	}
}
