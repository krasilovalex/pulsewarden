package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	usecasecheck "github.com/wayzzoo/pulsewarden/internal/usecase/check"
)

type DueRunner interface {
	Execute(
		ctx context.Context,
		limit int,
	) (usecasecheck.RunDueResult, error)
}

type Config struct {
	Logger       *slog.Logger
	Runner       DueRunner
	PollInterval time.Duration
	BatchSize    int
}

type Worker struct {
	logger       *slog.Logger
	runner       DueRunner
	pollInterval time.Duration
	batchSize    int
}

func New(cfg Config) (*Worker, error) {
	if cfg.Logger == nil {
		return nil, fmt.Errorf("worker logger must not be nil")
	}

	if cfg.Runner == nil {
		return nil, fmt.Errorf("worker runner must not be nil")
	}

	if cfg.PollInterval <= 0 {
		return nil, fmt.Errorf(
			"worker pool interval must be greater than zero",
		)
	}

	if cfg.BatchSize <= 0 {
		return nil, fmt.Errorf(
			"worker batch size must be greater than zero",
		)
	}

	return &Worker{
		logger:       cfg.Logger,
		runner:       cfg.Runner,
		pollInterval: cfg.PollInterval,
		batchSize:    cfg.BatchSize,
	}, nil
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.run(ctx, ticker.C)
}

func (w *Worker) run(
	ctx context.Context,
	ticks <-chan time.Time,
) {
	w.runCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			w.runCycle(ctx)
		}
	}
}

func (w *Worker) runCycle(ctx context.Context) {
	result, err := w.runner.Execute(
		ctx,
		w.batchSize,
	)
	if err != nil {
		if ctx.Err() != nil {
			return
		}

		w.logger.Error(
			"scheduler cycle failed",
			slog.Any("error", err),
		)

		return
	}

	for _, failure := range result.Failures {
		w.logger.Error(
			"monitor check failed",
			slog.String(
				"monitor_id",
				failure.MonitorID.String(),
			),
			slog.Any("error", failure.Err),
		)
	}

	if result.Claimed == 0 {
		return
	}

	w.logger.Info(
		"monitor checks completed",
		slog.Int("claimed", result.Claimed),
		slog.Int("succeeded", result.Succeeded),
		slog.Int("failed", result.Failed),
	)
}
