package gateway

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
)

func New(cfg config.Config) (http.Handler, error) {
	proxy := &httputil.ReverseProxy{
		FlushInterval: -1,

		Rewrite: func(pr *httputil.ProxyRequest) {
			rest := strings.TrimPrefix(pr.In.URL.Path, cfg.IncomingPrefix)
			pr.Out.URL.Scheme = cfg.UpstreamBase.Scheme
			pr.Out.URL.Host = cfg.UpstreamBase.Host
			pr.Out.URL.Path = slashParseFix(cfg.UpstreamBase.Path, rest)
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
			pr.Out.Host = cfg.UpstreamBase.Host

			if cfg.UpstreamKey != "" {
				pr.Out.Header.Set("Authorization", "Bearer "+cfg.UpstreamKey)
			}
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("upstream request failed", "method", r.Method, "path", r.URL.Path, "err", err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("/", withAudit(withAuth(cfg, proxy)))

	return mux, nil
}

func withAuth(cfg config.Config, next http.Handler) http.Handler {
	if len(cfg.APIKeys) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" || !validKey(token, cfg.APIKeys) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return header[len(prefix):]
	}
	return ""
}

func validKey(token string, keys []string) bool {
	var ok bool
	for _, k := range keys {
		if subtle.ConstantTimeCompare([]byte(token), []byte(k)) == 1 {
			ok = true
		}
	}
	return ok
}

func withAudit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.wroteHeader = true
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func slashParseFix(a, b string) string {
	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash && a != "" && b != "":
		return a + "/" + b
	}
	return a + b
}
