package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
)

const adminTok = "admin-secret"

func adminGateway(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	u, _ := url.Parse(up.URL)
	cfg := config.Config{UpstreamBase: u, IncomingPrefix: "/v1", AdminToken: adminTok}
	gw, err := New(cfg, mustStore(t, cfg), nil, nil)
	if err != nil {
		up.Close()
		t.Fatalf("New: %v", err)
	}
	srv := httptest.NewServer(gw)
	return srv, func() { srv.Close(); up.Close() }
}

func adminReq(t *testing.T, method, url, token, body string) *http.Response {
	t.Helper()
	var r *strings.Reader = strings.NewReader(body)
	req, _ := http.NewRequest(method, url, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func TestAdminRequiresToken(t *testing.T) {
	srv, cleanup := adminGateway(t)
	defer cleanup()

	resp := adminReq(t, "GET", srv.URL+"/admin/config", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no token: got %d, want 401", resp.StatusCode)
	}
	resp = adminReq(t, "GET", srv.URL+"/admin/config", "wrong", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong token: got %d, want 401", resp.StatusCode)
	}
	resp = adminReq(t, "GET", srv.URL+"/admin/config", adminTok, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid token: got %d, want 200", resp.StatusCode)
	}
}

func TestAdminKeyIssuanceTakesEffect(t *testing.T) {
	srv, cleanup := adminGateway(t)
	defer cleanup()

	if code := doPost(t, srv.URL, ""); code != http.StatusOK {
		t.Fatalf("pre-key: got %d, want 200", code)
	}

	resp := adminReq(t, "POST", srv.URL+"/admin/keys", adminTok, `{"label":"ci"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create key: got %d, want 201", resp.StatusCode)
	}
	var out struct{ ID, Key string }
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if out.Key == "" || out.ID == "" {
		t.Fatal("create key returned no key/id")
	}

	if code := doPost(t, srv.URL, ""); code != http.StatusUnauthorized {
		t.Errorf("no-key after issuance: got %d, want 401", code)
	}
	if code := doPost(t, srv.URL, out.Key); code != http.StatusOK {
		t.Errorf("issued key: got %d, want 200", code)
	}

	del := adminReq(t, "DELETE", srv.URL+"/admin/keys/"+out.ID, adminTok, "")
	del.Body.Close()
	if del.StatusCode != http.StatusNoContent {
		t.Errorf("revoke: got %d, want 204", del.StatusCode)
	}
	if code := doPost(t, srv.URL, ""); code != http.StatusOK {
		t.Errorf("after revoke: got %d, want 200", code)
	}
}

func TestAdminConfigUpdate(t *testing.T) {
	srv, cleanup := adminGateway(t)
	defer cleanup()

	bad := adminReq(t, "PUT", srv.URL+"/admin/config", adminTok, `{"upstream":"not-a-url"}`)
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid upstream: got %d, want 400", bad.StatusCode)
	}

	ok := adminReq(t, "PUT", srv.URL+"/admin/config", adminTok, `{"max_body_bytes":123456,"outbound_mode":"off"}`)
	ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("valid update: got %d, want 200", ok.StatusCode)
	}

	resp := adminReq(t, "GET", srv.URL+"/admin/config", adminTok, "")
	defer resp.Body.Close()
	var v struct {
		MaxBodyBytes int64  `json:"max_body_bytes"`
		OutboundMode string `json:"outbound_mode"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&v)
	if v.MaxBodyBytes != 123456 || v.OutboundMode != "off" {
		t.Errorf("config after update = %+v, want max=123456 mode=off", v)
	}
}
