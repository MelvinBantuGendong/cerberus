package ingest

import (
	"sort"

	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type Detector interface {
	Check(Segment) verdict.Verdict
}

func Dispatch(dir verdict.Direction, segs []Segment, dets []Detector) verdict.Verdict {
	out := verdict.Verdict{
		Action:       verdict.Allow,
		Direction:    dir,
		TrustLevel:   verdict.Default,
		Categories:   []string{},
		MatchedRules: []string{},
	}
	cats := map[string]struct{}{}
	rules := map[string]struct{}{}
	for _, seg := range segs {
		for _, d := range dets {
			v := d.Check(seg)
			if v.Score > out.Score {
				out.Score = v.Score
			}
			if severity(v.Action) > severity(out.Action) {
				out.Action = v.Action
				out.TrustLevel = seg.Trust
			}
			for _, c := range v.Categories {
				cats[c] = struct{}{}
			}
			for _, r := range v.MatchedRules {
				rules[r] = struct{}{}
			}
		}
	}
	out.Categories = sortedKeys(cats)
	out.MatchedRules = sortedKeys(rules)
	return out
}

func severity(a verdict.Action) int {
	switch a {
	case verdict.Block:
		return 2
	case verdict.Flag:
		return 1
	default:
		return 0
	}
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
