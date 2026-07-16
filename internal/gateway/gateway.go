package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/store"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

type OutboundFactory func(systemPrompt string) []ingest.Detector

func New(cfg config.Config, st *store.Store, inbound []ingest.Detector, outbound OutboundFactory) (http.Handler, error) {
	hasOutbound := outbound != nil

	proxy := &httputil.ReverseProxy{
		FlushInterval: -1,

		Rewrite: func(pr *httputil.ProxyRequest) {
			snap := st.Snapshot()
			rest := strings.TrimPrefix(pr.In.URL.Path, cfg.IncomingPrefix)
			pr.Out.URL.Scheme = snap.Upstream.Scheme
			pr.Out.URL.Host = snap.Upstream.Host
			pr.Out.URL.Path = slashParseFix(snap.Upstream.Path, rest)
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
			pr.Out.Host = snap.Upstream.Host

			if snap.UpstreamKey != "" {
				pr.Out.Header.Set("Authorization", "Bearer "+snap.UpstreamKey)
			}
			if hasOutbound && snap.OutboundMode != config.OutboundOff {
				pr.Out.Header.Set("Accept-Encoding", "identity")
			}
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("upstream request failed", "method", r.Method, "path", r.URL.Path, "err", err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
		},
	}
	if hasOutbound {
		proxy.ModifyResponse = modifyResponse(st, outbound)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	if cfg.AdminToken != "" {
		mux.Handle("/admin/", adminHandler(cfg.AdminToken, st))
	}
	mux.Handle("/", withAudit(withAuth(st, withScan(st, inbound, hasOutbound, proxy))))

	return mux, nil
}

func withAuth(st *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !st.Snapshot().AuthEnabled {
			next.ServeHTTP(w, r)
			return
		}
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" || !st.ValidKey(token) {
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

func withScan(st *store.Store, inbound []ingest.Detector, hasOutbound bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}

		snap := st.Snapshot()
		r.Body = http.MaxBytesReader(w, r.Body, snap.MaxBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "cannot read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))

		rs := stateFrom(r.Context())
		if rs != nil && hasOutbound && snap.OutboundMode != config.OutboundOff {
			rs.systemPrompt = ingest.SystemPrompt(body)
		}

		if len(body) > 0 && len(inbound) > 0 {
			segs, perr := ingest.Ingest(body)
			if perr != nil {
				slog.Debug("scan skipped: body is not a chat-completions request", "path", r.URL.Path, "err", perr)
			} else {
				v := ingest.Dispatch(verdict.Inbound, segs, inbound)
				if rs != nil {
					rs.inbound = &v
				}
				if v.Action == verdict.Block {
					writeBlocked(w, v)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func modifyResponse(st *store.Store, outbound OutboundFactory) func(*http.Response) error {
	return func(resp *http.Response) error {
		mode := st.Snapshot().OutboundMode
		if mode == config.OutboundOff {
			return nil
		}

		sp := ""
		if rs := stateFrom(resp.Request.Context()); rs != nil {
			sp = rs.systemPrompt
		}
		detectors := outbound(sp)
		record := func(v verdict.Verdict) {
			if rs := stateFrom(resp.Request.Context()); rs != nil {
				rs.outbound = &v
			}
		}
		isSSE := strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")

		if mode == config.OutboundStream && isSSE {
			resp.Header.Del("Content-Length")
			resp.Body = newStreamScanner(resp.Body, detectors, record)
			return nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		v := ingest.Dispatch(verdict.Outbound, ingest.FromText(extractResponseText(body, isSSE)), detectors)
		record(v)

		out := body
		switch v.Action {
		case verdict.Block:
			out = buildRefusal(isSSE)
			resp.StatusCode = http.StatusOK
			resp.Header.Set("X-Cerberus-Outbound", "block")
		case verdict.Flag:
			resp.Header.Set("X-Cerberus-Outbound", "flag")
		}
		resp.Body = io.NopCloser(bytes.NewReader(out))
		resp.ContentLength = int64(len(out))
		resp.Header.Set("Content-Length", strconv.Itoa(len(out)))
		resp.TransferEncoding = nil
		return nil
	}
}

func writeBlocked(w http.ResponseWriter, v verdict.Verdict) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(v)
}

type ctxKey int

const stateKey ctxKey = iota

type reqState struct {
	inbound      *verdict.Verdict
	outbound     *verdict.Verdict
	systemPrompt string
}

func stateFrom(ctx context.Context) *reqState {
	rs, _ := ctx.Value(stateKey).(*reqState)
	return rs
}

func withAudit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rs := &reqState{}
		r = r.WithContext(context.WithValue(r.Context(), stateKey, rs))

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if rs.inbound != nil {
			attrs = append(attrs, verdictAttrs(rs.inbound)...)
		}
		slog.Info("request", attrs...)

		if rs.outbound != nil {
			resp := []any{"path", r.URL.Path}
			resp = append(resp, verdictAttrs(rs.outbound)...)
			slog.Info("response", resp...)
		}
	})
}

func verdictAttrs(v *verdict.Verdict) []any {
	return []any{
		"action", v.Action,
		"score", v.Score,
		"categories", v.Categories,
		"matched_rules", v.MatchedRules,
		"trust_level", v.TrustLevel,
	}
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
