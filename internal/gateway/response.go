package gateway

import (
	"encoding/json"
	"strings"
)

const refusalText = "I'm sorry, but I can't provide that response."

func extractResponseText(body []byte, isSSE bool) string {
	if isSSE {
		return extractSSEText(body)
	}
	return extractJSONText(body)
}

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func extractSSEText(body []byte) string {
	var b strings.Builder
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var c sseChunk
		if !unmarshalChunk(data, &c) {
			continue
		}
		for _, ch := range c.Choices {
			b.WriteString(ch.Delta.Content)
		}
	}
	return b.String()
}

func unmarshalChunk(data string, c *sseChunk) bool {
	return json.Unmarshal([]byte(data), c) == nil
}

type jsonResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func extractJSONText(body []byte) string {
	var r jsonResponse
	if json.Unmarshal(body, &r) != nil {
		return ""
	}
	var b strings.Builder
	for _, ch := range r.Choices {
		b.WriteString(ch.Message.Content)
	}
	return b.String()
}

func buildRefusal(isSSE bool) []byte {
	content := jsonQuote(refusalText)
	if isSSE {
		chunk := `{"id":"cerberus-block","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":` + content + `},"finish_reason":null}]}`
		stop := `{"id":"cerberus-block","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`
		return []byte("data: " + chunk + "\n\ndata: " + stop + "\n\ndata: [DONE]\n\n")
	}
	return []byte(`{"id":"cerberus-block","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":` + content + `},"finish_reason":"stop"}]}`)
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
