package corpus

import "testing"

func TestConfusionMatrix(t *testing.T) {
	m := Evaluate(Samples())
	o := m.Overall

	t.Logf("overall: TP=%d FP=%d TN=%d FN=%d  catch=%.0f%%  fp=%.0f%%",
		o.TP, o.FP, o.TN, o.FN, o.CatchRate()*100, o.FPRate()*100)
	for _, cat := range m.Categories() {
		c := m.ByCategory[cat]
		t.Logf("  %-20s TP=%d FP=%d TN=%d FN=%d", cat, c.TP, c.FP, c.TN, c.FN)
	}
	if len(m.Misses) > 0 {
		t.Logf("  missed attacks: %v", m.Misses)
	}
	if len(m.FalseHits) > 0 {
		t.Logf("  false positives: %v", m.FalseHits)
	}

	if r := o.CatchRate(); r < 0.90 {
		t.Errorf("catch rate %.2f below 0.90 (missed: %v)", r, m.Misses)
	}
	if r := o.FPRate(); r > 0.05 {
		t.Errorf("false-positive rate %.2f above 0.05 (false hits: %v)", r, m.FalseHits)
	}
}
