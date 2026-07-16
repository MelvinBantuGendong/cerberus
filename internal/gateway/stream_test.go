package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/MelvinBantuGendong/cerberus/internal/config"
	"github.com/MelvinBantuGendong/cerberus/internal/detect"
)

func sseChunks(chunks ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl, _ := w.(http.Flusher)
		for _, c := range chunks {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%s}}]}\n\n", jsonQuote(c))
			if fl != nil {
				fl.Flush()
			}
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		if fl != nil {
			fl.Flush()
		}
	}
}

const secret = "sk-abcdefghijklmnopqrstuvwxyz012345"

func TestStreamAllowsBenign(t *testing.T) {
	srv, cleanup := outboundGateway(t, config.OutboundStream, sseChunks("hello ", "there ", "friend"), detect.Outbound)
	defer cleanup()

	_, body := postBody(t, srv.URL, userMsg)
	for _, w := range []string{"hello", "there", "friend", "[DONE]"} {
		if !strings.Contains(body, w) {
			t.Errorf("benign stream missing %q in output: %s", w, body)
		}
	}
}

func TestStreamBlocksSecretBeforeRelease(t *testing.T) {
	srv, cleanup := outboundGateway(t, config.OutboundStream, sseChunks("your key is ", secret, " keep it safe"), detect.Outbound)
	defer cleanup()

	_, body := postBody(t, srv.URL, userMsg)
	if strings.Contains(body, secret) {
		t.Errorf("secret reached the client: %s", body)
	}
	if !strings.Contains(body, refusalText) {
		t.Errorf("expected a refusal in the stream: %s", body)
	}
}

func TestStreamBlocksBoundarySpanningSecret(t *testing.T) {
	srv, cleanup := outboundGateway(t, config.OutboundStream,
		sseChunks("token sk-abcdefghij", "klmnopqrstuvwxyz012345 end"), detect.Outbound)
	defer cleanup()

	_, body := postBody(t, srv.URL, userMsg)
	if strings.Contains(body, secret) {
		t.Errorf("boundary-spanning secret leaked: %s", body)
	}
	if !strings.Contains(body, refusalText) {
		t.Errorf("expected a refusal: %s", body)
	}
}

func TestStreamReleasesSafePrefixThenCuts(t *testing.T) {
	var chunks []string
	for i := 0; i < 300; i++ {
		chunks = append(chunks, "word ")
	}
	chunks = append(chunks, "then the "+secret+" leaks")
	srv, cleanup := outboundGateway(t, config.OutboundStream, sseChunks(chunks...), detect.Outbound)
	defer cleanup()

	_, body := postBody(t, srv.URL, userMsg)

	if n := strings.Count(body, "word"); n < 100 {
		t.Errorf("safe prefix was not streamed through: only %d word frames released", n)
	}
	if strings.Contains(body, secret) {
		t.Errorf("secret after the safe prefix leaked: %s", body)
	}
	if !strings.Contains(body, refusalText) {
		t.Errorf("expected a refusal after the safe prefix: %.120s", body)
	}
}

func TestStreamNonSSEFallsBackToBuffer(t *testing.T) {
	srv, cleanup := outboundGateway(t, config.OutboundStream, jsonUpstream("here is "+secret), detect.Outbound)
	defer cleanup()

	resp, body := postBody(t, srv.URL, userMsg)
	if resp.Header.Get("X-Cerberus-Outbound") != "block" {
		t.Errorf("X-Cerberus-Outbound = %q, want block (buffer fallback)", resp.Header.Get("X-Cerberus-Outbound"))
	}
	if strings.Contains(body, secret) {
		t.Errorf("secret leaked in JSON fallback: %s", body)
	}
}
