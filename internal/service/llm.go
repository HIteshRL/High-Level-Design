// LLM inference service — HTTP client to OpenAI-compatible API.
// Maps to design.swift: Text LLM Inference Node
//
// Uses standard net/http for streaming, no external LLM SDK dependency.
// Compatible with OpenAI, Azure OpenAI, Ollama, LiteLLM proxy, vLLM, etc.
package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prakyathpnayak/roognis/internal/config"
	"github.com/prakyathpnayak/roognis/internal/models"
)

// LLM is the inference client.
type LLM struct {
	cfg    *config.Config
	client *http.Client
}

// NewLLM creates a new LLM inference service.
func NewLLM(cfg *config.Config) *LLM {
	return &LLM{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.LLMTimeoutSeconds) * time.Second,
		},
	}
}

// Complete sends a non-streaming chat completion request.
func (l *LLM) Complete(ctx context.Context, messages []models.LLMMessage, opts ...RequestOption) (*models.LLMResponse, error) {
	o := l.defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	req := models.LLMRequest{
		Model:       o.Model,
		Messages:    messages,
		Temperature: o.Temperature,
		MaxTokens:   o.MaxTokens,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		l.cfg.LLMAPIBase+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if l.cfg.LLMAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+l.cfg.LLMAPIKey)
	}

	start := time.Now()
	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// M3 fix: Limit error response body to 64 KB
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("llm: status %d: %s", resp.StatusCode, string(respBody))
	}

	var llmResp models.LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("llm: decode: %w", err)
	}

	slog.Info("llm.complete",
		"model", llmResp.Model,
		"tokens", llmResp.Usage.TotalTokens,
		"latency_ms", time.Since(start).Milliseconds(),
	)

	return &llmResp, nil
}

// StreamCallback is called for each chunk during streaming.
type StreamCallback func(chunk models.LLMResponse) error

// CompleteStream sends a streaming chat completion request, calling cb for each chunk.
func (l *LLM) CompleteStream(ctx context.Context, messages []models.LLMMessage, cb StreamCallback, opts ...RequestOption) error {
	o := l.defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	req := models.LLMRequest{
		Model:       o.Model,
		Messages:    messages,
		Temperature: o.Temperature,
		MaxTokens:   o.MaxTokens,
		Stream:      true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("llm: marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		l.cfg.LLMAPIBase+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("llm: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if l.cfg.LLMAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+l.cfg.LLMAPIKey)
	}

	// Use a client without global timeout for streaming (context handles cancellation).
	// H7 fix: Set transport-level timeouts to avoid leaked connections.
	streamClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("llm: stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("llm: stream status %d: %s", resp.StatusCode, string(respBody))
	}

	// M8 fix: Increase scanner buffer for long SSE lines
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Up to 1 MB per line
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk models.LLMResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("llm.stream.parse_error", "data", data, "error", err)
			continue
		}

		if err := cb(chunk); err != nil {
			return fmt.Errorf("llm: callback: %w", err)
		}
	}

	return scanner.Err()
}

// RequestOption modifies the default request parameters.
type RequestOption func(*requestOpts)

type requestOpts struct {
	Model       string
	Temperature float64
	MaxTokens   int
}

func (l *LLM) defaultOpts() requestOpts {
	return requestOpts{
		Model:       l.cfg.LLMModel,
		Temperature: l.cfg.LLMTemperature,
		MaxTokens:   l.cfg.LLMMaxTokens,
	}
}

// WithModel overrides the LLM model.
func WithModel(m string) RequestOption {
	return func(o *requestOpts) { o.Model = m }
}

// WithTemperature overrides the temperature.
func WithTemperature(t float64) RequestOption {
	return func(o *requestOpts) { o.Temperature = t }
}

// WithMaxTokens overrides max tokens.
func WithMaxTokens(n int) RequestOption {
	return func(o *requestOpts) { o.MaxTokens = n }
}
