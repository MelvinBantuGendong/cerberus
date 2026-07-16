package verdict

import (
	"encoding/json"
	"testing"
)

func TestAllowingMatchesContract(t *testing.T) {
	b, err := json.Marshal(Allowing(Inbound, Default))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"action":"allow","score":0,"categories":[],"matched_rules":[],"direction":"inbound","trust_level":"default"}`
	if got := string(b); got != want {
		t.Errorf("verdict JSON =\n  %s\nwant\n  %s", got, want)
	}
}
