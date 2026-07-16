package gateway

import (
	"log"
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
			log.Printf("gateway: upstream error for %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("/", proxy)

	return withAccessLog(mux), nil
}

func withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
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
