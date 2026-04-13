package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/taskservice/internal/config"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
	"example.com/taskservice/internal/worker"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	taskUsecase := task.NewService(taskRepo)
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// worker
	w := worker.NewRecurrenceWorker(taskUsecase, logger, cfg.Worker)
	ticker := time.NewTicker(cfg.Worker.Interval)
	defer ticker.Stop()

	go func() {
		if err := w.Run(ctx); err != nil {
			logger.Error("worker initial run failed", "error", err)
		}
		for {
			select {
			case <-ctx.Done():
				logger.Info("recurrence worker stopped")
				return
			case t := <-ticker.C:
				logger.Info("worker tick", "time", t)
				if err := w.Run(ctx); err != nil {
					logger.Error("worker run failed", "error", err)
				}
			}
		}
	}()

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
