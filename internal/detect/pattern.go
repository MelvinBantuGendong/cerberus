package detect

import (
	"regexp"
	"sort"

	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type rule struct {
	name   string
	weight float64
	re     *regexp.Regexp
}

type patternDetector struct {
	category  string
	direction verdict.Direction
	rules     []rule
}

func (d patternDetector) Check(s ingest.Segment) verdict.Verdict {
	var score float64
	var matched []string
	for _, r := range d.rules {
		if r.re.MatchString(s.Text) {
			matched = append(matched, r.name)
			if r.weight > score {
				score = r.weight
			}
		}
	}
	if len(matched) == 0 {
		return verdict.Allowing(d.direction, s.Trust)
	}
	sort.Strings(matched)
	return verdict.Verdict{
		Action:       decide(score, s.Trust),
		Score:        score,
		Categories:   []string{d.category},
		MatchedRules: matched,
		Direction:    d.direction,
		TrustLevel:   s.Trust,
	}
}

func mustRule(name string, weight float64, pattern string) rule {
	return rule{name: name, weight: weight, re: regexp.MustCompile(pattern)}
}
