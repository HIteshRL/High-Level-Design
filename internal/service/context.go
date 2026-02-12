// Context injector — stub for RAG pipeline integration.
// Maps to design.swift: Context Injector (RAG Context) + Context Assembler
//
// In the walking skeleton, this returns an empty context.
// When the RAG pipeline (ETL → Embedding → VectorDB → Retriever → Re-Ranker)
// is built, this service will call the Retriever and Re-Ranker, then assemble
// the context window with a hard token budget.
package service

import (
	"context"
	"log/slog"

	"github.com/prakyathpnayak/roognis/internal/models"
)

// ContextInjector enriches prompts with RAG context.
type ContextInjector struct {
	// Future: retriever, reranker, vectorDB client, token budget
}

// NewContextInjector creates the context injector stub.
func NewContextInjector() *ContextInjector {
	return &ContextInjector{}
}

// Inject takes the user messages and returns an augmented message list
// with relevant context prepended as system messages.
//
// Walking skeleton: pass-through (no RAG context yet).
func (ci *ContextInjector) Inject(ctx context.Context, messages []models.LLMMessage) []models.LLMMessage {
	slog.Debug("context_injector.inject", "msg_count", len(messages), "rag", "stub")

	// Future implementation:
	// 1. Extract the last user message
	// 2. Generate embedding via Embedding Generator
	// 3. Query Vector Database (pgvector) for top-k similar chunks
	// 4. Re-rank results
	// 5. Assemble context with token budget enforcement
	// 6. Prepend as system message(s)

	// For now, add a minimal system prompt
	systemMsg := models.LLMMessage{
		Role:    "system",
		Content: "You are a helpful educational AI assistant. Provide clear, accurate, and pedagogically sound responses.",
	}

	// Check if system message already exists
	if len(messages) > 0 && messages[0].Role == "system" {
		return messages
	}

	return append([]models.LLMMessage{systemMsg}, messages...)
}
