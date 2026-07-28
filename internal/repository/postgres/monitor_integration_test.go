package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

func TestMonitorRepositoryIntegration(t *testing.T) {
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

	cleanMonitorsTable(t, pool)

	repository := NewMonitorRepository(pool)

	t.Run("create and get monitor", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		nextCheckAt := time.Now().
			UTC().
			Truncate(time.Microsecond).
			Add(time.Minute)

		created, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:               "Example API",
				URL:                "https://example.com/health",
				Interval:           30 * time.Second,
				Timeout:            1500 * time.Millisecond,
				ExpectedStatusFrom: 200,
				ExpectedStatusTo:   204,
				Enabled:            true,
				NextCheckAt:        nextCheckAt,
			},
		)
		if err != nil {
			t.Fatalf("create monitor: %v", err)
		}

		if created.ID == uuid.Nil {
			t.Fatal("created monitor ID is nil")
		}

		if created.Name != "Example API" {
			t.Fatalf(
				"created name = %q, want %q",
				created.Name,
				"Example API",
			)
		}

		if created.Method != "GET" {
			t.Fatalf(
				"created method = %q, want GET",
				created.Method,
			)
		}

		if created.Interval != 30*time.Second {
			t.Fatalf(
				"created interval = %s, want %s",
				created.Interval,
				30*time.Second,
			)
		}

		if created.Timeout != 1500*time.Millisecond {
			t.Fatalf(
				"created timeout = %s, want %s",
				created.Timeout,
				1500*time.Millisecond,
			)
		}

		if !created.NextCheckAt.Equal(nextCheckAt) {
			t.Fatalf(
				"created next_check_at = %s, want %s",
				created.NextCheckAt,
				nextCheckAt,
			)
		}

		if created.CreatedAt.IsZero() {
			t.Fatal("created_at is zero")
		}

		if created.UpdatedAt.IsZero() {
			t.Fatal("updated_at is zero")
		}

		loaded, err := repository.GetByID(
			context.Background(),
			created.ID,
		)
		if err != nil {
			t.Fatalf("get monitor by ID: %v", err)
		}

		assertMonitorEqual(t, loaded, created)
	})

	t.Run("get missing monitor", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		result, err := repository.GetByID(
			context.Background(),
			uuid.New(),
		)

		if !errors.Is(err, ErrMonitorNotFound) {
			t.Fatalf(
				"error = %v, want %v",
				err,
				ErrMonitorNotFound,
			)
		}

		if result.ID != uuid.Nil {
			t.Fatalf(
				"result ID = %s, want nil UUID",
				result.ID,
			)
		}
	})

	t.Run("list monitors", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		first, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "First",
				URL:      "https://example.com/1",
				Interval: 30 * time.Second,
				Timeout:  time.Second,
				Enabled:  true,
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		second, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "Second",
				URL:      "https://example.com/2",
				Interval: 30 * time.Second,
				Timeout:  time.Second,
				Enabled:  true,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		result, err := repository.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 2 {
			t.Fatalf("count - %d, want 2", len(result))
		}

		if result[0].ID != second.ID {
			t.Fatalf(
				"first item = %s, want %s",
				result[0].ID,
				second.ID,
			)
		}

		if result[1].ID != first.ID {
			t.Fatalf(
				"second item = %s, want %s",
				result[1].ID,
				first.ID,
			)
		}
	})

	t.Run("update monitor", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		created, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:               "Old API",
				URL:                "https://example.com/old",
				Interval:           30 * time.Second,
				Timeout:            time.Second,
				ExpectedStatusFrom: 200,
				ExpectedStatusTo:   299,
				Enabled:            true,
				NextCheckAt: time.Now().
					UTC().
					Truncate(time.Microsecond).
					Add(time.Minute),
			},
		)
		if err != nil {
			t.Fatalf("create monitor: %v", err)
		}

		originalCreatedAt := created.CreatedAt

		created.Name = "Updated API"
		created.URL = "https://example.com/new"
		created.Interval = 45 * time.Second
		created.Timeout = 2 * time.Second
		created.ExpectedStatusFrom = 201
		created.ExpectedStatusTo = 204
		created.Enabled = false
		created.UpdatedAt = time.Now().
			UTC().
			Truncate(time.Microsecond).
			Add(time.Minute)

		updated, err := repository.Update(
			context.Background(),
			created,
		)
		if err != nil {
			t.Fatalf("update monitor: %v", err)
		}

		assertMonitorEqual(t, updated, created)

		if !updated.CreatedAt.Equal(originalCreatedAt) {
			t.Fatalf(
				"CreatedAt = %s, want unchanged %s",
				updated.CreatedAt,
				originalCreatedAt,
			)
		}

		loaded, err := repository.GetByID(
			context.Background(),
			created.ID,
		)
		if err != nil {
			t.Fatalf("get updated monitor: %v", err)
		}

		assertMonitorEqual(t, loaded, updated)
	})

	t.Run("update missing monitor", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		result, err := repository.Update(
			context.Background(),
			monitor.Monitor{
				ID:                 uuid.New(),
				Name:               "Missing",
				URL:                "https://example.com",
				Method:             "GET",
				Interval:           30 * time.Second,
				Timeout:            time.Second,
				ExpectedStatusFrom: 200,
				ExpectedStatusTo:   299,
				Enabled:            true,
				NextCheckAt:        time.Now().UTC(),
				UpdatedAt:          time.Now().UTC(),
			},
		)

		if !errors.Is(err, ErrMonitorNotFound) {
			t.Fatalf("error = %v, want %v", err, result.ID)
		}
	})

	t.Run("claim due monitors", func(t *testing.T) {
		cleanMonitorsTable(t, pool)

		claimAt := time.Now().
			UTC().
			Truncate(time.Microsecond)

		oldestDue, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "Oldest Due",
				URL:      "https://example.com/oldest",
				Interval: 30 * time.Second,
				Timeout:  time.Second,
				Enabled:  true,
				NextCheckAt: claimAt.
					Add(-2 * time.Minute),
			},
		)
		if err != nil {
			t.Fatalf("create oldest due monitor: %v", err)
		}

		secondDue, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "Second Due",
				URL:      "https://example.com/second",
				Interval: 45 * time.Second,
				Timeout:  time.Second,
				Enabled:  true,
				NextCheckAt: claimAt.
					Add(-time.Minute),
			},
		)
		if err != nil {
			t.Fatalf("create second due monitor: %v", err)
		}

		future, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "Future",
				URL:      "https://example.com/future",
				Interval: 30 * time.Second,
				Timeout:  time.Second,
				Enabled:  true,
				NextCheckAt: claimAt.
					Add(time.Minute),
			},
		)
		if err != nil {
			t.Fatalf("create future monitor: %v", err)
		}

		disabled, err := repository.Create(
			context.Background(),
			monitor.NewMonitor{
				Name:     "Disabled",
				URL:      "https://example.com/disabled",
				Interval: 30 * time.Second,
				Timeout:  time.Second,
				Enabled:  false,
				NextCheckAt: claimAt.
					Add(-3 * time.Minute),
			},
		)
		if err != nil {
			t.Fatalf("create disabled monitor: %v", err)
		}

		firstBatch, err := repository.ClaimDue(
			context.Background(),
			claimAt,
			1,
		)
		if err != nil {
			t.Fatalf("claim first batch: %v", err)
		}

		if len(firstBatch) != 1 {
			t.Fatalf(
				"first batch length = %d, want 1",
				len(firstBatch),
			)
		}

		if firstBatch[0].ID != oldestDue.ID {
			t.Fatalf(
				"claimed monitor = %s, want %s",
				firstBatch[0].ID,
				oldestDue.ID,
			)
		}

		expectedOldestNextCheckAt := claimAt.Add(
			oldestDue.Interval,
		)

		if !firstBatch[0].NextCheckAt.Equal(
			expectedOldestNextCheckAt,
		) {
			t.Fatalf(
				"oldest next check at = %s, want %s",
				firstBatch[0].NextCheckAt,
				expectedOldestNextCheckAt,
			)
		}

		secondBatch, err := repository.ClaimDue(
			context.Background(),
			claimAt,
			10,
		)
		if err != nil {
			t.Fatalf("claim second batch: %v", err)
		}

		if len(secondBatch) != 1 {
			t.Fatalf(
				"second batch length = %d, want 1",
				len(secondBatch),
			)
		}

		if secondBatch[0].ID != secondDue.ID {
			t.Fatalf(
				"claimed monitor = %s, want %s",
				secondBatch[0].ID,
				secondDue.ID,
			)
		}

		expectedSecondNextCheckAt := claimAt.Add(
			secondDue.Interval,
		)

		if !secondBatch[0].NextCheckAt.Equal(
			expectedSecondNextCheckAt,
		) {
			t.Fatalf(
				"second next check at = %s, want %s",
				secondBatch[0].NextCheckAt,
				expectedSecondNextCheckAt,
			)
		}

		thirdBatch, err := repository.ClaimDue(
			context.Background(),
			claimAt,
			10,
		)
		if err != nil {
			t.Fatalf("claim third batch: %v", err)
		}

		if len(thirdBatch) != 0 {
			t.Fatalf(
				"third batch length = %d, want 0",
				len(thirdBatch),
			)
		}

		loadedFuture, err := repository.GetByID(
			context.Background(),
			future.ID,
		)
		if err != nil {
			t.Fatalf("get future monitor: %v", err)
		}

		if !loadedFuture.NextCheckAt.Equal(
			future.NextCheckAt,
		) {
			t.Fatalf(
				"future next check at = %s, want unchanged %s",
				loadedFuture.NextCheckAt,
				future.NextCheckAt,
			)
		}

		loadedDisabled, err := repository.GetByID(
			context.Background(),
			disabled.ID,
		)
		if err != nil {
			t.Fatalf("get disabled monitor: %v", err)
		}

		if !loadedDisabled.NextCheckAt.Equal(
			disabled.NextCheckAt,
		) {
			t.Fatalf(
				"disabled next check at = %s, want unchanged %s",
				loadedDisabled.NextCheckAt,
				disabled.NextCheckAt,
			)
		}
	})
}

func cleanMonitorsTable(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if _, err := pool.Exec(
		ctx,
		"TRUNCATE TABLE check_results, monitors",
	); err != nil {
		t.Fatalf("truncate monitors table: %v", err)
	}
}

func assertMonitorEqual(
	t *testing.T,
	got monitor.Monitor,
	want monitor.Monitor,
) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("ID = %s, want %s", got.ID, want.ID)
	}

	if got.Name != want.Name {
		t.Fatalf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.URL != want.URL {
		t.Fatalf("URL = %q, want %q", got.URL, want.URL)
	}

	if got.Method != want.Method {
		t.Fatalf("Method = %q, want %q", got.Method, want.Method)
	}

	if got.Interval != want.Interval {
		t.Fatalf(
			"Interval = %s, want %s",
			got.Interval,
			want.Interval,
		)
	}

	if got.Timeout != want.Timeout {
		t.Fatalf(
			"Timeout = %s, want %s",
			got.Timeout,
			want.Timeout,
		)
	}

	if got.ExpectedStatusFrom != want.ExpectedStatusFrom {
		t.Fatalf(
			"ExpectedStatusFrom = %d, want %d",
			got.ExpectedStatusFrom,
			want.ExpectedStatusFrom,
		)
	}

	if got.ExpectedStatusTo != want.ExpectedStatusTo {
		t.Fatalf(
			"ExpectedStatusTo = %d, want %d",
			got.ExpectedStatusTo,
			want.ExpectedStatusTo,
		)
	}

	if got.Enabled != want.Enabled {
		t.Fatalf(
			"Enabled = %t, want %t",
			got.Enabled,
			want.Enabled,
		)
	}

	if !got.NextCheckAt.Equal(want.NextCheckAt) {
		t.Fatalf(
			"NextCheckAt = %s, want %s",
			got.NextCheckAt,
			want.NextCheckAt,
		)
	}

	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf(
			"CreatedAt = %s, want %s",
			got.CreatedAt,
			want.CreatedAt,
		)
	}

	if !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Fatalf(
			"UpdatedAt = %s, want %s",
			got.UpdatedAt,
			want.UpdatedAt,
		)
	}
}
