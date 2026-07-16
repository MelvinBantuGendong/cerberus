package detect

import "github.com/MelvinBantuGendong/cerberus/internal/verdict"

func newIndirectDetector() patternDetector {
	return patternDetector{
		category:  "indirect_injection",
		direction: verdict.Inbound,
		rules: []rule{
			mustRule("addresses_the_ai", 0.7, `(?i)\b(attention|note|important|message|instruction)\b.{0,20}\b(ai|assistant|llm|model|chatbot|language model)\b`),
			mustRule("when_you_read_this", 0.7, `(?i)\b(when|if|whenever|after)\s+you\s+(read|process|see|encounter|receive|parse)\b`),
			mustRule("override_user", 0.75, `(?i)ignore\s+(the\s+|what\s+the\s+)?(user|human|person)`),
			mustRule("impersonate_role", 0.65, `(?im)^\s*(system|assistant)\s*:`),
			mustRule("instruction_header", 0.65, `(?i)#{2,}\s*(system|instruction|new\s+instruction|prompt)`),
			mustRule("act_on_behalf", 0.65, `(?i)(the\s+)?(user|human)\s+(actually|really|secretly)\s+(wants|needs|asked)`),
			mustRule("hide_from_user", 0.85, `(?i)(do not|don'?t|never)\s+(tell|inform|reveal\s+to|mention\s+to|notify)\s+(the\s+)?(user|human|person)`),
			mustRule("exfiltrate", 0.85, `(?i)\b(send|forward|email|post|upload|exfiltrate|leak)\b.{0,40}(https?://|www\.|@|api[_ ]?key|token|secret|password)`),
		},
	}
}
