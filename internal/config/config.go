package config

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	ListenAddr     string
	UpstreamBase   *url.URL
	IncomingPrefix string
	UpstreamKey string
	APIKeys []string
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

	return Config{
		ListenAddr:     getenv("CERBERUS_LISTEN", ":8080"),
		UpstreamBase:   u,
		IncomingPrefix: getenv("CERBERUS_INCOMING_PREFIX", "/v1"),
		UpstreamKey:    cmp.Or(os.Getenv("CERBERUS_UPSTREAM_KEY"), os.Getenv("OPENROUTER_API_KEY")),
		APIKeys:        splitKeys(os.Getenv("CERBERUS_API_KEYS")),
	}, nil
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
