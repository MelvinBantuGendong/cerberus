package store

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
)

func testConfig(keys ...string) config.Config {
	u, _ := url.Parse("https://upstream.example/v1")
	return config.Config{
		UpstreamBase: u,
		UpstreamKey:  "sk-upstream",
		MaxBodyBytes: 1 << 20,
		OutboundMode: config.OutboundBuffer,
		APIKeys:      keys,
	}
}

func TestKeyLifecycle(t *testing.T) {
	s, err := FromConfig(testConfig())
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if s.Snapshot().AuthEnabled {
		t.Error("auth should be disabled with no keys")
	}

	token, k, err := s.GenerateKey("ci")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if !s.ValidKey(token) {
		t.Error("generated key should validate")
	}
	if s.ValidKey("cbk_wrong") {
		t.Error("wrong token should not validate")
	}
	if !s.Snapshot().AuthEnabled {
		t.Error("auth should be enabled after issuing a key")
	}

	if ok, err := s.RevokeKey(k.ID); err != nil || !ok {
		t.Errorf("revoke should succeed: ok=%v err=%v", ok, err)
	}
	if s.ValidKey(token) {
		t.Error("revoked key should not validate")
	}
}

func TestBootstrapKeysAreHashed(t *testing.T) {
	s, err := FromConfig(testConfig("bootstrap-secret"))
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if !s.ValidKey("bootstrap-secret") {
		t.Error("bootstrapped key should validate")
	}
	for _, k := range s.ListKeys() {
		if k.Hash != "" {
			t.Error("ListKeys must not expose the hash")
		}
	}
}

func TestUpdateSettings(t *testing.T) {
	s, _ := FromConfig(testConfig())

	newUp := "https://other.example/v1"
	off := config.OutboundOff
	if err := s.UpdateSettings(SettingsPatch{Upstream: &newUp, OutboundMode: &off}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	snap := s.Snapshot()
	if snap.Upstream.String() != newUp {
		t.Errorf("upstream = %q, want %q", snap.Upstream.String(), newUp)
	}
	if snap.OutboundMode != config.OutboundOff {
		t.Errorf("mode = %q, want off", snap.OutboundMode)
	}

	bad := "not-a-url"
	if err := s.UpdateSettings(SettingsPatch{Upstream: &bad}); err == nil {
		t.Error("expected error for invalid upstream")
	}
	if s.Snapshot().Upstream.String() != newUp {
		t.Error("rejected update mutated state")
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	cfg := testConfig()
	cfg.StatePath = path

	s1, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	token, _, _ := s1.GenerateKey("persisted")
	mode := config.OutboundStream
	if err := s1.UpdateSettings(SettingsPatch{OutboundMode: &mode}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	if info, err := os.Stat(path); err != nil {
		t.Fatalf("state file not written: %v", err)
	} else if info.Mode().Perm() != 0o600 {
		t.Errorf("state file perms = %v, want 0600", info.Mode().Perm())
	}

	s2, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !s2.ValidKey(token) {
		t.Error("persisted key did not survive reload")
	}
	if s2.Snapshot().OutboundMode != config.OutboundStream {
		t.Error("persisted setting did not survive reload")
	}
}
