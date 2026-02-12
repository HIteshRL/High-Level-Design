// JWT authentication middleware.
// Maps to design.swift: AuthN/AuthZ (OAuth/OIDC) — JWT flavour.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prakyathpnayak/roognis/internal/config"
	"github.com/prakyathpnayak/roognis/internal/db"
	"github.com/prakyathpnayak/roognis/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

// Claims is the JWT payload.
type Claims struct {
	jwt.RegisteredClaims
	UserID   uuid.UUID       `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
}

// CreateToken generates a signed JWT for the given user.
func CreateToken(cfg *config.Config, user *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.AppName,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWTExpiryMinutes) * time.Minute)),
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("jwt: sign: %w", err)
	}
	return signed, nil
}

// Auth returns middleware that validates JWT bearer tokens.
func Auth(cfg *config.Config, pool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"missing authorization header","code":401}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization format","code":401}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token","code":401}`, http.StatusUnauthorized)
				return
			}

			// Fetch full user from DB to ensure they still exist and are active.
			user, err := pool.GetUserByID(r.Context(), claims.UserID)
			if err != nil || user == nil || !user.IsActive {
				http.Error(w, `{"error":"user not found or deactivated","code":401}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userContextKey).(*models.User)
	return u
}

// RequireRole returns middleware that rejects requests from users without the required role.
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil || !allowed[user.Role] {
				http.Error(w, `{"error":"forbidden","code":403}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
