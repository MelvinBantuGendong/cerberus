package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/detect"
	"github.com/MelvinBantuGendong/cerberus/internal/gateway"
	"github.com/MelvinBantuGendong/cerberus/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	st, err := store.FromConfig(cfg)
	if err != nil {
		slog.Error("store", "err", err)
		os.Exit(1)
	}

	catalog := make([]gateway.DetectorInfo, 0, len(detect.Catalog()))
	for _, d := range detect.Catalog() {
		catalog = append(catalog, gateway.DetectorInfo{ID: d.ID, Name: d.Name, Description: d.Description, Direction: d.Direction})
	}

	handler, err := gateway.New(cfg, st, detect.Default(), detect.Outbound, catalog)
	if err != nil {
		slog.Error("gateway", "err", err)
		os.Exit(1)
	}

	view := st.View()
	authState := fmt.Sprintf("%d key(s)", view.KeyCount)
	if view.KeyCount == 0 {
		authState = "disabled"
		slog.Warn("auth disabled: all requests accepted, add a Cerberus key (CERBERUS_API_KEYS or POST /admin/keys)")
	}
	if view.KeyCount > 0 && !view.UpstreamKeySet {
		slog.Warn("keys are set but no upstream key: validated client tokens are forwarded upstream unchanged and will fail")
	}

	adminState := "disabled (set CERBERUS_ADMIN_TOKEN to enable)"
	if cfg.AdminToken != "" {
		adminState = "enabled"
		slog.Warn("admin API enabled: /admin/* is protected only by CERBERUS_ADMIN_TOKEN; do not expose it publicly")
		if cfg.StatePath == "" {
			slog.Warn("no CERBERUS_STATE_PATH: admin changes are in-memory and lost on restart")
		}
	}

	keyState := "passthrough (client Authorization forwarded)"
	if view.UpstreamKeySet {
		keyState = "injected"
	}
	slog.Info("cerberus starting",
		"listen", cfg.ListenAddr,
		"upstream", view.Upstream,
		"incoming_prefix", cfg.IncomingPrefix,
		"upstream_key", keyState,
		"auth", authState,
		"max_body_bytes", view.MaxBodyBytes,
		"outbound_mode", view.OutboundMode,
		"admin", adminState,
	)

	srv := &http.Server{Addr: cfg.ListenAddr, Handler: handler}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
