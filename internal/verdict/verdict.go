package verdict

type Action string

const (
	Allow Action = "allow"
	Block Action = "block"
	Flag  Action = "flag"
)

type Direction string

const (
	Inbound  Direction = "inbound"
	Outbound Direction = "outbound"
)

type TrustLevel string

const (
	Trusted     TrustLevel = "trusted"
	SemiTrusted TrustLevel = "semi_trusted"
	Untrusted   TrustLevel = "untrusted"
	Default     TrustLevel = "default"
)

type Verdict struct {
	Action       Action     `json:"action"`
	Score        float64    `json:"score"`
	Categories   []string   `json:"categories"`
	MatchedRules []string   `json:"matched_rules"`
	Direction    Direction  `json:"direction"`
	TrustLevel   TrustLevel `json:"trust_level"`
}

func Allowing(dir Direction, trust TrustLevel) Verdict {
	return Verdict{
		Action:       Allow,
		Categories:   []string{},
		MatchedRules: []string{},
		Direction:    dir,
		TrustLevel:   trust,
	}
}
