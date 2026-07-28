package check

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type DueMonitorRepository interface {
	ClaimDue(
		ctx context.Context,
		dueAt time.Time,
		limit int,
	) ([]monitor.Monitor, error)
}

type MonitorRunner interface {
	Execute(
		ctx context.Context,
		input monitor.Monitor,
	) (checkresult.CheckResult, error)
}

type RunDueFailure struct {
	MonitorID uuid.UUID
	Err       error
}

type RunDueResult struct {
	Claimed   int
	Succeeded int
	Failed    int
	Failures  []RunDueFailure
}

type RunDue struct {
	repository DueMonitorRepository
	runner     MonitorRunner
	now        func() time.Time
}

func NewRunDue(
	repository DueMonitorRepository,
	runner MonitorRunner,
) *RunDue {
	return &RunDue{
		repository: repository,
		runner:     runner,
		now:        time.Now,
	}
}

func (uc *RunDue) Execute(
	ctx context.Context,
	limit int,
) (RunDueResult, error) {
	dueAt := uc.now().UTC()

	monitors, err := uc.repository.ClaimDue(
		ctx,
		dueAt,
		limit,
	)
	if err != nil {
		return RunDueResult{}, fmt.Errorf(
			"claim due monitors: %w",
			err,
		)
	}

	result := RunDueResult{
		Claimed:  len(monitors),
		Failures: make([]RunDueFailure, 0),
	}

	for _, item := range monitors {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf(
				"run due monitor checks: %w",
				err,
			)
		}

		_, err := uc.runner.Execute(ctx, item)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return result, fmt.Errorf(
					"run due monitor checks: %w",
					ctxErr,
				)
			}

			result.Failed++
			result.Failures = append(result.Failures,
				RunDueFailure{
					MonitorID: item.ID,
					Err:       err,
				},
			)
			continue
		}

		result.Succeeded++
	}
	return result, nil
}
