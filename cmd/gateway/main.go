package main

import (
	"log"
	"net/http"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/gateway"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	handler, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("gateway: %v", err)
	}

	keyState := "passthrough (client Authorization forwarded)"
	if cfg.UpstreamKey != "" {
		keyState = "injected from env"
	}
	log.Printf("cerberus listening on %s", cfg.ListenAddr)
	log.Printf("  upstream: %s  (incoming prefix %q, key: %s)", cfg.UpstreamBase, cfg.IncomingPrefix, keyState)

	srv := &http.Server{Addr: cfg.ListenAddr, Handler: handler}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
