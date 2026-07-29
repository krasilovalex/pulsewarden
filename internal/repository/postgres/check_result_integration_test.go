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

	t.Run("list results by monitor ID", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		firstMonitor := createCheckResultTestMonitor(
			t,
			monitorRepository,
		)

		secondMonitor := createCheckResultTestMonitor(
			t,
			monitorRepository,
		)

		baseTime := time.Now().
			UTC().
			Truncate(time.Microsecond)

		statusOK := 200
		statusUnavailable := 503
		statusNoContent := 204

		newestFirstResult, err := checkresult.New(
			firstMonitor.ID,
			checkresult.StatusUp,
			&statusOK,
			40*time.Millisecond,
			nil,
			baseTime.Add(2*time.Minute),
		)
		if err != nil {
			t.Fatalf(
				"create newest domain result: %v",
				err,
			)
		}

		newestFirstResult, err =
			checkResultRepository.Create(
				context.Background(),
				newestFirstResult,
			)
		if err != nil {
			t.Fatalf(
				"persist newest first result: %v",
				err,
			)
		}

		oldestFirstResult, err := checkresult.New(
			firstMonitor.ID,
			checkresult.StatusDown,
			&statusUnavailable,
			120*time.Millisecond,
			nil,
			baseTime,
		)
		if err != nil {
			t.Fatalf(
				"create oldest domain result: %v",
				err,
			)
		}

		oldestFirstResult, err =
			checkResultRepository.Create(
				context.Background(),
				oldestFirstResult,
			)
		if err != nil {
			t.Fatalf(
				"persist oldest first result: %v",
				err,
			)
		}

		secondMonitorResult, err := checkresult.New(
			secondMonitor.ID,
			checkresult.StatusUp,
			&statusNoContent,
			25*time.Millisecond,
			nil,
			baseTime.Add(time.Minute),
		)
		if err != nil {
			t.Fatalf(
				"create second monitor result: %v",
				err,
			)
		}

		_, err = checkResultRepository.Create(
			context.Background(),
			secondMonitorResult,
		)
		if err != nil {
			t.Fatalf(
				"persist second monitor result: %v",
				err,
			)
		}

		results, err :=
			checkResultRepository.ListByMonitorID(
				context.Background(),
				firstMonitor.ID,
				10,
			)
		if err != nil {
			t.Fatalf(
				"list results by monitor ID: %v",
				err,
			)
		}

		if len(results) != 2 {
			t.Fatalf(
				"results length = %d, want 2",
				len(results),
			)
		}

		assertCheckResultEqual(
			t,
			results[0],
			newestFirstResult,
		)

		assertCheckResultEqual(
			t,
			results[1],
			oldestFirstResult,
		)
	})

	t.Run("list results respects limit", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		createdMonitor := createCheckResultTestMonitor(
			t,
			monitorRepository,
		)

		statusCode := 200
		baseTime := time.Now().
			UTC().
			Truncate(time.Microsecond)

		for index := range 3 {
			input, err := checkresult.New(
				createdMonitor.ID,
				checkresult.StatusUp,
				&statusCode,
				time.Duration(index+1)*time.Millisecond,
				nil,
				baseTime.Add(
					time.Duration(index)*time.Minute,
				),
			)
			if err != nil {
				t.Fatalf(
					"create domain result %d: %v",
					index,
					err,
				)
			}

			if _, err := checkResultRepository.Create(
				context.Background(),
				input,
			); err != nil {
				t.Fatalf(
					"persist result %d: %v",
					index,
					err,
				)
			}
		}

		results, err :=
			checkResultRepository.ListByMonitorID(
				context.Background(),
				createdMonitor.ID,
				2,
			)
		if err != nil {
			t.Fatalf(
				"list limited results: %v",
				err,
			)
		}

		if len(results) != 2 {
			t.Fatalf(
				"results length = %d, want 2",
				len(results),
			)
		}

		if !results[0].CheckedAt.Equal(
			baseTime.Add(2 * time.Minute),
		) {
			t.Fatalf(
				"first checked at = %s, want %s",
				results[0].CheckedAt,
				baseTime.Add(2*time.Minute),
			)
		}

		if !results[1].CheckedAt.Equal(
			baseTime.Add(time.Minute),
		) {
			t.Fatalf(
				"second checked at = %s, want %s",
				results[1].CheckedAt,
				baseTime.Add(time.Minute),
			)
		}
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
