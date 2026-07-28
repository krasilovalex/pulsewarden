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

type checkerFunc func(
	context.Context,
	monitor.Monitor,
) (checkresult.CheckResult, error)

func (f checkerFunc) Check(
	ctx context.Context,
	input monitor.Monitor,
) (checkresult.CheckResult, error) {
	return f(ctx, input)
}

type resultRepositoryFunc func(
	context.Context,
	checkresult.CheckResult,
) (checkresult.CheckResult, error)

func (f resultRepositoryFunc) Create(
	ctx context.Context,
	input checkresult.CheckResult,
) (checkresult.CheckResult, error) {
	return f(ctx, input)
}

func TestRunExecute(t *testing.T) {
	monitorID := uuid.New()
	resultID := uuid.New()

	inputMonitor := monitor.Monitor{
		ID:      monitorID,
		Name:    "Example API",
		URL:     "https://example.com/health",
		Method:  "GET",
		Timeout: 2 * time.Second,
	}

	statusCode := 204

	checkedResult := checkresult.CheckResult{
		MonitorID:  monitorID,
		Status:     checkresult.StatusUp,
		StatusCode: &statusCode,
		Latency:    125 * time.Millisecond,
		CheckedAt: time.Date(
			2026,
			time.July,
			28,
			12,
			0,
			0,
			0,
			time.UTC,
		),
	}

	persistedResult := checkedResult
	persistedResult.ID = resultID

	var checkerInput monitor.Monitor
	var repositoryInput checkresult.CheckResult

	useCase := NewRun(
		checkerFunc(func(
			_ context.Context,
			input monitor.Monitor,
		) (checkresult.CheckResult, error) {
			checkerInput = input
			return checkedResult, nil
		}),
		resultRepositoryFunc(func(
			_ context.Context,
			input checkresult.CheckResult,
		) (checkresult.CheckResult, error) {
			repositoryInput = input
			return persistedResult, nil
		}),
	)

	result, err := useCase.Execute(
		context.Background(),
		inputMonitor,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if checkerInput.ID != monitorID {
		t.Fatalf(
			"checker monitor ID = %s, want %s",
			checkerInput.ID,
			monitorID,
		)
	}

	if repositoryInput.MonitorID != monitorID {
		t.Fatalf(
			"repository monitor ID = %s, want %s",
			repositoryInput.MonitorID,
			monitorID,
		)
	}

	if repositoryInput.ID != uuid.Nil {
		t.Fatalf(
			"repository result ID = %s, want nil UUID",
			repositoryInput.ID,
		)
	}

	if result.ID != resultID {
		t.Fatalf(
			"result ID = %s, want %s",
			result.ID,
			resultID,
		)
	}

	if result.Status != checkresult.StatusUp {
		t.Fatalf(
			"result status = %q, want %q",
			result.Status,
			checkresult.StatusUp,
		)
	}
}

func TestRunExecuteStopsOnCheckerError(t *testing.T) {
	checkerErr := errors.New("checker failure")
	repositoryCalled := false

	useCase := NewRun(
		checkerFunc(func(
			context.Context,
			monitor.Monitor,
		) (checkresult.CheckResult, error) {
			return checkresult.CheckResult{}, checkerErr
		}),
		resultRepositoryFunc(func(
			context.Context,
			checkresult.CheckResult,
		) (checkresult.CheckResult, error) {
			repositoryCalled = true
			return checkresult.CheckResult{}, nil
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		monitor.Monitor{
			ID: uuid.New(),
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, checkerErr) {
		t.Fatalf(
			"error = %v, want wrapped checker error",
			err,
		)
	}

	if repositoryCalled {
		t.Fatal("repository was called after checker error")
	}
}

func TestRunExecuteWrapsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("repository failure")

	checkedResult := checkresult.CheckResult{
		MonitorID: uuid.New(),
		Status:    checkresult.StatusDown,
		Latency:   25 * time.Millisecond,
		CheckedAt: time.Now().UTC(),
	}

	errorMessage := "connection refused"
	checkedResult.Error = &errorMessage

	useCase := NewRun(
		checkerFunc(func(
			context.Context,
			monitor.Monitor,
		) (checkresult.CheckResult, error) {
			return checkedResult, nil
		}),
		resultRepositoryFunc(func(
			context.Context,
			checkresult.CheckResult,
		) (checkresult.CheckResult, error) {
			return checkresult.CheckResult{}, repositoryErr
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		monitor.Monitor{
			ID: checkedResult.MonitorID,
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"error = %v, want wrapped repository error",
			err,
		)
	}
}
