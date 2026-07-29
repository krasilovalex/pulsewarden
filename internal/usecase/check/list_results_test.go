package check

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type historyMonitorRepositoryFunc func(
	context.Context,
	uuid.UUID,
) (monitor.Monitor, error)

func (f historyMonitorRepositoryFunc) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (monitor.Monitor, error) {
	return f(ctx, id)
}

type historyResultRepositoryFunc func(
	context.Context,
	uuid.UUID,
	int,
) ([]checkresult.CheckResult, error)

func (f historyResultRepositoryFunc) ListByMonitorID(
	ctx context.Context,
	monitorID uuid.UUID,
	limit int,
) ([]checkresult.CheckResult, error) {
	return f(ctx, monitorID, limit)
}

func TestListResultsExecute(t *testing.T) {
	monitorID := uuid.New()
	resultID := uuid.New()

	expected := []checkresult.CheckResult{
		{
			ID:        resultID,
			MonitorID: monitorID,
			Status:    checkresult.StatusUp,
			Latency:   42 * time.Millisecond,
			CheckedAt: time.Now().UTC(),
		},
	}

	var (
		gotMonitorID uuid.UUID
		gotLimit     int
	)

	useCase := NewListResults(
		historyMonitorRepositoryFunc(func(
			_ context.Context,
			id uuid.UUID,
		) (monitor.Monitor, error) {
			gotMonitorID = id

			return monitor.Monitor{
				ID: id,
			}, nil
		}),
		historyResultRepositoryFunc(func(
			_ context.Context,
			id uuid.UUID,
			limit int,
		) ([]checkresult.CheckResult, error) {
			if id != monitorID {
				t.Fatalf(
					"result monitor ID = %s, want %s",
					id,
					monitorID,
				)
			}

			gotLimit = limit

			return expected, nil
		}),
	)

	results, err := useCase.Execute(
		context.Background(),
		monitorID,
		50,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if gotMonitorID != monitorID {
		t.Fatalf(
			"monitor ID = %s, want %s",
			gotMonitorID,
			monitorID,
		)
	}

	if gotLimit != 50 {
		t.Fatalf(
			"limit = %d, want 50",
			gotLimit,
		)
	}

	if len(results) != 1 {
		t.Fatalf(
			"results length = %d, want 1",
			len(results),
		)
	}

	if results[0].ID != resultID {
		t.Fatalf(
			"result ID = %s, want %s",
			results[0].ID,
			resultID,
		)
	}
}

func TestListResultsStopsWhenMonitorDoesNotExist(
	t *testing.T,
) {
	monitorErr := errors.New("monitor not found")
	resultsCalled := false

	useCase := NewListResults(
		historyMonitorRepositoryFunc(func(
			context.Context,
			uuid.UUID,
		) (monitor.Monitor, error) {
			return monitor.Monitor{}, monitorErr
		}),
		historyResultRepositoryFunc(func(
			context.Context,
			uuid.UUID,
			int,
		) ([]checkresult.CheckResult, error) {
			resultsCalled = true

			return nil, nil
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		uuid.New(),
		100,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, monitorErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			monitorErr,
		)
	}

	if resultsCalled {
		t.Fatal("results repository was called")
	}
}

func TestListResultsWrapsResultsRepositoryError(
	t *testing.T,
) {
	repositoryErr := errors.New("database unavailable")

	useCase := NewListResults(
		historyMonitorRepositoryFunc(func(
			_ context.Context,
			id uuid.UUID,
		) (monitor.Monitor, error) {
			return monitor.Monitor{
				ID: id,
			}, nil
		}),
		historyResultRepositoryFunc(func(
			context.Context,
			uuid.UUID,
			int,
		) ([]checkresult.CheckResult, error) {
			return nil, repositoryErr
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		uuid.New(),
		100,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			repositoryErr,
		)
	}
}
