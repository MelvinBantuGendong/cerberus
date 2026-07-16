package detect

import (
	"strings"

	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

func Outbound(systemPrompt string) []ingest.Detector {
	return []ingest.Detector{
		piiDetector,
		secretDetector,
		newSystemPromptLeak(systemPrompt),
	}
}

var piiDetector = patternDetector{
	category: "pii_leak",
	rules: []rule{
		mustRule("ssn", 0.85, `\b\d{3}-\d{2}-\d{4}\b`),
		mustRule("email", 0.6, `\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
		mustRule("credit_card", 0.7, `\b(?:\d[ -]?){13,16}\b`),
		mustRule("phone", 0.5, `\b(?:\+?1[ -]?)?\(?\d{3}\)?[ -]?\d{3}[ -]?\d{4}\b`),
	},
}

var secretDetector = patternDetector{
	category: "secret_leak",
	rules: []rule{
		mustRule("private_key", 0.97, `-----BEGIN (?:RSA |EC |OPENSSH |DSA |PGP )?PRIVATE KEY-----`),
		mustRule("openai_key", 0.95, `\bsk-[A-Za-z0-9]{20,}\b`),
		mustRule("aws_access_key", 0.95, `\bAKIA[0-9A-Z]{16}\b`),
		mustRule("github_token", 0.95, `\bgh[pousr]_[A-Za-z0-9]{36,}\b`),
		mustRule("slack_token", 0.9, `\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),
		mustRule("jwt", 0.85, `\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`),
	},
}

type systemPromptLeak struct {
	shingles map[string]struct{}
}

const leakShingle = 8

func newSystemPromptLeak(systemPrompt string) systemPromptLeak {
	words := strings.Fields(ingest.Normalize(strings.ToLower(systemPrompt)))
	n := leakShingle
	if len(words) < n {
		n = len(words)
	}
	set := map[string]struct{}{}
	if n >= 4 {
		for i := 0; i+n <= len(words); i++ {
			set[strings.Join(words[i:i+n], " ")] = struct{}{}
		}
	}
	return systemPromptLeak{shingles: set}
}

func (d systemPromptLeak) Check(s ingest.Segment) verdict.Verdict {
	if len(d.shingles) == 0 {
		return verdict.Allowing(verdict.Outbound, s.Trust)
	}
	words := strings.Fields(strings.ToLower(s.Text))
	if len(words) < leakShingle {
		return verdict.Allowing(verdict.Outbound, s.Trust)
	}
	matched := 0
	for i := 0; i+leakShingle <= len(words); i++ {
		if _, ok := d.shingles[strings.Join(words[i:i+leakShingle], " ")]; ok {
			matched++
		}
	}
	if matched == 0 {
		return verdict.Allowing(verdict.Outbound, s.Trust)
	}
	score := 0.5 + 0.15*float64(matched)
	if score > 0.95 {
		score = 0.95
	}
	return verdict.Verdict{
		Action:       decide(score, s.Trust),
		Score:        score,
		Categories:   []string{"system_prompt_leak"},
		MatchedRules: []string{"system_prompt_echo"},
		Direction:    verdict.Outbound,
		TrustLevel:   s.Trust,
	}
}
