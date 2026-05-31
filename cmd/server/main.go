package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"raevtar/internal/config"
	"raevtar/internal/handler"
	"raevtar/internal/repo"
	"raevtar/internal/service"
)

func main() {
	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	db := repo.InitSQLite(cfg.DatabasePath)
	repo.AutoMigrate(db)

	repos := repo.New(db)
	svc := service.New(repos, cfg)

	// seed categories if empty
	if err := svc.SeedData(); err != nil {
		slog.Warn("seed skipped (maybe already seeded)", "err", err)
	}

	r := handler.New(svc, cfg)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler.WithSecurityHeaders(r),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("raevtar running", "addr", cfg.Addr, "domain", cfg.Domain)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server died", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for signal
	sig := <-quit
	slog.Info("shutting down", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}
