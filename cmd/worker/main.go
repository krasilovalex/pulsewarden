package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/krasilovalex/pulsewarden/internal/platform/config"
	"github.com/krasilovalex/pulsewarden/internal/platform/lifecycle"
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
	logger := log.New(os.Stdout, "worker: ", log.LstdFlags|log.LUTC|log.Lmsgprefix)

	ctx, stop := lifecycle.SignalContext(context.Background())
	defer stop()

	logger.Printf(
		"started environment=%s log_level=%s shutdown_timeout=%s",
		cfg.Environment,
		cfg.LogLevel,
		cfg.ShutdownTimeout,
	)

	<-ctx.Done()

	logger.Println("shutdown requested")
	logger.Println("stopped")

	return 0
}
