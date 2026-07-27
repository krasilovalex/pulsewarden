package checkresult

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		26,
		16,
		0,
		0,
		0,
		time.UTC,
	)

	statusOK := 200
	statusInternalError := 500
	timeoutError := "context deadline exceeded"

	tests := []struct {
		name       string
		monitorID  uuid.UUID
		status     Status
		statusCode *int
		latency    time.Duration
		error      *string
		checkedAt  time.Time
		wantErr    bool
	}{
		{
			name:       "up result",
			monitorID:  uuid.New(),
			status:     StatusUp,
			statusCode: &statusOK,
			latency:    150 * time.Millisecond,
			checkedAt:  now,
		},
		{
			name:       "down result with status code",
			monitorID:  uuid.New(),
			status:     StatusDown,
			statusCode: &statusInternalError,
			latency:    200 * time.Millisecond,
			checkedAt:  now,
		},
		{
			name:      "down result with network error",
			monitorID: uuid.New(),
			status:    StatusDown,
			latency:   time.Second,
			error:     &timeoutError,
			checkedAt: now,
		},
		{
			name:       "nil monitor id",
			status:     StatusUp,
			statusCode: &statusOK,
			checkedAt:  now,
			wantErr:    true,
		},
		{
			name:       "invalid status",
			monitorID:  uuid.New(),
			status:     Status("unknown"),
			statusCode: &statusOK,
			checkedAt:  now,
			wantErr:    true,
		},
		{
			name:       "negative latency",
			monitorID:  uuid.New(),
			status:     StatusUp,
			statusCode: &statusOK,
			latency:    -time.Millisecond,
			checkedAt:  now,
			wantErr:    true,
		},
		{
			name:       "zero checked at",
			monitorID:  uuid.New(),
			status:     StatusUp,
			statusCode: &statusOK,
			wantErr:    true,
		},
		{
			name:       "invalid status code",
			monitorID:  uuid.New(),
			status:     StatusUp,
			statusCode: intPointer(99),
			checkedAt:  now,
			wantErr:    true,
		},
		{
			name:      "up without status code",
			monitorID: uuid.New(),
			status:    StatusUp,
			checkedAt: now,
			wantErr:   true,
		},
		{
			name:       "up with error",
			monitorID:  uuid.New(),
			status:     StatusUp,
			statusCode: &statusOK,
			error:      &timeoutError,
			checkedAt:  now,
			wantErr:    true,
		},
		{
			name:      "down without status code or error",
			monitorID: uuid.New(),
			status:    StatusDown,
			checkedAt: now,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := New(
				test.monitorID,
				test.status,
				test.statusCode,
				test.latency,
				test.error,
				test.checkedAt,
			)

			if test.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, ErrInvalidCheckResult) {
					t.Fatalf(
						"error = %v, want ErrInvalidCheckResult",
						err,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if result.MonitorID != test.monitorID {
				t.Fatalf(
					"MonitorID = %s, want %s",
					result.MonitorID,
					test.monitorID,
				)
			}

			if result.Status != test.status {
				t.Fatalf(
					"Status = %q, want %q",
					result.Status,
					test.status,
				)
			}

			if !result.CheckedAt.Equal(now) {
				t.Fatalf(
					"CheckedAt = %s, want %s",
					result.CheckedAt,
					now,
				)
			}
		})
	}
}

func TestCheckResultIsUp(t *testing.T) {
	result := CheckResult{
		Status: StatusUp,
	}

	if !result.IsUp() {
		t.Fatal("IsUp() = false, want true")
	}
}

func intPointer(value int) *int {
	return &value
}
