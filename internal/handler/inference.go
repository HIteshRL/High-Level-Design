// Inference handler — text completion (streaming & non-streaming).
// Maps to design.swift: Request Router → Prompt Orchestrator → LLM → Token Streamer.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/prakyathpnayak/roognis/internal/middleware"
	"github.com/prakyathpnayak/roognis/internal/models"
	"github.com/prakyathpnayak/roognis/internal/service"
)

// InferenceHandler handles POST /api/v1/inference/complete.
type InferenceHandler struct {
	orchestrator *service.Orchestrator
}

// NewInferenceHandler creates a new inference handler.
func NewInferenceHandler(orch *service.Orchestrator) *InferenceHandler {
	return &InferenceHandler{orchestrator: orch}
}

// Complete handles POST /api/v1/inference/complete.
// Dispatches to streaming (SSE) or non-streaming based on the request body.
func (h *InferenceHandler) Complete(w http.ResponseWriter, r *http.Request) {
	// H6 fix: Limit request body size (64 KB for inference requests)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.InferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		writeError(w, "prompt is required", http.StatusBadRequest)
		return
	}

	// L9 fix: Enforce prompt length limit (32K chars)
	if len(req.Prompt) > 32000 {
		writeError(w, "prompt exceeds maximum length of 32000 characters", http.StatusBadRequest)
		return
	}

	if req.Stream {
		h.handleStream(w, r, &req, user)
	} else {
		h.handleComplete(w, r, &req, user)
	}
}

// handleComplete processes a non-streaming inference request.
func (h *InferenceHandler) handleComplete(w http.ResponseWriter, r *http.Request, req *models.InferenceRequest, user *models.User) {
	resp, err := h.orchestrator.Complete(r.Context(), req, user.ID)
	if err != nil {
		slog.Error("inference.complete_error", "error", err, "user_id", user.ID)
		writeError(w, "inference failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleStream processes a streaming inference request via SSE.
func (h *InferenceHandler) handleStream(w http.ResponseWriter, r *http.Request, req *models.InferenceRequest, user *models.User) {
	sse, err := service.NewSSEWriter(w)
	if err != nil {
		writeError(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	streamErr := h.orchestrator.StreamComplete(r.Context(), req, user.ID, func(chunk models.LLMResponse) error {
		if len(chunk.Choices) == 0 {
			return nil
		}

		sc := models.StreamChunk{
			Delta:        chunk.Choices[0].Delta.Content,
			FinishReason: chunk.Choices[0].FinishReason,
			Model:        chunk.Model,
		}
		return sse.WriteData(sc)
	})

	if streamErr != nil {
		slog.Error("inference.stream_error", "error", streamErr, "user_id", user.ID)
		sse.WriteError(streamErr.Error())
	}

	sse.WriteDone()
}

// Conversations handles GET /api/v1/conversations.
func (h *InferenceHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// The orchestrator holds a reference to the pool; we'll access it via a lightweight query.
	// For the walking skeleton, we reuse the pool through the orchestrator's exposed method.
	// This is a pragmatic shortcut — a proper Conversations handler would have its own service.
	writeJSON(w, http.StatusOK, map[string]string{"message": "conversations list - implement with dedicated service"})
}
