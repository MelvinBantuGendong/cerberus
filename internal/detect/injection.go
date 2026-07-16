package detect

func newInjectionDetector() patternDetector {
	return patternDetector{
		category: "prompt_injection",
		rules: []rule{
			mustRule("ignore_previous", 0.9, `(?i)ignore\s+(all\s+|the\s+|any\s+)*(previous|prior|above|preceding|earlier)\s+(instructions?|prompts?|directions?|messages?|context)`),
			mustRule("disregard_above", 0.9, `(?i)disregard\s+(all\s+|the\s+|any\s+)*(previous|prior|above|preceding|earlier)`),
			mustRule("forget_instructions", 0.85, `(?i)forget\s+(everything|all|your\s+(instructions|rules|prompt))`),
			mustRule("override_instructions", 0.85, `(?i)override\s+(your|the|all)\s+(previous\s+)?(instructions|rules|settings|prompt)`),
			mustRule("reveal_prompt", 0.7, `(?i)(reveal|repeat|print|show|output)\s+(your|the)\s+(system\s+prompt|initial\s+(instructions|prompt)|instructions)`),
			mustRule("new_instructions", 0.6, `(?i)(new|updated|revised)\s+instructions?\s*:`),
			mustRule("you_are_now", 0.6, `(?i)you\s+are\s+now\s+`),
		},
	}
}
