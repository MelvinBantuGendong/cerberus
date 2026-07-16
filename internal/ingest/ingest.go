package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

const (
	maxDecodeDepth = 3
	maxSegments    = 256
)

type Segment struct {
	Text  string
	Role  string
	Trust verdict.TrustLevel
	Path  string
	Depth int
}

type request struct {
	Messages []message `json:"messages"`
}

type message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	Name       string          `json:"name"`
	ToolCalls  []toolCall      `json:"tool_calls"`
	ToolCallID string          `json:"tool_call_id"`
}

type toolCall struct {
	Function functionCall `json:"function"`
}

type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type contentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func Ingest(body []byte) ([]Segment, error) {
	req, err := Parse(body)
	if err != nil {
		return nil, err
	}
	var out []Segment
	for _, s := range Extract(req) {
		expand(s, &out)
	}
	return out, nil
}

func Parse(body []byte) (*request, error) {
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}
	if len(req.Messages) == 0 {
		return nil, errors.New("request has no messages[]")
	}
	return &req, nil
}

func Extract(req *request) []Segment {
	lastUser := -1
	for i, m := range req.Messages {
		if m.Role == "user" {
			lastUser = i
		}
	}

	var segs []Segment
	for i, m := range req.Messages {
		trust := classify(m.Role, i == lastUser)
		base := fmt.Sprintf("messages[%d]", i)

		texts := contentTexts(m.Content)
		for j, t := range texts {
			path := base + ".content"
			if len(texts) > 1 {
				path = fmt.Sprintf("%s.content[%d]", base, j)
			}
			segs = appendText(segs, t, m.Role, path, trust)
		}
		for k, tc := range m.ToolCalls {
			path := fmt.Sprintf("%s.tool_calls[%d].arguments", base, k)
			segs = appendText(segs, tc.Function.Arguments, m.Role, path, trust)
		}
	}
	return segs
}

func appendText(segs []Segment, text, role, path string, trust verdict.TrustLevel) []Segment {
	if strings.TrimSpace(text) == "" {
		return segs
	}
	return append(segs, Segment{Text: text, Role: role, Trust: trust, Path: path})
}

func contentTexts(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return []string{s}
	}
	var parts []contentPart
	if err := json.Unmarshal(raw, &parts); err == nil {
		var out []string
		for _, p := range parts {
			if p.Type == "text" && p.Text != "" {
				out = append(out, p.Text)
			}
		}
		return out
	}
	return nil
}

func classify(role string, isLastUser bool) verdict.TrustLevel {
	switch role {
	case "system":
		return verdict.Trusted
	case "tool":
		return verdict.Untrusted
	case "user":
		if isLastUser {
			return verdict.SemiTrusted
		}
	}
	return verdict.Default
}

func expand(s Segment, out *[]Segment) {
	if len(*out) >= maxSegments || s.Depth > maxDecodeDepth {
		return
	}
	s.Text = Normalize(s.Text)
	if s.Text == "" {
		return
	}
	*out = append(*out, s)
	for _, child := range decodeChildren(s) {
		expand(child, out)
	}
}
