package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
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
