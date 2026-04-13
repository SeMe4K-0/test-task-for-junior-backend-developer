package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type config struct {
	HTTPAddr    string
	DatabaseDSN string
	Worker      WorkerCfg
}

type WorkerCfg struct {
	LookaheadDays int
	Interval      time.Duration
}

func LoadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		Worker: WorkerCfg{
			LookaheadDays: envOrDefaultInt("WORKER_LOOKAHEAD_DAYS", 7),
			Interval:      envOrDefaultDuration("WORKER_INTERVAL", 60*time.Second),
		},
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func envOrDefaultDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}
