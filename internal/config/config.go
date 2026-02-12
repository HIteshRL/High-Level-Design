// Package config provides application configuration loaded from environment variables.
// Maps to design.swift: cross-cutting configuration for all subsystems.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application settings.
type Config struct {
	// App
	AppName    string
	AppEnv     string
	LogLevel   string
	ServerPort string

	// Auth
	JWTSecret        string
	JWTExpiryMinutes int

	// Database
	DatabaseURL string

	// Redis
	RedisURL string
	CacheTTL time.Duration

	// LLM
	LLMModel          string
	LLMAPIKey         string
	LLMAPIBase        string
	LLMTemperature    float64
	LLMMaxTokens      int
	LLMTimeoutSeconds int

	// Rate Limiting
	RateLimitRPM int

	// CORS
	CORSOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		AppName:    envOrDefault("APP_NAME", "roognis"),
		AppEnv:     envOrDefault("APP_ENV", "development"),
		LogLevel:   envOrDefault("LOG_LEVEL", "debug"),
		ServerPort: envOrDefault("SERVER_PORT", "8080"),

		JWTSecret:        envOrDefault("JWT_SECRET", "CHANGE_ME_generate_with_openssl_rand_hex_32"),
		JWTExpiryMinutes: envOrDefaultInt("JWT_EXPIRY_MINUTES", 30),

		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://roognis:roognis_secret@localhost:5432/roognis?sslmode=disable"),

		RedisURL: envOrDefault("REDIS_URL", "redis://:roognis_redis_secret@localhost:6379/0"),
		CacheTTL: time.Duration(envOrDefaultInt("CACHE_TTL_SECONDS", 3600)) * time.Second,

		LLMModel:          envOrDefault("LLM_MODEL", "gpt-4o-mini"),
		LLMAPIKey:         envOrDefault("LLM_API_KEY", ""),
		LLMAPIBase:        envOrDefault("LLM_API_BASE", "https://api.openai.com/v1"),
		LLMTemperature:    envOrDefaultFloat("LLM_TEMPERATURE", 0.7),
		LLMMaxTokens:      envOrDefaultInt("LLM_MAX_TOKENS", 2048),
		LLMTimeoutSeconds: envOrDefaultInt("LLM_TIMEOUT_SECONDS", 60),

		RateLimitRPM: envOrDefaultInt("RATE_LIMIT_RPM", 60),

		CORSOrigins: strings.Split(envOrDefault("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:8080"), ","),
	}
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// Validate ensures critical configuration is set. Panics in non-development
// mode if JWT_SECRET is the default insecure value or too short.
func (c *Config) Validate() {
	if !c.IsDevelopment() {
		if c.JWTSecret == "CHANGE_ME_generate_with_openssl_rand_hex_32" || len(c.JWTSecret) < 32 {
			panic("FATAL: JWT_SECRET must be set to a random value >= 32 chars in non-development mode")
		}
		if c.LLMAPIKey == "" {
			panic("FATAL: LLM_API_KEY must be set in non-development mode")
		}
	}
}

// Addr returns the server listen address.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.ServerPort)
}

// LoadDotEnv loads a .env file into the process environment (simple parser).
func LoadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip surrounding quotes (single or double)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envOrDefaultFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
