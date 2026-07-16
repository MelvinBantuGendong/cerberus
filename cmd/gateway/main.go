package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/gateway"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	handler, err := gateway.New(cfg)
	if err != nil {
		slog.Error("gateway", "err", err)
		os.Exit(1)
	}

	authState := fmt.Sprintf("%d key(s)", len(cfg.APIKeys))
	if len(cfg.APIKeys) == 0 {
		authState = "disabled"
		slog.Warn("auth disabled: all requests accepted, set CERBERUS_API_KEYS to require a Cerberus key")
	}
	if len(cfg.APIKeys) > 0 && cfg.UpstreamKey == "" {
		slog.Warn("CERBERUS_API_KEYS is set but no upstream key: validated client tokens are forwarded upstream unchanged and will fail, set CERBERUS_UPSTREAM_KEY or OPENROUTER_API_KEY")
	}

	keyState := "passthrough (client Authorization forwarded)"
	if cfg.UpstreamKey != "" {
		keyState = "injected from env"
	}
	slog.Info("cerberus starting",
		"listen", cfg.ListenAddr,
		"upstream", cfg.UpstreamBase.String(),
		"incoming_prefix", cfg.IncomingPrefix,
		"upstream_key", keyState,
		"auth", authState,
	)

	srv := &http.Server{Addr: cfg.ListenAddr, Handler: handler}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
