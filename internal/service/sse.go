// SSE token streamer — writes Server-Sent Events to the HTTP response.
// Maps to design.swift: Token Streamer
package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter writes Server-Sent Events to an http.ResponseWriter.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates an SSE writer and sets up the response headers.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("sse: streaming not supported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return &SSEWriter{w: w, flusher: flusher}, nil
}

// WriteEvent writes a named SSE event with JSON data.
func (s *SSEWriter) WriteEvent(event string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("sse: marshal: %w", err)
	}

	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return fmt.Errorf("sse: write event: %w", err)
		}
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", jsonData); err != nil {
		return fmt.Errorf("sse: write data: %w", err)
	}
	s.flusher.Flush()
	return nil
}

// WriteData writes a data-only SSE event.
func (s *SSEWriter) WriteData(data any) error {
	return s.WriteEvent("", data)
}

// WriteDone writes the [DONE] sentinel and closes the stream.
func (s *SSEWriter) WriteDone() {
	fmt.Fprint(s.w, "data: [DONE]\n\n")
	s.flusher.Flush()
}

// WriteError writes an error event.
func (s *SSEWriter) WriteError(msg string) {
	s.WriteEvent("error", map[string]string{"error": msg})
}
