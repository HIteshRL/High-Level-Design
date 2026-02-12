// Package db manages the PostgreSQL connection pool and provides
// query functions. Uses pgx directly for maximum performance.
package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Pool wraps a pgx connection pool.
type Pool struct {
	*pgxpool.Pool
}

// New creates a new connection pool from the given DSN.
func New(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db: parse config: %w", err)
	}

	cfg.MaxConns = 20
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	// Redact credentials from log output
	slog.Info("database connected")
	return &Pool{pool}, nil
}

// RunMigrations executes the embedded SQL migration files.
// In production, use golang-migrate CLI. This is a bootstrap convenience.
func (p *Pool) RunMigrations(ctx context.Context) error {
	data, err := MigrationsFS.ReadFile("migrations/000001_init.up.sql")
	if err != nil {
		return fmt.Errorf("db: read migration: %w", err)
	}

	_, err = p.Exec(ctx, string(data))
	if err != nil {
		return fmt.Errorf("db: run migration: %w", err)
	}

	slog.Info("database migrations applied")
	return nil
}
