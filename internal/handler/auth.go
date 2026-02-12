// Auth handler — register, login, get current user.
// Maps to design.swift: AuthN/AuthZ (OAuth/OIDC).
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/prakyathpnayak/roognis/internal/config"
	"github.com/prakyathpnayak/roognis/internal/middleware"
	"github.com/prakyathpnayak/roognis/internal/models"
	"github.com/prakyathpnayak/roognis/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	auth *service.Auth
	cfg  *config.Config
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(auth *service.Auth, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// H6 fix: Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeError(w, "username, email, and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		writeError(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	user, err := h.auth.Register(r.Context(), &req)
	if err != nil {
		slog.Warn("auth.register_error", "error", err)
		// H5 fix: Don't expose internal error details
		writeError(w, "registration failed — username or email may already be taken", http.StatusConflict)
		return
	}

	token, err := middleware.CreateToken(h.cfg, user)
	if err != nil {
		slog.Error("auth.create_token_error", "error", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, models.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   h.cfg.JWTExpiryMinutes * 60,
	})
}

// Login handles POST /api/v1/auth/token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.auth.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := middleware.CreateToken(h.cfg, user)
	if err != nil {
		slog.Error("auth.create_token_error", "error", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, models.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   h.cfg.JWTExpiryMinutes * 60,
	})
}

// Me handles GET /api/v1/auth/me (authenticated).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// ── Helpers ─────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	writeJSON(w, code, models.ErrorResponse{
		Error:   http.StatusText(code),
		Message: msg,
		Code:    code,
	})
}
