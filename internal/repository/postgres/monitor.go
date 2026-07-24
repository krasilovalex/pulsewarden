package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

var ErrMonitorNotFound = errors.New("monitor not found")

type MonitorRepository struct {
	pool    *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewMonitorRepository(pool *pgxpool.Pool) *MonitorRepository {
	return &MonitorRepository{
		pool: pool,
		builder: sq.StatementBuilder.
			PlaceholderFormat(sq.Dollar),
	}
}

func (r *MonitorRepository) Create(ctx context.Context, input monitor.NewMonitor) (monitor.Monitor, error) {
	input = input.WithDefaults()

	id := uuid.New()
	now := time.Now().UTC()

	query, args, err := r.builder.
		Insert("monitors").
		Columns(
			"id",
			"name",
			"url",
			"method",
			"interval_seconds",
			"timeout_milliseconds",
			"expected_status_from",
			"expected_status_to",
			"enabled",
			"next_check_at",
			"created_at",
			"updated_at",
		).
		Values(
			id,
			input.Name,
			input.URL,
			input.Method,
			durationSeconds(input.Interval),
			durationMilliseconds(input.Timeout),
			input.ExpectedStatusFrom,
			input.ExpectedStatusTo,
			input.Enabled,
			input.NextCheckAt.UTC(),
			now,
			now,
		).
		Suffix(`
		RETURNING
			id,
			name,
			url,
			method,
			interval_seconds,
			timeout_milliseconds,
			expected_status_from,
			expected_status_to,
			enabled,
			next_check_at,
			created_at,
			updated_at`).
		ToSql()

	if err != nil {
		return monitor.Monitor{}, fmt.Errorf("build create monitor query: %w", err)
	}

	result, err := scanMonitor(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		return monitor.Monitor{}, fmt.Errorf("create monitor: %w", err)
	}

	return result, nil
}

func (r *MonitorRepository) GetByID(ctx context.Context, id uuid.UUID) (monitor.Monitor, error) {
	query, args, err := r.builder.
		Select(
			"id",
			"name",
			"url",
			"method",
			"interval_seconds",
			"timeout_milliseconds",
			"expected_status_from",
			"expected_status_to",
			"enabled",
			"next_check_at",
			"created_at",
			"updated_at",
		).
		From("monitors").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return monitor.Monitor{}, fmt.Errorf("build get monitor query: %w", err)
	}

	result, err := scanMonitor(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return monitor.Monitor{}, ErrMonitorNotFound
	}
	if err != nil {
		return monitor.Monitor{}, fmt.Errorf("get monitor: %w", err)
	}

	return result, nil
}

func (r *MonitorRepository) List(ctx context.Context) ([]monitor.Monitor, error) {
	query, args, err := r.builder.
		Select(
			"id",
			"name",
			"url",
			"method",
			"interval_seconds",
			"timeout_milliseconds",
			"expected_status_from",
			"expected_status_to",
			"enabled",
			"next_check_at",
			"created_at",
			"updated_at",
		).
		From("monitors").
		OrderBy("created_at DESC", "id DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf(
			"build list monitors query: %w",
			err,
		)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"list monitors: %w",
			err,
		)
	}
	defer rows.Close()

	monitors := make([]monitor.Monitor, 0)

	for rows.Next() {
		item, err := scanMonitor(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan monitor: %w",
				err,
			)
		}

		monitors = append(monitors, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate monitors: %w",
			err,
		)
	}

	return monitors, nil

}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMonitor(row rowScanner) (monitor.Monitor, error) {
	var (
		result              monitor.Monitor
		intervalSeconds     int64
		timeoutMilliseconds int64
	)

	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.URL,
		&result.Method,
		&intervalSeconds,
		&timeoutMilliseconds,
		&result.ExpectedStatusFrom,
		&result.ExpectedStatusTo,
		&result.Enabled,
		&result.NextCheckAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return monitor.Monitor{}, err
	}

	result.Interval = time.Duration(intervalSeconds) * time.Second
	result.Timeout = time.Duration(timeoutMilliseconds) * time.Millisecond

	return result, nil
}

func durationSeconds(value time.Duration) int64 {
	return int64(value / time.Second)
}

func durationMilliseconds(value time.Duration) int64 {
	return int64(value / time.Millisecond)
}
