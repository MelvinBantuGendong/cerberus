package detect

// Info describes a toggleable detector so an admin UI can render controls and
// map user choices back to the store's disabled_rules set. The ID equals the
// detector's Check() category and is the key used in disabled_rules.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Direction   string `json:"direction"`
}

// Catalog returns the built-in detectors in display order. IDs here must match
// the ID() reported by each detector so toggles disable the right one.
func Catalog() []Info {
	return []Info{
		{
			ID:          "prompt_injection",
			Name:        "Direct Prompt Injection",
			Description: `Blocks instruction-override attempts such as "ignore previous instructions" or "reveal your system prompt".`,
			Direction:   "inbound",
		},
		{
			ID:          "jailbreak",
			Name:        "Jailbreak",
			Description: `Blocks DAN-style and "unrestricted mode" persona attacks that try to strip safety guidelines.`,
			Direction:   "inbound",
		},
		{
			ID:          "indirect_injection",
			Name:        "Indirect Injection",
			Description: `Catches instructions smuggled inside tool output or documents ("when you read this…", "do not tell the user").`,
			Direction:   "inbound",
		},
		{
			ID:          "obfuscation",
			Name:        "Mixed-Script Obfuscation",
			Description: "Flags tokens mixing Latin/Cyrillic/Greek characters used to sneak payloads past filters.",
			Direction:   "inbound",
		},
		{
			ID:          "pii_leak",
			Name:        "PII Leak",
			Description: "Blocks outgoing SSNs, emails, credit-card and phone numbers in model responses.",
			Direction:   "outbound",
		},
		{
			ID:          "secret_leak",
			Name:        "Secret Leak",
			Description: "Blocks outgoing API keys, private keys, JWTs and provider tokens.",
			Direction:   "outbound",
		},
		{
			ID:          "system_prompt_leak",
			Name:        "System-Prompt Leak",
			Description: "Detects the model echoing your system prompt back to the client.",
			Direction:   "outbound",
		},
	}
}
