package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auto-pr-review/internal/config"
	apphttp "auto-pr-review/internal/http"
	"auto-pr-review/internal/postgres"
	"auto-pr-review/internal/redisstore"
	"auto-pr-review/internal/scheduler"
	"auto-pr-review/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := postgres.NewRepository(ctx, cfg.Database)
	if err != nil {
		logger.Error("connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer store.Close()

	jobStore, err := redisstore.NewRepository(ctx, cfg.Redis)
	if err != nil {
		logger.Error("connect redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := jobStore.Close(); err != nil {
			logger.Error("close redis", slog.String("error", err.Error()))
		}
	}()

	githubLinks := service.NewGitHubLinkService(store)
	scheduler := scheduler.New(store, jobStore, logger)
	periodicJobs := service.NewPeriodicJobService(store, scheduler)
	handler := apphttp.NewServer(githubLinks, periodicJobs, cfg.Jobs.PeriodOptions, logger, cfg.HTTP.RequestTimeout)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.String("addr", cfg.HTTP.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}
