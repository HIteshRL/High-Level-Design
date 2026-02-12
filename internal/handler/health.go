// Health check handler.
// Maps to design.swift: Health + Readiness probes.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/prakyathpnayak/roognis/internal/db"
	"github.com/prakyathpnayak/roognis/internal/models"
	"github.com/prakyathpnayak/roognis/internal/service"
)

// Health handles GET /health.
type Health struct {
	pool  *db.Pool
	cache *service.Cache
}

// NewHealth creates a new health handler.
func NewHealth(pool *db.Pool, cache *service.Cache) *Health {
	return &Health{pool: pool, cache: cache}
}

// ServeHTTP implements http.Handler.
func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := models.HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	}

	// Check database
	if err := h.pool.Ping(r.Context()); err != nil {
		resp.Database = "down"
		resp.Status = "degraded"
	} else {
		resp.Database = "up"
	}

	// Check Redis
	if err := h.cache.Ping(r.Context()); err != nil {
		resp.Redis = "down"
		if resp.Status == "ok" {
			resp.Status = "degraded"
		}
	} else {
		resp.Redis = "up"
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}
