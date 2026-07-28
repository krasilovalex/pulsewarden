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

type dueMonitorRepositoryFunc func(
	context.Context,
	time.Time,
	int,
) ([]monitor.Monitor, error)

func (f dueMonitorRepositoryFunc) ClaimDue(
	ctx context.Context,
	dueAt time.Time,
	limit int,
) ([]monitor.Monitor, error) {
	return f(ctx, dueAt, limit)
}

type monitorRunnerFunc func(
	context.Context,
	monitor.Monitor,
) (checkresult.CheckResult, error)

func (f monitorRunnerFunc) Execute(
	ctx context.Context,
	input monitor.Monitor,
) (checkresult.CheckResult, error) {
	return f(ctx, input)
}

func TestRunDueExecute(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		28,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	firstID := uuid.New()
	secondID := uuid.New()
	thirdID := uuid.New()

	monitors := []monitor.Monitor{
		{
			ID:   firstID,
			Name: "First",
		},
		{
			ID:   secondID,
			Name: "Second",
		},
		{
			ID:   thirdID,
			Name: "Third",
		},
	}

	runErr := errors.New("save check result failed")

	var (
		claimedAt time.Time
		gotLimit  int
		runIDs    []uuid.UUID
	)

	useCase := NewRunDue(
		dueMonitorRepositoryFunc(func(
			_ context.Context,
			dueAt time.Time,
			limit int,
		) ([]monitor.Monitor, error) {
			claimedAt = dueAt
			gotLimit = limit

			return monitors, nil
		}),
		monitorRunnerFunc(func(
			_ context.Context,
			input monitor.Monitor,
		) (checkresult.CheckResult, error) {
			runIDs = append(runIDs, input.ID)

			if input.ID == secondID {
				return checkresult.CheckResult{}, runErr
			}

			return checkresult.CheckResult{
				ID:        uuid.New(),
				MonitorID: input.ID,
			}, nil
		}),
	)

	useCase.now = func() time.Time {
		return now
	}

	result, err := useCase.Execute(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !claimedAt.Equal(now) {
		t.Fatalf(
			"claim due at = %s, want %s",
			claimedAt,
			now,
		)
	}

	if gotLimit != 10 {
		t.Fatalf(
			"claim limit = %d, want 10",
			gotLimit,
		)
	}

	if len(runIDs) != 3 {
		t.Fatalf(
			"runner calls = %d, want 3",
			len(runIDs),
		)
	}

	for index, wantID := range []uuid.UUID{
		firstID,
		secondID,
		thirdID,
	} {
		if runIDs[index] != wantID {
			t.Fatalf(
				"runner call %d ID = %s, want %s",
				index,
				runIDs[index],
				wantID,
			)
		}
	}

	if result.Claimed != 3 {
		t.Fatalf(
			"claimed = %d, want 3",
			result.Claimed,
		)
	}

	if result.Succeeded != 2 {
		t.Fatalf(
			"succeeded = %d, want 2",
			result.Succeeded,
		)
	}

	if result.Failed != 1 {
		t.Fatalf(
			"failed = %d, want 1",
			result.Failed,
		)
	}

	if len(result.Failures) != 1 {
		t.Fatalf(
			"failures length = %d, want 1",
			len(result.Failures),
		)
	}

	failure := result.Failures[0]

	if failure.MonitorID != secondID {
		t.Fatalf(
			"failure monitor ID = %s, want %s",
			failure.MonitorID,
			secondID,
		)
	}

	if !errors.Is(failure.Err, runErr) {
		t.Fatalf(
			"failure error = %v, want %v",
			failure.Err,
			runErr,
		)
	}
}

func TestRunDueExecuteClaimError(t *testing.T) {
	claimErr := errors.New("database unavailable")
	runnerCalled := false

	useCase := NewRunDue(
		dueMonitorRepositoryFunc(func(
			context.Context,
			time.Time,
			int,
		) ([]monitor.Monitor, error) {
			return nil, claimErr
		}),
		monitorRunnerFunc(func(
			context.Context,
			monitor.Monitor,
		) (checkresult.CheckResult, error) {
			runnerCalled = true

			return checkresult.CheckResult{}, nil
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		10,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, claimErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			claimErr,
		)
	}

	if runnerCalled {
		t.Fatal("runner was called after claim error")
	}
}

func TestRunDueExecuteStopsOnContextCancellation(
	t *testing.T,
) {
	firstID := uuid.New()
	secondID := uuid.New()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	var runIDs []uuid.UUID

	useCase := NewRunDue(
		dueMonitorRepositoryFunc(func(
			context.Context,
			time.Time,
			int,
		) ([]monitor.Monitor, error) {
			return []monitor.Monitor{
				{ID: firstID},
				{ID: secondID},
			}, nil
		}),
		monitorRunnerFunc(func(
			_ context.Context,
			input monitor.Monitor,
		) (checkresult.CheckResult, error) {
			runIDs = append(runIDs, input.ID)

			cancel()

			return checkresult.CheckResult{
				MonitorID: input.ID,
			}, nil
		}),
	)

	result, err := useCase.Execute(ctx, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Execute() error = %v, want context canceled",
			err,
		)
	}

	if len(runIDs) != 1 {
		t.Fatalf(
			"runner calls = %d, want 1",
			len(runIDs),
		)
	}

	if runIDs[0] != firstID {
		t.Fatalf(
			"runner monitor ID = %s, want %s",
			runIDs[0],
			firstID,
		)
	}

	if result.Claimed != 2 {
		t.Fatalf(
			"claimed = %d, want 2",
			result.Claimed,
		)
	}

	if result.Succeeded != 1 {
		t.Fatalf(
			"succeeded = %d, want 1",
			result.Succeeded,
		)
	}
}
