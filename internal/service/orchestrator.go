// Prompt orchestrator — the pipeline conductor.
// Maps to design.swift: Request Router → Prompt Orchestrator → Context Injector
//
//	→ LLM Inference Node → Token Streamer (SSE)
//
// Coordinates: conversation management, message history, cache lookup,
// RAG context injection, LLM invocation, cache write-back, and persistence.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/prakyathpnayak/roognis/internal/db"
	"github.com/prakyathpnayak/roognis/internal/models"
)

// Orchestrator is the inference pipeline conductor.
type Orchestrator struct {
	llm    *LLM
	cache  *Cache
	ctxInj *ContextInjector
	pool   *db.Pool
}

// NewOrchestrator creates a new orchestrator wiring together the pipeline stages.
func NewOrchestrator(llm *LLM, cache *Cache, ctxInj *ContextInjector, pool *db.Pool) *Orchestrator {
	return &Orchestrator{
		llm:    llm,
		cache:  cache,
		ctxInj: ctxInj,
		pool:   pool,
	}
}

// Complete runs the full non-streaming inference pipeline.
func (o *Orchestrator) Complete(ctx context.Context, req *models.InferenceRequest, userID uuid.UUID) (*models.InferenceResponse, error) {
	start := time.Now()

	// 1. Resolve or create conversation
	conversationID, err := o.resolveConversation(ctx, req.ConversationID, userID)
	if err != nil {
		return nil, err
	}

	// 2. Build message history
	messages, err := o.buildMessages(ctx, conversationID, req.Prompt)
	if err != nil {
		return nil, err
	}

	// 3. Check cache
	model := req.Model
	if model == "" {
		model = o.llm.cfg.LLMModel
	}
	temp := o.llm.cfg.LLMTemperature
	if req.Temperature != nil {
		temp = *req.Temperature
	}
	maxTok := o.llm.cfg.LLMMaxTokens
	if req.MaxTokens != nil {
		maxTok = *req.MaxTokens
	}
	cacheKey := SemanticHash(req.Prompt, model, userID.String(), temp, maxTok)

	var cachedResp models.InferenceResponse
	if found, _ := o.cache.GetJSON(ctx, cacheKey, &cachedResp); found {
		slog.Info("orchestrator.cache_hit", "conversation_id", conversationID)
		cachedResp.Cached = true
		cachedResp.ConversationID = conversationID
		cachedResp.LatencyMs = float64(time.Since(start).Milliseconds())
		return &cachedResp, nil
	}

	// 4. Inject RAG context
	messages = o.ctxInj.Inject(ctx, messages)

	// 5. Call LLM
	var opts []RequestOption
	if req.Model != "" {
		opts = append(opts, WithModel(req.Model))
	}
	if req.Temperature != nil {
		opts = append(opts, WithTemperature(*req.Temperature))
	}
	if req.MaxTokens != nil {
		opts = append(opts, WithMaxTokens(*req.MaxTokens))
	}

	llmResp, err := o.llm.Complete(ctx, messages, opts...)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: llm: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("orchestrator: llm returned no choices")
	}

	content := llmResp.Choices[0].Message.Content
	latencyMs := float64(time.Since(start).Milliseconds())
	totalTokens := llmResp.Usage.TotalTokens

	// 6. Build response
	resp := &models.InferenceResponse{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Content:        content,
		Model:          llmResp.Model,
		TokenCount:     &totalTokens,
		LatencyMs:      latencyMs,
		Cached:         false,
	}

	// 7. Cache response
	if err := o.cache.SetJSON(ctx, cacheKey, resp); err != nil {
		slog.Warn("orchestrator.cache_set_error", "error", err)
	}

	// 8. Persist user + assistant messages
	o.persistMessages(ctx, conversationID, req.Prompt, content, llmResp.Model, totalTokens, latencyMs)

	return resp, nil
}

// StreamComplete runs the streaming inference pipeline.
func (o *Orchestrator) StreamComplete(ctx context.Context, req *models.InferenceRequest, userID uuid.UUID, cb StreamCallback) error {
	// 1. Resolve or create conversation
	conversationID, err := o.resolveConversation(ctx, req.ConversationID, userID)
	if err != nil {
		return err
	}

	// 2. Build message history
	messages, err := o.buildMessages(ctx, conversationID, req.Prompt)
	if err != nil {
		return err
	}

	// 3. Inject RAG context
	messages = o.ctxInj.Inject(ctx, messages)

	// 4. Prepare options
	var opts []RequestOption
	if req.Model != "" {
		opts = append(opts, WithModel(req.Model))
	}
	if req.Temperature != nil {
		opts = append(opts, WithTemperature(*req.Temperature))
	}
	if req.MaxTokens != nil {
		opts = append(opts, WithMaxTokens(*req.MaxTokens))
	}

	// 5. Stream from LLM, forwarding chunks to caller
	var fullContent string
	err = o.llm.CompleteStream(ctx, messages, func(chunk models.LLMResponse) error {
		if len(chunk.Choices) > 0 {
			fullContent += chunk.Choices[0].Delta.Content
		}
		return cb(chunk)
	}, opts...)
	if err != nil {
		return fmt.Errorf("orchestrator: stream: %w", err)
	}

	// 6. Persist after stream completes
	o.persistMessages(ctx, conversationID, req.Prompt, fullContent, o.llm.cfg.LLMModel, 0, 0)

	return nil
}

// resolveConversation returns an existing conversation ID or creates a new one.
func (o *Orchestrator) resolveConversation(ctx context.Context, convID *uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	if convID != nil {
		conv, err := o.pool.GetConversation(ctx, *convID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("orchestrator: get conversation: %w", err)
		}
		if conv != nil {
			// C3 fix: Explicit ownership check — never silently fall through
			if conv.UserID != userID {
				return uuid.Nil, fmt.Errorf("orchestrator: not authorized to access this conversation")
			}
			return conv.ID, nil
		}
	}

	// Create a new conversation
	newConv := &models.Conversation{
		ID:     uuid.New(),
		UserID: userID,
	}
	if err := o.pool.CreateConversation(ctx, newConv); err != nil {
		return uuid.Nil, fmt.Errorf("orchestrator: create conversation: %w", err)
	}

	slog.Info("orchestrator.new_conversation", "id", newConv.ID, "user_id", userID)
	return newConv.ID, nil
}

// buildMessages loads conversation history and appends the new user prompt.
func (o *Orchestrator) buildMessages(ctx context.Context, conversationID uuid.UUID, prompt string) ([]models.LLMMessage, error) {
	dbMsgs, err := o.pool.GetConversationMessages(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: get history: %w", err)
	}

	// Limit to last 20 messages to stay within context window
	const maxHistory = 20
	if len(dbMsgs) > maxHistory {
		dbMsgs = dbMsgs[len(dbMsgs)-maxHistory:]
	}

	messages := make([]models.LLMMessage, 0, len(dbMsgs)+1)
	for _, m := range dbMsgs {
		messages = append(messages, models.LLMMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	// Append current user prompt
	messages = append(messages, models.LLMMessage{
		Role:    "user",
		Content: prompt,
	})

	return messages, nil
}

// persistMessages saves the user prompt and assistant response to the database.
func (o *Orchestrator) persistMessages(ctx context.Context, convID uuid.UUID, prompt, response, model string, tokens int, latencyMs float64) {
	// User message
	userMsg := &models.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           models.RoleUserMsg,
		Content:        prompt,
	}
	if err := o.pool.CreateMessage(ctx, userMsg); err != nil {
		slog.Error("orchestrator.persist_user_msg", "error", err)
	}

	// Assistant message
	var tokenCount *int
	if tokens > 0 {
		tokenCount = &tokens
	}
	var lat *float64
	if latencyMs > 0 {
		lat = &latencyMs
	}

	assistantMsg := &models.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           models.RoleAssistantMsg,
		Content:        response,
		TokenCount:     tokenCount,
		ModelUsed:      &model,
		LatencyMs:      lat,
	}
	if err := o.pool.CreateMessage(ctx, assistantMsg); err != nil {
		slog.Error("orchestrator.persist_assistant_msg", "error", err)
	}
}
