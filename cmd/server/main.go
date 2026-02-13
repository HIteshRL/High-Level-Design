// Server entry point — wires all components and starts the HTTP server.
// Maps to design.swift: API Gateway / Ingress (HTTP/gRPC) + Service Mesh wiring.
//
// Walking skeleton architecture:
//
//	External Clients → Load Balancer → Rate Limiter → API Gateway
//	→ Auth → Request Router → Prompt Orchestrator → Context Injector (RAG stub)
//	→ LLM Inference Node → Token Streamer (SSE) → Client
//
// Also wired: Response Cache (Redis), Interaction Logger (Postgres),
// and Health checks.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/prakyathpnayak/roognis/internal/config"
	"github.com/prakyathpnayak/roognis/internal/db"
	"github.com/prakyathpnayak/roognis/internal/handler"
	"github.com/prakyathpnayak/roognis/internal/middleware"
	"github.com/prakyathpnayak/roognis/internal/service"
)

func main() {
	// ── Load config ─────────────────────────────────────────────────
	_ = config.LoadDotEnv(".env")
	cfg := config.Load()
	cfg.Validate()

	// ── Structured logging ──────────────────────────────────────────
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("starting", "app", cfg.AppName, "env", cfg.AppEnv, "port", cfg.ServerPort)

	// ── Database ────────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.RunMigrations(ctx); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// ── Redis ───────────────────────────────────────────────────────
	var rdb *redis.Client
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Warn("redis URL parse error, running without cache", "error", err)
	} else {
		rdb = redis.NewClient(opts)
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Warn("redis connection failed, running without cache", "error", err)
			rdb = nil
		} else {
			slog.Info("redis connected")
			defer rdb.Close()
		}
	}

	// ── Services ────────────────────────────────────────────────────
	cache := service.NewCache(rdb, cfg.CacheTTL)
	llm := service.NewLLM(cfg)
	ctxInjector := service.NewContextInjector()
	orchestrator := service.NewOrchestrator(llm, cache, ctxInjector, pool)
	authSvc := service.NewAuth(pool)

	// ── Handlers ────────────────────────────────────────────────────
	healthHandler := handler.NewHealth(pool, cache)
	authHandler := handler.NewAuthHandler(authSvc, cfg)
	inferenceHandler := handler.NewInferenceHandler(orchestrator)
	attachmentHandler := handler.NewAttachmentHandler()

	// ── Middleware ───────────────────────────────────────────────────
	rateLimiter := middleware.NewRateLimiter(rdb, cfg.RateLimitRPM)
	authMiddleware := middleware.Auth(cfg, pool)

	// ── Router ──────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Public routes
	mux.Handle("GET /health", healthHandler)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/token", authHandler.Login)

	// Protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/v1/auth/me", authHandler.Me)
	protectedMux.HandleFunc("POST /api/v1/inference/complete", inferenceHandler.Complete)
	protectedMux.HandleFunc("GET /api/v1/conversations", inferenceHandler.Conversations)
	protectedMux.HandleFunc("GET /api/v1/conversations/{id}/messages", inferenceHandler.ConversationMessages)
	protectedMux.HandleFunc("GET /api/v1/conversation-messages", inferenceHandler.ConversationMessagesByQuery)
	protectedMux.HandleFunc("GET /api/v1/architecture/attachment-points", attachmentHandler.Catalog)
	protectedMux.HandleFunc("POST /api/v1/image/jobs", attachmentHandler.SubmitImageJob)
	protectedMux.HandleFunc("GET /api/v1/image/jobs/{id}", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/video/jobs", attachmentHandler.SubmitVideoJob)
	protectedMux.HandleFunc("GET /api/v1/video/jobs/{id}", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/rag/documents", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/rag/search", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/psychographic/events", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("GET /api/v1/psychographic/persona/{user_id}", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/quiz/generate", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("POST /api/v1/quiz/attempts", attachmentHandler.NotImplemented)
	protectedMux.HandleFunc("GET /api/v1/analytics/kpi", attachmentHandler.NotImplemented)

	// Wire protected routes through auth middleware
	mux.Handle("/api/v1/", authMiddleware(protectedMux))

	// ── Build middleware chain ───────────────────────────────────────
	// Order (outermost → innermost): Logger → CORS → RateLimiter → Router
	var h http.Handler = mux
	h = rateLimiter.Middleware(h)
	h = middleware.CORS(cfg.CORSOrigins)(h)
	h = middleware.Logger(h)

	// ── Server ──────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second, // Long for SSE streaming
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful shutdown ───────────────────────────────────────────
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
