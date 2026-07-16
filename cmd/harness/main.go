package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/MelvinBantuGendong/cerberus/internal/corpus"
)

func main() {
	m := corpus.Evaluate(corpus.Samples())

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "CATEGORY\tTP\tFP\tTN\tFN\tCATCH\tFP-RATE")
	for _, cat := range m.Categories() {
		c := m.ByCategory[cat]
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%.0f%%\t%.0f%%\n",
			cat, c.TP, c.FP, c.TN, c.FN, c.CatchRate()*100, c.FPRate()*100)
	}
	o := m.Overall
	fmt.Fprintln(w, "\t\t\t\t\t\t")
	fmt.Fprintf(w, "OVERALL\t%d\t%d\t%d\t%d\t%.0f%%\t%.0f%%\n",
		o.TP, o.FP, o.TN, o.FN, o.CatchRate()*100, o.FPRate()*100)
	w.Flush()

	if len(m.Misses) > 0 {
		fmt.Printf("\nmissed attacks:  %v\n", m.Misses)
	}
	if len(m.FalseHits) > 0 {
		fmt.Printf("false positives: %v\n", m.FalseHits)
	}
}
