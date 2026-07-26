package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/wayzzoo/pulsewarden/internal/platform/config"
	"github.com/wayzzoo/pulsewarden/internal/platform/lifecycle"
	"github.com/wayzzoo/pulsewarden/internal/platform/logger"
	"github.com/wayzzoo/pulsewarden/internal/platform/postgres"
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
	log, err := logger.New(os.Stdout, cfg.Logging.Level, "worker")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
		return 1
	}

	ctx, stop := lifecycle.SignalContext(context.Background())
	defer stop()

	pool, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		log.Error(
			"initialize PostgreSQL",
			slog.Any("error", err),
		)
		return 1
	}
	defer pool.Close()

	log.Info(
		"application started",
		slog.String("environment", cfg.Environment),
		slog.String("shutdown_timeout", cfg.Shutdown.Timeout.String()),
	)

	<-ctx.Done()

	log.Info("shutdown requested")
	log.Info("stopped")

	return 0
}
