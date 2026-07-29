package check

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type ResultHistoryMonitorRepository interface {
	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (monitor.Monitor, error)
}

type ResultHistoryRepository interface {
	ListByMonitorID(
		ctx context.Context,
		monitorID uuid.UUID,
		limit int,
	) ([]checkresult.CheckResult, error)
}

type ListResults struct {
	monitors ResultHistoryMonitorRepository
	results  ResultHistoryRepository
}

func NewListResults(
	monitors ResultHistoryMonitorRepository,
	results ResultHistoryRepository,
) *ListResults {
	return &ListResults{
		monitors: monitors,
		results:  results,
	}
}

func (uc *ListResults) Execute(
	ctx context.Context,
	monitorID uuid.UUID,
	limit int,
) ([]checkresult.CheckResult, error) {
	if _, err := uc.monitors.GetByID(
		ctx,
		monitorID,
	); err != nil {
		return nil, fmt.Errorf(
			"get monitor: %w",
			err,
		)
	}

	results, err := uc.results.ListByMonitorID(
		ctx,
		monitorID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list monitor check results: %w",
			err,
		)
	}

	return results, nil
}
