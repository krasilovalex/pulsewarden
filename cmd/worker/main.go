package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	appworker "github.com/wayzzoo/pulsewarden/internal/app/worker"
	"github.com/wayzzoo/pulsewarden/internal/checker"
	"github.com/wayzzoo/pulsewarden/internal/platform/config"
	"github.com/wayzzoo/pulsewarden/internal/platform/lifecycle"
	"github.com/wayzzoo/pulsewarden/internal/platform/logger"
	"github.com/wayzzoo/pulsewarden/internal/platform/postgres"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
	usecasecheck "github.com/wayzzoo/pulsewarden/internal/usecase/check"
)

const (
	workerPollInterval = time.Second
	workerBatchSize    = 100
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

	log, err := logger.New(
		os.Stdout,
		cfg.Logging.Level,
		"worker",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
		return 1
	}

	ctx, stop := lifecycle.SignalContext(
		context.Background(),
	)
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

	monitorRepository :=
		repositorypostgres.NewMonitorRepository(pool)

	checkResultRepository :=
		repositorypostgres.NewCheckResultRepository(pool)

	httpChecker := checker.NewHTTPChecker(
		&http.Client{},
	)

	runCheck := usecasecheck.NewRun(
		httpChecker,
		checkResultRepository,
	)

	runDue := usecasecheck.NewRunDue(
		monitorRepository,
		runCheck,
	)

	application, err := appworker.New(
		appworker.Config{
			Logger:       log,
			Runner:       runDue,
			PollInterval: workerPollInterval,
			BatchSize:    workerBatchSize,
		},
	)
	if err != nil {
		log.Error(
			"initialize worker",
			slog.Any("error", err),
		)

		return 1
	}

	log.Info(
		"application started",
		slog.String(
			"environment",
			cfg.Environment,
		),
		slog.String(
			"shutdown_timeout",
			cfg.Shutdown.Timeout.String(),
		),
		slog.String(
			"poll_interval",
			workerPollInterval.String(),
		),
		slog.Int(
			"batch_size",
			workerBatchSize,
		),
	)

	application.Run(ctx)

	log.Info("shutdown requested")
	log.Info("application stopped")

	return 0
}
