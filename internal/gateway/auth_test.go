package gateway

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/store"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func mustStore(t *testing.T, cfg config.Config) *store.Store {
	t.Helper()
	st, err := store.FromConfig(cfg)
	if err != nil {
		t.Fatalf("store.FromConfig: %v", err)
	}
	return st
}

func okGateway(t *testing.T, cfg config.Config) (*httptest.Server, func()) {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	if cfg.UpstreamBase == nil {
		u, _ := url.Parse(up.URL)
		cfg.UpstreamBase = u
	}
	if cfg.IncomingPrefix == "" {
		cfg.IncomingPrefix = "/v1"
	}
	gw, err := New(cfg, mustStore(t, cfg), nil, nil)
	if err != nil {
		up.Close()
		t.Fatalf("New: %v", err)
	}
	srv := httptest.NewServer(gw)
	return srv, func() { srv.Close(); up.Close() }
}

func doPost(t *testing.T, base, bearer string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, base+"/v1/chat/completions", nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestAuthDisabledAllows(t *testing.T) {
	srv, cleanup := okGateway(t, config.Config{})
	defer cleanup()
	if code := doPost(t, srv.URL, ""); code != http.StatusOK {
		t.Errorf("auth disabled: got %d, want 200", code)
	}
}

func TestAuthValidKeyAllows(t *testing.T) {
	srv, cleanup := okGateway(t, config.Config{APIKeys: []string{"good"}})
	defer cleanup()
	if code := doPost(t, srv.URL, "good"); code != http.StatusOK {
		t.Errorf("valid key: got %d, want 200", code)
	}
}

func TestAuthInvalidKeyRejected(t *testing.T) {
	srv, cleanup := okGateway(t, config.Config{APIKeys: []string{"good"}})
	defer cleanup()
	if code := doPost(t, srv.URL, "bad"); code != http.StatusUnauthorized {
		t.Errorf("invalid key: got %d, want 401", code)
	}
}

func TestAuthMissingKeyRejected(t *testing.T) {
	srv, cleanup := okGateway(t, config.Config{APIKeys: []string{"good"}})
	defer cleanup()
	if code := doPost(t, srv.URL, ""); code != http.StatusUnauthorized {
		t.Errorf("missing key: got %d, want 401", code)
	}
}

func TestHealthzOpenUnderAuth(t *testing.T) {
	srv, cleanup := okGateway(t, config.Config{APIKeys: []string{"good"}})
	defer cleanup()
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthz under auth: got %d, want 200", resp.StatusCode)
	}
}

func TestManagedModeValidatesThenInjects(t *testing.T) {
	var gotAuth string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	u, _ := url.Parse(up.URL)
	cfg := config.Config{
		UpstreamBase:   u,
		IncomingPrefix: "/v1",
		APIKeys:        []string{"sk-cerberus"},
		UpstreamKey:    "sk-upstream",
	}
	gw, err := New(cfg, mustStore(t, cfg), nil, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	srv := httptest.NewServer(gw)
	defer srv.Close()

	if code := doPost(t, srv.URL, "sk-cerberus"); code != http.StatusOK {
		t.Fatalf("valid managed request: got %d, want 200", code)
	}
	if want := "Bearer sk-upstream"; gotAuth != want {
		t.Errorf("upstream Authorization = %q, want %q", gotAuth, want)
	}
}
