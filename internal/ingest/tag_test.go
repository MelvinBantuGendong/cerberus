package ingest

import (
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

func TestTrustTagOverridesRole(t *testing.T) {
	body := []byte(`{"messages":[
		{"role":"system","content":"you are helpful"},
		{"role":"user","content":"retrieved doc text","cerberus":{"trust":"untrusted"}}
	]}`)
	req, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	segs := Extract(req)

	s, ok := findByPath(segs, "messages[1].content")
	if !ok {
		t.Fatal("missing tagged segment")
	}
	if s.Trust != verdict.Untrusted {
		t.Errorf("tagged trust = %q, want untrusted", s.Trust)
	}
}

func TestTrustTagUnknownFallsBackToRole(t *testing.T) {
	body := []byte(`{"messages":[
		{"role":"system","content":"sys","cerberus":{"trust":"bogus"}}
	]}`)
	req, _ := Parse(body)
	segs := Extract(req)
	s, ok := findByPath(segs, "messages[0].content")
	if !ok {
		t.Fatal("missing segment")
	}

	if s.Trust != verdict.Trusted {
		t.Errorf("trust = %q, want trusted (role fallback)", s.Trust)
	}
}

func TestParseTrust(t *testing.T) {
	for _, ok := range []string{"trusted", "semi_trusted", "untrusted", "default"} {
		if _, valid := ParseTrust(ok); !valid {
			t.Errorf("ParseTrust(%q) = invalid, want valid", ok)
		}
	}
	if _, valid := ParseTrust("elevated"); valid {
		t.Error("ParseTrust(\"elevated\") = valid, want invalid")
	}
}
