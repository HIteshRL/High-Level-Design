// Hand-written SQL queries for the walking skeleton.
// Maps to design.swift: Interaction Logger (conversations/messages) + AuthN/AuthZ (users).
package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prakyathpnayak/roognis/internal/models"
)

// ── Users ──────────────────────────────────────────────────────────

// CreateUser inserts a new user row.
func (p *Pool) CreateUser(ctx context.Context, u *models.User) error {
	_, err := p.Exec(ctx, `
		INSERT INTO users (id, username, email, hashed_password, full_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Username, u.Email, u.HashedPassword, u.FullName, u.Role, u.IsActive,
	)
	if err != nil {
		return fmt.Errorf("db.CreateUser: %w", err)
	}
	return nil
}

// GetUserByUsername finds a user by username (nil if not found).
func (p *Pool) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := p.QueryRow(ctx, `
		SELECT id, username, email, hashed_password, full_name, role, is_active, created_at, updated_at
		FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.HashedPassword, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("db.GetUserByUsername: %w", err)
	}
	return &u, nil
}

// GetUserByID finds a user by ID (nil if not found).
func (p *Pool) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := p.QueryRow(ctx, `
		SELECT id, username, email, hashed_password, full_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.HashedPassword, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("db.GetUserByID: %w", err)
	}
	return &u, nil
}

// ── Conversations ──────────────────────────────────────────────────

// CreateConversation inserts a new conversation.
func (p *Pool) CreateConversation(ctx context.Context, c *models.Conversation) error {
	_, err := p.Exec(ctx, `
		INSERT INTO conversations (id, user_id, title)
		VALUES ($1, $2, $3)`,
		c.ID, c.UserID, c.Title,
	)
	if err != nil {
		return fmt.Errorf("db.CreateConversation: %w", err)
	}
	return nil
}

// GetConversation fetches a single conversation by ID.
func (p *Pool) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var c models.Conversation
	err := p.QueryRow(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations WHERE id = $1`, id,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("db.GetConversation: %w", err)
	}
	return &c, nil
}

// ListConversations returns all conversations for a given user (most recent first).
func (p *Pool) ListConversations(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	rows, err := p.Query(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 50`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("db.ListConversations: %w", err)
	}
	defer rows.Close()

	var convos []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListConversations scan: %w", err)
		}
		convos = append(convos, c)
	}
	return convos, rows.Err()
}

// ── Messages ───────────────────────────────────────────────────────

// CreateMessage inserts a new message.
func (p *Pool) CreateMessage(ctx context.Context, m *models.Message) error {
	_, err := p.Exec(ctx, `
		INSERT INTO messages (id, conversation_id, role, content, token_count, model_used, latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.TokenCount, m.ModelUsed, m.LatencyMs,
	)
	if err != nil {
		return fmt.Errorf("db.CreateMessage: %w", err)
	}
	return nil
}

// GetConversationMessages returns the messages for a conversation in chronological order.
func (p *Pool) GetConversationMessages(ctx context.Context, conversationID uuid.UUID) ([]models.Message, error) {
	rows, err := p.Query(ctx, `
		SELECT id, conversation_id, role, content, token_count, model_used, latency_ms, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC`, conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("db.GetConversationMessages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.TokenCount, &m.ModelUsed, &m.LatencyMs, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("db.GetConversationMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
