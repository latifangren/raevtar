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
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}
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

	// Start scheduler for scheduled post publishing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runScheduler(ctx, svc)

	// Start stale node checker — fires webhook alert if a server hasn't been seen in 15+ minutes
	staleCheckInterval := 5 * time.Minute
	staleThreshold := 15 * time.Minute
	go runStaleChecker(ctx, svc, staleCheckInterval, staleThreshold)

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}

// runStaleChecker periodically checks for servers that haven't been seen recently.
func runStaleChecker(ctx context.Context, svc *service.Service, checkInterval, staleThreshold time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Run once immediately on startup
	cutoff := time.Now().UTC().Add(-staleThreshold)
	if stale := svc.Monitor.CheckStaleServers(cutoff); len(stale) > 0 {
		slog.Warn("stale checker: initial sweep", "stale_count", len(stale))
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("stale checker: stopping")
			return
		case <-ticker.C:
			cutoff := time.Now().UTC().Add(-staleThreshold)
			if stale := svc.Monitor.CheckStaleServers(cutoff); len(stale) > 0 {
				slog.Warn("stale checker: found", "stale_count", len(stale))
			}
		}
	}
}

// runScheduler ticks every 60s and publishes due scheduled posts.
func runScheduler(ctx context.Context, svc *service.Service) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Run once immediately on startup
	if n, err := svc.Blog.PublishScheduled(); err != nil {
		slog.Error("scheduler: initial publish failed", "error", err)
	} else if n > 0 {
		slog.Info("scheduler: initial publish", "count", n)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler: stopping")
			return
		case <-ticker.C:
			if n, err := svc.Blog.PublishScheduled(); err != nil {
				slog.Error("scheduler: publish failed", "error", err)
			} else if n > 0 {
				slog.Info("scheduler: published", "count", n)
			}
		}
	}
}
