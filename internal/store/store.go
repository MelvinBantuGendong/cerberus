package store

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
)

type Key struct {
	ID      string    `json:"id"`
	Label   string    `json:"label"`
	Hash    string    `json:"hash"`
	Created time.Time `json:"created"`
}

type settings struct {
	Upstream      string              `json:"upstream"`
	UpstreamKey   string              `json:"upstream_key"`
	MaxBodyBytes  int64               `json:"max_body_bytes"`
	OutboundMode  config.OutboundMode `json:"outbound_mode"`
	DisabledRules []string            `json:"disabled_rules"`
}

type persisted struct {
	Settings settings `json:"settings"`
	Keys     []Key    `json:"keys"`
}

type Store struct {
	mu       sync.RWMutex
	path     string
	set      settings
	upstream *url.URL
	keys     []Key
}

type Snapshot struct {
	Upstream      *url.URL
	UpstreamKey   string
	MaxBodyBytes  int64
	OutboundMode  config.OutboundMode
	DisabledRules map[string]bool
	AuthEnabled   bool
}

func FromConfig(cfg config.Config) (*Store, error) {
	boot := settings{
		Upstream:     cfg.UpstreamBase.String(),
		UpstreamKey:  cfg.UpstreamKey,
		MaxBodyBytes: cfg.MaxBodyBytes,
		OutboundMode: cfg.OutboundMode,
	}
	return newStore(cfg.StatePath, boot, cfg.APIKeys)
}

func newStore(path string, boot settings, bootKeys []string) (*Store, error) {
	s := &Store{path: path}
	if path != "" {
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			var p persisted
			if err := json.Unmarshal(data, &p); err != nil {
				return nil, fmt.Errorf("state file %s: %w", path, err)
			}
			if err := s.apply(p.Settings, p.Keys); err != nil {
				return nil, err
			}
			return s, nil
		case !errors.Is(err, os.ErrNotExist):
			return nil, err
		}
	}

	var keys []Key
	for _, k := range bootKeys {
		if k = strings.TrimSpace(k); k != "" {
			keys = append(keys, Key{ID: randID(), Label: "bootstrap", Hash: hashKey(k), Created: time.Now().UTC()})
		}
	}
	if err := s.apply(boot, keys); err != nil {
		return nil, err
	}
	return s, s.save()
}

func (s *Store) apply(set settings, keys []Key) error {
	u, err := url.Parse(set.Upstream)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("upstream %q is not an absolute URL", set.Upstream)
	}
	if set.MaxBodyBytes <= 0 {
		set.MaxBodyBytes = 4 << 20
	}
	switch set.OutboundMode {
	case config.OutboundOff, config.OutboundBuffer, config.OutboundStream:
	default:
		set.OutboundMode = config.OutboundBuffer
	}
	s.set = set
	s.upstream = u
	s.keys = keys
	return nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(persisted{Settings: s.set, Keys: s.keys}, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{
		Upstream:      s.upstream,
		UpstreamKey:   s.set.UpstreamKey,
		MaxBodyBytes:  s.set.MaxBodyBytes,
		OutboundMode:  s.set.OutboundMode,
		DisabledRules: disabledSet(s.set.DisabledRules),
		AuthEnabled:   len(s.keys) > 0,
	}
}

func (s *Store) ValidKey(token string) bool {
	h := hashKey(token)
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ok bool
	for _, k := range s.keys {
		if subtle.ConstantTimeCompare([]byte(h), []byte(k.Hash)) == 1 {
			ok = true
		}
	}
	return ok
}

type View struct {
	Upstream       string              `json:"upstream"`
	UpstreamKeySet bool                `json:"upstream_key_set"`
	MaxBodyBytes   int64               `json:"max_body_bytes"`
	OutboundMode   config.OutboundMode `json:"outbound_mode"`
	DisabledRules  []string            `json:"disabled_rules"`
	KeyCount       int                 `json:"key_count"`
}

func (s *Store) View() View {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]string, len(s.set.DisabledRules))
	copy(rules, s.set.DisabledRules)
	return View{
		Upstream:       s.set.Upstream,
		UpstreamKeySet: s.set.UpstreamKey != "",
		MaxBodyBytes:   s.set.MaxBodyBytes,
		OutboundMode:   s.set.OutboundMode,
		DisabledRules:  rules,
		KeyCount:       len(s.keys),
	}
}

func normalizeRules(rules []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		r = strings.TrimSpace(r)
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func disabledSet(rules []string) map[string]bool {
	if len(rules) == 0 {
		return nil
	}
	m := make(map[string]bool, len(rules))
	for _, r := range rules {
		m[r] = true
	}
	return m
}

type SettingsPatch struct {
	Upstream      *string
	UpstreamKey   *string
	MaxBodyBytes  *int64
	OutboundMode  *config.OutboundMode
	DisabledRules *[]string
}

func (s *Store) UpdateSettings(p SettingsPatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.set
	if p.Upstream != nil {
		next.Upstream = *p.Upstream
	}
	if p.UpstreamKey != nil {
		next.UpstreamKey = *p.UpstreamKey
	}
	if p.MaxBodyBytes != nil {
		next.MaxBodyBytes = *p.MaxBodyBytes
	}
	if p.OutboundMode != nil {
		next.OutboundMode = *p.OutboundMode
	}
	if p.DisabledRules != nil {
		next.DisabledRules = normalizeRules(*p.DisabledRules)
	}

	u, err := url.Parse(next.Upstream)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("upstream %q is not an absolute URL", next.Upstream)
	}
	if next.MaxBodyBytes <= 0 {
		return errors.New("max_body_bytes must be positive")
	}
	switch next.OutboundMode {
	case config.OutboundOff, config.OutboundBuffer, config.OutboundStream:
	default:
		return fmt.Errorf("invalid outbound_mode %q", next.OutboundMode)
	}

	s.set = next
	s.upstream = u
	return s.save()
}

func (s *Store) GenerateKey(label string) (token string, k Key, err error) {
	token, err = randToken()
	if err != nil {
		return "", Key{}, err
	}
	k = Key{ID: randID(), Label: label, Hash: hashKey(token), Created: time.Now().UTC()}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys = append(s.keys, k)
	if err := s.save(); err != nil {
		s.keys = s.keys[:len(s.keys)-1]
		return "", Key{}, err
	}
	return token, k, nil
}

func (s *Store) RevokeKey(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, k := range s.keys {
		if k.ID == id {
			old := s.keys
			next := make([]Key, 0, len(old)-1)
			next = append(next, old[:i]...)
			next = append(next, old[i+1:]...)
			s.keys = next
			if err := s.save(); err != nil {
				s.keys = old
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) ListKeys() []Key {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Key, 0, len(s.keys))
	for _, k := range s.keys {
		out = append(out, Key{ID: k.ID, Label: k.Label, Created: k.Created})
	}
	return out
}

func hashKey(k string) string {
	sum := sha256.Sum256([]byte(k))
	return hex.EncodeToString(sum[:])
}

func randToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cbk_" + hex.EncodeToString(b), nil
}

func randID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
