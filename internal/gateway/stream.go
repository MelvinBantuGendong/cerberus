package gateway

import (
	"bufio"
	"io"
	"strings"

	"github.com/MelvinBantuGendong/cerberus/internal/ingest"
	"github.com/MelvinBantuGendong/cerberus/internal/verdict"
)

const streamWindow = 512

type pendingFrame struct {
	raw string
	end int
}

func newStreamScanner(upstream io.ReadCloser, detectors []ingest.Detector, record func(verdict.Verdict)) io.ReadCloser {
	pr, pw := io.Pipe()
	go runStreamScan(upstream, pw, detectors, record)
	return pr
}

func runStreamScan(upstream io.ReadCloser, pw *io.PipeWriter, detectors []ingest.Detector, record func(verdict.Verdict)) {
	defer upstream.Close()

	var (
		text    strings.Builder
		pending []pendingFrame
		lines   []string
	)

	scan := func() verdict.Verdict {
		return ingest.Dispatch(verdict.Outbound, ingest.FromText(text.String()), detectors)
	}
	refuse := func(v verdict.Verdict) {
		record(v)
		_, _ = pw.Write(buildRefusal(true))
		pw.Close()
	}
	release := func(boundary int) error {
		for len(pending) > 0 && pending[0].end <= boundary {
			if _, err := io.WriteString(pw, pending[0].raw); err != nil {
				return err
			}
			pending = pending[1:]
		}
		return nil
	}

	flushFrame := func() (bool, error) {
		if len(lines) == 0 {
			return false, nil
		}
		raw := strings.Join(lines, "\n") + "\n\n"
		text.WriteString(frameContent(lines))
		lines = lines[:0]
		pending = append(pending, pendingFrame{raw: raw, end: text.Len()})

		if v := scan(); v.Action == verdict.Block {
			refuse(v)
			return true, nil
		}
		return false, release(text.Len() - streamWindow)
	}

	sc := bufio.NewScanner(upstream)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if line := sc.Text(); line != "" {
			lines = append(lines, line)
			continue
		}
		if blocked, err := flushFrame(); blocked {
			return
		} else if err != nil {
			pw.CloseWithError(err)
			return
		}
	}
	if blocked, err := flushFrame(); blocked {
		return
	} else if err != nil {
		pw.CloseWithError(err)
		return
	}
	if err := sc.Err(); err != nil {
		pw.CloseWithError(err)
		return
	}

	if v := scan(); v.Action == verdict.Block {
		refuse(v)
		return
	} else {
		record(v)
	}
	for _, f := range pending {
		if _, err := io.WriteString(pw, f.raw); err != nil {
			pw.CloseWithError(err)
			return
		}
	}
	pw.Close()
}

func frameContent(lines []string) string {
	var b strings.Builder
	for _, ln := range lines {
		if !strings.HasPrefix(ln, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(ln, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var c sseChunk
		if unmarshalChunk(data, &c) {
			for _, ch := range c.Choices {
				b.WriteString(ch.Delta.Content)
			}
		}
	}
	return b.String()
}
