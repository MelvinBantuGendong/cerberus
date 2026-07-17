package gateway

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/store"
)

type DetectorInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Direction   string `json:"direction"`
}

func adminHandler(token string, st *store.Store, catalog []DetectorInfo) http.Handler {
	m := http.NewServeMux()

	m.HandleFunc("GET /admin/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, st.View())
	})

	m.HandleFunc("PUT /admin/config", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Upstream      *string   `json:"upstream"`
			UpstreamKey   *string   `json:"upstream_key"`
			MaxBodyBytes  *int64    `json:"max_body_bytes"`
			OutboundMode  *string   `json:"outbound_mode"`
			DisabledRules *[]string `json:"disabled_rules"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		patch := store.SettingsPatch{
			Upstream:      p.Upstream,
			UpstreamKey:   p.UpstreamKey,
			MaxBodyBytes:  p.MaxBodyBytes,
			DisabledRules: p.DisabledRules,
		}
		if p.OutboundMode != nil {
			mode := config.OutboundMode(*p.OutboundMode)
			patch.OutboundMode = &mode
		}
		if err := st.UpdateSettings(patch); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, st.View())
	})

	m.HandleFunc("GET /admin/detectors", func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			catalog = []DetectorInfo{}
		}
		writeJSON(w, http.StatusOK, catalog)
	})

	m.HandleFunc("GET /admin/keys", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, st.ListKeys())
	})

	m.HandleFunc("POST /admin/keys", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Label string `json:"label"`
		}
		_ = json.NewDecoder(r.Body).Decode(&p)
		token, k, err := st.GenerateKey(p.Label)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"id": k.ID, "label": k.Label, "key": token, "created": k.Created,
		})
	})

	m.HandleFunc("DELETE /admin/keys/{id}", func(w http.ResponseWriter, r *http.Request) {
		ok, err := st.RevokeKey(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "key not found", http.StatusNotFound)
	})

	return adminAuth(token, m)
}

func adminAuth(token string, next http.Handler) http.Handler {
	want := []byte(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := bearerToken(r.Header.Get("Authorization"))
		if got == "" || subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
