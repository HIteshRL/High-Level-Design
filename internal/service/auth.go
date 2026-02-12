// Auth service — password hashing and user authentication.
// Maps to design.swift: AuthN/AuthZ (OAuth/OIDC)
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/prakyathpnayak/roognis/internal/db"
	"github.com/prakyathpnayak/roognis/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Auth provides user authentication services.
type Auth struct {
	pool *db.Pool
}

// NewAuth creates a new auth service.
func NewAuth(pool *db.Pool) *Auth {
	return &Auth{pool: pool}
}

// HashPassword generates a bcrypt hash of the password.
// H4 fix: Pre-hash with SHA-256 to avoid bcrypt's 72-byte truncation.
func HashPassword(password string) (string, error) {
	sha := sha256.Sum256([]byte(password))
	preHash := hex.EncodeToString(sha[:])
	hash, err := bcrypt.GenerateFromPassword([]byte(preHash), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks a password against its hash.
func VerifyPassword(password, hash string) bool {
	sha := sha256.Sum256([]byte(password))
	preHash := hex.EncodeToString(sha[:])
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(preHash)) == nil
}

// Register creates a new user.
func (a *Auth) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if username exists
	existing, err := a.pool.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("auth: check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already taken")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// C2 fix: Always assign student role at registration.
	// Admin/teacher roles must be granted by an admin through a separate flow.
	role := models.RoleStudent

	user := &models.User{
		ID:             uuid.New(),
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: hash,
		FullName:       req.FullName,
		Role:           role,
		IsActive:       true,
	}

	if err := a.pool.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	return user, nil
}

// Authenticate verifies credentials and returns the user.
func (a *Auth) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	user, err := a.pool.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("auth: lookup: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !VerifyPassword(password, user.HashedPassword) {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("account deactivated")
	}
	return user, nil
}
