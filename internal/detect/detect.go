package detect

import (
	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

func Default() []ingest.Detector {
	return []ingest.Detector{
		newInjectionDetector(),
		newJailbreakDetector(),
		newIndirectDetector(),
		mixedScript{},
	}
}

type thresholds struct{ block, flag float64 }

var trustThresholds = map[verdict.TrustLevel]thresholds{
	verdict.Trusted:     {block: 0.95, flag: 0.80},
	verdict.SemiTrusted: {block: 0.80, flag: 0.50},
	verdict.Default:     {block: 0.80, flag: 0.50},
	verdict.Untrusted:   {block: 0.50, flag: 0.30},
}

func decide(score float64, trust verdict.TrustLevel) verdict.Action {
	th, ok := trustThresholds[trust]
	if !ok {
		th = trustThresholds[verdict.Default]
	}
	switch {
	case score >= th.block:
		return verdict.Block
	case score >= th.flag:
		return verdict.Flag
	default:
		return verdict.Allow
	}
}
