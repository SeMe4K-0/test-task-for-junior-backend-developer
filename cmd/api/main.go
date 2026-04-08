package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := infrastructurepostgres.ApplyMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	repo := postgresrepo.New(pool)
	service := taskusecase.NewService(repo)
	taskHandler := httphandlers.NewTaskHandler(service)
	templateHandler := httphandlers.NewTemplateHandler(service)
	healthHandler := httphandlers.NewHealthHandler(func(r *http.Request) error {
		healthCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		return repo.Ping(healthCtx)
	})
	docsHandler := swaggerdocs.NewHandler()

	router := transporthttp.NewRouter(logger, healthHandler, taskHandler, templateHandler, docsHandler)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr      string
	DatabaseDSN   string
	MigrationsDir string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:   envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		MigrationsDir: envOrDefault("MIGRATIONS_DIR", "migrations"),
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
