package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

func TestCheckResultRepositoryIntegration(t *testing.T) {
	if os.Getenv("PULSEWARDEN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}

	dsn := os.Getenv("PULSEWARDEN_POSTGRES_DSN")
	if dsn == "" {
		t.Fatal("PULSEWARDEN_POSTGRES_DSN is required")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	monitorRepository := NewMonitorRepository(pool)
	checkResultRepository := NewCheckResultRepository(pool)

	t.Run("create up result", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		createdMonitor := createCheckResultTestMonitor(
			t,
			monitorRepository,
		)

		statusCode := 204
		checkedAt := time.Now().
			UTC().
			Truncate(time.Microsecond)

		input, err := checkresult.New(
			createdMonitor.ID,
			checkresult.StatusUp,
			&statusCode,
			125*time.Millisecond,
			nil,
			checkedAt,
		)
		if err != nil {
			t.Fatalf("create domain check result: %v", err)
		}

		created, err := checkResultRepository.Create(
			context.Background(),
			input,
		)
		if err != nil {
			t.Fatalf("create check result: %v", err)
		}

		if created.ID == uuid.Nil {
			t.Fatal("created check result ID is nil")
		}

		input.ID = created.ID

		assertCheckResultEqual(t, created, input)
	})

	t.Run("create down result", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		createdMonitor := createCheckResultTestMonitor(
			t,
			monitorRepository,
		)

		errorMessage := "dial tcp: connection refused"
		checkedAt := time.Now().
			UTC().
			Truncate(time.Microsecond)

		input, err := checkresult.New(
			createdMonitor.ID,
			checkresult.StatusDown,
			nil,
			37*time.Millisecond,
			&errorMessage,
			checkedAt,
		)
		if err != nil {
			t.Fatalf("create domain check result: %v", err)
		}

		created, err := checkResultRepository.Create(
			context.Background(),
			input,
		)
		if err != nil {
			t.Fatalf("create check result: %v", err)
		}

		if created.ID == uuid.Nil {
			t.Fatal("created check result ID is nil")
		}

		input.ID = created.ID

		assertCheckResultEqual(t, created, input)
	})
}

func createCheckResultTestMonitor(
	t *testing.T,
	repository *MonitorRepository,
) monitor.Monitor {
	t.Helper()

	created, err := repository.Create(
		context.Background(),
		monitor.NewMonitor{
			Name:     "Check Result Test",
			URL:      "https://example.com/health",
			Interval: 30 * time.Second,
			Timeout:  time.Second,
			Enabled:  true,
			NextCheckAt: time.Now().
				UTC().
				Truncate(time.Microsecond),
		},
	)
	if err != nil {
		t.Fatalf("create test monitor: %v", err)
	}

	return created
}

func assertCheckResultEqual(
	t *testing.T,
	got checkresult.CheckResult,
	want checkresult.CheckResult,
) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("ID = %s, want %s", got.ID, want.ID)
	}

	if got.MonitorID != want.MonitorID {
		t.Fatalf(
			"MonitorID = %s, want %s",
			got.MonitorID,
			want.MonitorID,
		)
	}

	if got.Status != want.Status {
		t.Fatalf(
			"Status = %q, want %q",
			got.Status,
			want.Status,
		)
	}

	assertOptionalIntEqual(
		t,
		"StatusCode",
		got.StatusCode,
		want.StatusCode,
	)

	if got.Latency != want.Latency {
		t.Fatalf(
			"Latency = %s, want %s",
			got.Latency,
			want.Latency,
		)
	}

	assertOptionalStringEqual(
		t,
		"Error",
		got.Error,
		want.Error,
	)

	if !got.CheckedAt.Equal(want.CheckedAt) {
		t.Fatalf(
			"CheckedAt = %s, want %s",
			got.CheckedAt,
			want.CheckedAt,
		)
	}
}

func assertOptionalIntEqual(
	t *testing.T,
	field string,
	got *int,
	want *int,
) {
	t.Helper()

	if got == nil && want == nil {
		return
	}

	if got == nil || want == nil {
		t.Fatalf(
			"%s = %v, want %v",
			field,
			got,
			want,
		)
	}

	if *got != *want {
		t.Fatalf(
			"%s = %d, want %d",
			field,
			*got,
			*want,
		)
	}
}

func assertOptionalStringEqual(
	t *testing.T,
	field string,
	got *string,
	want *string,
) {
	t.Helper()

	if got == nil && want == nil {
		return
	}

	if got == nil || want == nil {
		t.Fatalf(
			"%s = %v, want %v",
			field,
			got,
			want,
		)
	}

	if *got != *want {
		t.Fatalf(
			"%s = %q, want %q",
			field,
			*got,
			*want,
		)
	}
}
