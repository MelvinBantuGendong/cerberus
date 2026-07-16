package corpus

import (
	"sort"

	"github.com/MelvinBantuGendong/cerberus/internal/detect"
	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type Counts struct{ TP, FP, TN, FN int }

func (c Counts) CatchRate() float64 {
	if c.TP+c.FN == 0 {
		return 0
	}
	return float64(c.TP) / float64(c.TP+c.FN)
}

func (c Counts) FPRate() float64 {
	if c.FP+c.TN == 0 {
		return 0
	}
	return float64(c.FP) / float64(c.FP+c.TN)
}

type Matrix struct {
	Overall    Counts
	ByCategory map[string]*Counts
	Misses     []string
	FalseHits  []string
}

func blocked(s Sample) bool {
	var v verdict.Verdict
	if s.Direction == verdict.Outbound {
		v = ingest.Dispatch(verdict.Outbound, ingest.FromText(s.Text), detect.Outbound(s.System))
	} else {
		v = ingest.Dispatch(verdict.Inbound, ingest.Segments(s.Text, s.Trust), detect.Default())
	}
	return v.Action == verdict.Block
}

func Evaluate(samples []Sample) Matrix {
	m := Matrix{ByCategory: map[string]*Counts{}}
	for _, s := range samples {
		c := m.ByCategory[s.Category]
		if c == nil {
			c = &Counts{}
			m.ByCategory[s.Category] = c
		}
		b := blocked(s)
		switch {
		case s.Attack && b:
			m.Overall.TP++
			c.TP++
		case s.Attack && !b:
			m.Overall.FN++
			c.FN++
			m.Misses = append(m.Misses, s.Name)
		case !s.Attack && b:
			m.Overall.FP++
			c.FP++
			m.FalseHits = append(m.FalseHits, s.Name)
		default:
			m.Overall.TN++
			c.TN++
		}
	}
	return m
}

func (m Matrix) Categories() []string {
	out := make([]string, 0, len(m.ByCategory))
	for k := range m.ByCategory {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
