package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	appapi "github.com/krasilovalex/pulsewarden/internal/app/api"
	"github.com/krasilovalex/pulsewarden/internal/platform/config"
	"github.com/krasilovalex/pulsewarden/internal/platform/lifecycle"
	"github.com/krasilovalex/pulsewarden/internal/platform/logger"
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
	log, err := logger.New(os.Stdout, cfg.LogLevel, "api")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
		return 1
	}

	ctx, stop := lifecycle.SignalContext(context.Background())
	defer stop()

	server := appapi.NewServer(appapi.ServerConfig{
		Address:           cfg.HTTPAddress,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		Logger:            log,
	})

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	log.Info(
		"HTTP server started",
		slog.String("environment", cfg.Environment),
		slog.String("address", cfg.HTTPAddress),
		slog.String("shutdown_timeout", cfg.ShutdownTimeout.String()),
		slog.String("http_read_header_timeout", cfg.HTTPReadHeaderTimeout.String()),
		slog.String("http_read_timeout", cfg.HTTPReadTimeout.String()),
		slog.String("http_write_timeout", cfg.HTTPWriteTimeout.String()),
		slog.String("http_idle_timeout", cfg.HTTPIdleTimeout.String()),
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
			cfg.ShutdownTimeout,
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
