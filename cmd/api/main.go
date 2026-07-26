package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	appapi "github.com/wayzzoo/pulsewarden/internal/app/api"
	"github.com/wayzzoo/pulsewarden/internal/platform/config"
	"github.com/wayzzoo/pulsewarden/internal/platform/lifecycle"
	"github.com/wayzzoo/pulsewarden/internal/platform/logger"
	"github.com/wayzzoo/pulsewarden/internal/platform/postgres"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
	usecasemonitor "github.com/wayzzoo/pulsewarden/internal/usecase/monitor"
)

func main() {
	os.Exit(run())
}

func run() int {

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}
	log, err := logger.New(os.Stdout, cfg.Logging.Level, "api")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
		return 1
	}

	pool, err := postgres.Open(context.Background(), cfg.Postgres)
	if err != nil {
		log.Error(
			"initialize PostgreSQL",
			slog.Any("error", err),
		)
		return 1
	}
	defer pool.Close()

	monitorRepository := repositorypostgres.NewMonitorRepository(pool)
	createMonitor := usecasemonitor.NewCreate(monitorRepository)
	listMonitors := usecasemonitor.NewList(monitorRepository)
	getMonitor := usecasemonitor.NewGet(monitorRepository)

	ctx, stop := lifecycle.SignalContext(context.Background())
	defer stop()

	server := appapi.NewServer(appapi.ServerConfig{
		Address:           cfg.HTTP.Address,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ReadinessTimeout:  cfg.Postgres.ReadinessTimeout,
		Logger:            log,
		Postgres:          pool,
		MonitorCreator:    createMonitor,
		MonitorLister:     listMonitors,
		MonitorGetter:     getMonitor,
	})

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	log.Info(
		"HTTP server started",
		slog.String("environment", cfg.Environment),
		slog.String("address", cfg.HTTP.Address),
		slog.String("shutdown_timeout", cfg.Shutdown.Timeout.String()),
		slog.String("http_read_header_timeout", cfg.HTTP.ReadHeaderTimeout.String()),
		slog.String("http_read_timeout", cfg.HTTP.ReadTimeout.String()),
		slog.String("http_write_timeout", cfg.HTTP.WriteTimeout.String()),
		slog.String("http_idle_timeout", cfg.HTTP.IdleTimeout.String()),
	)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed", slog.Any("error", err))
			return 1
		}
	case <-ctx.Done():
		log.Info("shutdown requested")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.Shutdown.Timeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error(
				"graceful HTTP shutdown failed",
				slog.Any("error", err),
			)

			if closeErr := server.Close(); closeErr != nil {
				log.Error(
					"force HTTP shutdown failed",
					slog.Any("error", closeErr),
				)
			}

			return 1
		}

		err := <-serverErrors
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(
				"HTTP server stopped with error",
				slog.Any("error", err),
			)
			return 1
		}
	}

	log.Info("application stopped")

	return 0
}
