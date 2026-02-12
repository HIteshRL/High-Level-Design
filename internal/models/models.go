// Package models defines domain types shared across the application.
// Maps to design.swift: User (AuthN/AuthZ), Conversation/Message (Interaction Logger),
// and inference request/response contracts.
package models

import (
	"time"

	"github.com/google/uuid"
)

// ── User (AuthN/AuthZ backing store) ────────────────────────────────

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleParent  UserRole = "parent"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"`
	FullName       *string   `json:"full_name,omitempty" db:"full_name"`
	Role           UserRole  `json:"role" db:"role"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ── Conversation & Messages (Interaction Logger) ────────────────────

type Conversation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     *string   `json:"title,omitempty" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MessageRole string

const (
	RoleUserMsg      MessageRole = "user"
	RoleAssistantMsg MessageRole = "assistant"
	RoleSystemMsg    MessageRole = "system"
)

type Message struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	ConversationID uuid.UUID   `json:"conversation_id" db:"conversation_id"`
	Role           MessageRole `json:"role" db:"role"`
	Content        string      `json:"content" db:"content"`
	TokenCount     *int        `json:"token_count,omitempty" db:"token_count"`
	ModelUsed      *string     `json:"model_used,omitempty" db:"model_used"`
	LatencyMs      *float64    `json:"latency_ms,omitempty" db:"latency_ms"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// ── API Request/Response ────────────────────────────────────────────

// InferenceRequest is the client → Request Router contract.
type InferenceRequest struct {
	Prompt         string     `json:"prompt" validate:"required,min=1,max=32000"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	Stream         bool       `json:"stream"`
	Model          string     `json:"model,omitempty"`
	Temperature    *float64   `json:"temperature,omitempty"`
	MaxTokens      *int       `json:"max_tokens,omitempty"`
}

// InferenceResponse is the non-streaming response.
type InferenceResponse struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Content        string    `json:"content"`
	Model          string    `json:"model"`
	TokenCount     *int      `json:"token_count,omitempty"`
	LatencyMs      float64   `json:"latency_ms"`
	Cached         bool      `json:"cached"`
}

// StreamChunk is a single SSE event: Token Streamer → Client.
type StreamChunk struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Delta          string    `json:"delta"`
	FinishReason   *string   `json:"finish_reason,omitempty"`
	Model          string    `json:"model,omitempty"`
}

// AuthRequest represents login credentials.
type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents user registration.
type RegisterRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=64"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8,max=128"`
	FullName *string  `json:"full_name,omitempty"`
	Role     UserRole `json:"role,omitempty"`
}

// TokenResponse is the JWT token returned on auth.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// HealthResponse for the health check endpoint.
type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// LLM internal types for talking to the OpenAI-compatible API.

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream"`
}

type LLMChoice struct {
	Index        int        `json:"index"`
	Message      LLMMessage `json:"message,omitempty"`
	Delta        LLMMessage `json:"delta,omitempty"`
	FinishReason *string    `json:"finish_reason,omitempty"`
}

type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type LLMResponse struct {
	ID      string      `json:"id"`
	Choices []LLMChoice `json:"choices"`
	Usage   LLMUsage    `json:"usage"`
	Model   string      `json:"model"`
}
