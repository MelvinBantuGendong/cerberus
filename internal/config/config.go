package config

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type OutboundMode string

const (
	OutboundOff    OutboundMode = "off"
	OutboundBuffer OutboundMode = "buffer"
	OutboundStream OutboundMode = "stream"
)

type Config struct {
	ListenAddr     string
	UpstreamBase   *url.URL
	IncomingPrefix string
	UpstreamKey    string
	APIKeys        []string
	MaxBodyBytes   int64
	OutboundMode   OutboundMode

	AdminToken string

	StatePath string
}

func Load() (Config, error) {
	rawUpstream := getenv("CERBERUS_UPSTREAM", "https://openrouter.ai/api/v1")
	u, err := url.Parse(rawUpstream)
	if err != nil {
		return Config{}, fmt.Errorf("CERBERUS_UPSTREAM %q is not a valid URL: %w", rawUpstream, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("CERBERUS_UPSTREAM %q must be an absolute URL (scheme + host)", rawUpstream)
	}

	mode, err := parseOutboundMode(os.Getenv("CERBERUS_OUTBOUND_MODE"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		ListenAddr:     getenv("CERBERUS_LISTEN", ":8080"),
		UpstreamBase:   u,
		IncomingPrefix: getenv("CERBERUS_INCOMING_PREFIX", "/v1"),
		UpstreamKey:    cmp.Or(os.Getenv("CERBERUS_UPSTREAM_KEY"), os.Getenv("OPENROUTER_API_KEY")),
		APIKeys:        splitKeys(os.Getenv("CERBERUS_API_KEYS")),
		MaxBodyBytes:   parseSize(os.Getenv("CERBERUS_MAX_BODY_BYTES"), 4<<20),
		OutboundMode:   mode,
		AdminToken:     os.Getenv("CERBERUS_ADMIN_TOKEN"),
		StatePath:      os.Getenv("CERBERUS_STATE_PATH"),
	}, nil
}

func parseOutboundMode(raw string) (OutboundMode, error) {
	switch OutboundMode(raw) {
	case "":
		return OutboundBuffer, nil
	case OutboundOff, OutboundBuffer, OutboundStream:
		return OutboundMode(raw), nil
	default:
		return "", fmt.Errorf("CERBERUS_OUTBOUND_MODE %q must be one of: off, buffer, stream", raw)
	}
}

func parseSize(raw string, fallback int64) int64 {
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
		return n
	}
	return fallback
}

func splitKeys(raw string) []string {
	var keys []string
	for _, k := range strings.Split(raw, ",") {
		if k = strings.TrimSpace(k); k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
