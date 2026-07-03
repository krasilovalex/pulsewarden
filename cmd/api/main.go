package main

import (
	"context"
	"log"
	"os"

	"github.com/krasilovalex/pulsewarden/internal/platform/lifecycle"
)

func main() {
	logger := log.New(os.Stdout, "api: ", log.LstdFlags|log.LUTC|log.Lmsgprefix)

	ctx, stop := lifecycle.SignalContext(context.Background())
	defer stop()

	logger.Println("started")

	<-ctx.Done()

	logger.Println("shutdown requested")
	logger.Println("stopped")
}
