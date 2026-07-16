package detect

func newJailbreakDetector() patternDetector {
	return patternDetector{
		category: "jailbreak",
		rules: []rule{
			mustRule("do_anything_now", 0.85, `(?i)do\s+anything\s+now`),
			mustRule("dan_mode", 0.8, `(?i)\bDAN\s+mode\b`),
			mustRule("no_restrictions", 0.85, `(?i)(you\s+(have\s+)?|with\s+)no\s+(restrictions|limits|limitations|rules|filters|guidelines)`),
			mustRule("ignore_safety", 0.85, `(?i)ignore\s+(your\s+|all\s+)?(safety|content|ethical)\s+(guidelines?|policy|policies|filters?|rules?)`),
			mustRule("act_as_unrestricted", 0.8, `(?i)act\s+as\s+(if\s+)?(an?\s+)?(unrestricted|uncensored|amoral|jailbroken)`),
			mustRule("unrestricted_mode", 0.8, `(?i)(enable|activate|enter|turn\s+on)\s+.{0,20}(jailbreak|unrestricted|uncensored|god)\s+mode`),
			mustRule("developer_mode", 0.7, `(?i)developer\s+mode`),
			mustRule("never_refuse", 0.7, `(?i)(never|do\s+not|don'?t)\s+refuse`),
			mustRule("pretend", 0.55, `(?i)pretend\s+(you\s+are|to\s+be|that)`),
		},
	}
}
