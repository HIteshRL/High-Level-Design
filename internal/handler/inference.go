// Inference handler — text completion (streaming & non-streaming).
// Maps to design.swift: Request Router → Prompt Orchestrator → LLM → Token Streamer.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"unicode/utf8"

	"github.com/google/uuid"
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

	if err := validatePrompt(req.Prompt); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
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
		if errors.Is(err, service.ErrConversationForbidden) {
			writeError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrConversationNotFound) {
			writeError(w, "conversation not found", http.StatusNotFound)
			return
		}
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
		if errors.Is(streamErr, service.ErrConversationForbidden) {
			sse.WriteError("forbidden")
			sse.WriteDone()
			return
		}
		if errors.Is(streamErr, service.ErrConversationNotFound) {
			sse.WriteError("conversation not found")
			sse.WriteDone()
			return
		}
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

	conversations, err := h.orchestrator.ListConversations(r.Context(), user.ID)
	if err != nil {
		slog.Error("inference.list_conversations_error", "error", err, "user_id", user.ID)
		writeError(w, "failed to list conversations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, conversations)
}

// ConversationMessages handles GET /api/v1/conversations/{id}/messages.
func (h *InferenceHandler) ConversationMessages(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conversationIDRaw := r.PathValue("id")
	if conversationIDRaw == "" {
		writeError(w, "conversation id is required", http.StatusBadRequest)
		return
	}

	conversationID, err := uuid.Parse(conversationIDRaw)
	if err != nil {
		writeError(w, "invalid conversation id", http.StatusBadRequest)
		return
	}

	messages, err := h.orchestrator.ListConversationMessages(r.Context(), conversationID, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrConversationForbidden) {
			writeError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrConversationNotFound) {
			writeError(w, "conversation not found", http.StatusNotFound)
			return
		}
		slog.Error("inference.list_conversation_messages_error", "error", err, "user_id", user.ID, "conversation_id", conversationID)
		writeError(w, "failed to list conversation messages", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

// ConversationMessagesByQuery handles GET /api/v1/conversation-messages?conversation_id=<uuid>.
// This provides a wildcard-free compatibility endpoint for clients.
func (h *InferenceHandler) ConversationMessagesByQuery(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conversationIDRaw := r.URL.Query().Get("conversation_id")
	if conversationIDRaw == "" {
		writeError(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	conversationID, err := uuid.Parse(conversationIDRaw)
	if err != nil {
		writeError(w, "invalid conversation id", http.StatusBadRequest)
		return
	}

	messages, err := h.orchestrator.ListConversationMessages(r.Context(), conversationID, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrConversationForbidden) {
			writeError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrConversationNotFound) {
			writeError(w, "conversation not found", http.StatusNotFound)
			return
		}
		slog.Error("inference.list_conversation_messages_query_error", "error", err, "user_id", user.ID, "conversation_id", conversationID)
		writeError(w, "failed to list conversation messages", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func validatePrompt(prompt string) error {
	if utf8.RuneCountInString(prompt) > 32000 {
		return errors.New("prompt exceeds maximum length of 32000 characters")
	}
	return nil
}
