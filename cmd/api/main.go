package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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

	log.Info(
		"application started",
		slog.String("environment", cfg.Environment),
		slog.String("shutdown_timeout", cfg.ShutdownTimeout.String()),
	)

	<-ctx.Done()

	log.Info("shutdown requested")
	log.Info("stopped")

	return 0
}
