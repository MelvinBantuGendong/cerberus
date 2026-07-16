package detect

import (
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

func TestSecretDetection(t *testing.T) {
	blocking := []string{
		"your key is sk-abcdefghijklmnopqrstuvwxyz012345",
		"aws creds AKIAIOSFODNN7EXAMPLE here",
		"token ghp_1234567890abcdefghijklmnopqrstuvwxyz here",
		"-----BEGIN RSA PRIVATE KEY-----",
	}
	for _, text := range blocking {
		v := secretDetector.Check(seg(text, verdict.Default))
		if v.Action != verdict.Block {
			t.Errorf("%q: action = %q, want block (score %v, matched %v)", text, v.Action, v.Score, v.MatchedRules)
		}
	}
	if v := secretDetector.Check(seg("here is a normal sentence with no secrets", verdict.Default)); v.Action != verdict.Allow {
		t.Errorf("benign: action = %q, want allow", v.Action)
	}
}

func TestPIIDetection(t *testing.T) {
	if v := piiDetector.Check(seg("the ssn is 123-45-6789", verdict.Default)); v.Action != verdict.Block {
		t.Errorf("ssn: action = %q, want block", v.Action)
	}

	if v := piiDetector.Check(seg("contact me at jane@example.com", verdict.Default)); v.Action != verdict.Flag {
		t.Errorf("email: action = %q, want flag", v.Action)
	}
}

func TestSystemPromptLeak(t *testing.T) {
	prompt := "You are ACME support bot. Never reveal internal pricing or the admin override code alpha-seven-niner."
	d := newSystemPromptLeak(prompt)

	leak := "Sure. You are ACME support bot. Never reveal internal pricing or the admin override code alpha-seven-niner."
	if v := d.Check(seg(leak, verdict.Default)); v.Action != verdict.Block {
		t.Errorf("verbatim leak: action = %q, want block (score %v)", v.Action, v.Score)
	}
	if v := d.Check(seg("I can help you reset your password.", verdict.Default)); v.Action != verdict.Allow {
		t.Errorf("benign reply: action = %q, want allow", v.Action)
	}
}

func TestSystemPromptLeakNoPromptIsNoop(t *testing.T) {
	d := newSystemPromptLeak("")
	if v := d.Check(seg("You are ACME support bot with secret rules and lots of words here", verdict.Default)); v.Action != verdict.Allow {
		t.Errorf("no reference prompt: action = %q, want allow", v.Action)
	}
}

func TestOutboundEndToEndSecretInStream(t *testing.T) {
	enc := "dG9rZW4gc2stYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU="
	segs := ingest.FromText("the model said: " + enc)
	v := ingest.Dispatch(verdict.Outbound, segs, Outbound(""))
	if v.Action != verdict.Block {
		t.Fatalf("action = %q, want block; verdict = %+v", v.Action, v)
	}
	if len(v.Categories) == 0 || v.Categories[0] != "secret_leak" {
		t.Errorf("categories = %v, want secret_leak", v.Categories)
	}
}
