package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
)

type CheckResultRepository struct {
	pool    *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewCheckResultRepository(
	pool *pgxpool.Pool,
) *CheckResultRepository {
	return &CheckResultRepository{
		pool: pool,
		builder: sq.StatementBuilder.
			PlaceholderFormat(sq.Dollar),
	}
}

func (r *CheckResultRepository) Create(
	ctx context.Context,
	input checkresult.CheckResult,
) (checkresult.CheckResult, error) {
	id := uuid.New()

	query, args, err := r.builder.
		Insert("check_results").
		Columns(
			"id",
			"monitor_id",
			"status",
			"status_code",
			"latency_milliseconds",
			"error_message",
			"checked_at",
		).
		Values(
			id,
			input.MonitorID,
			input.Status,
			input.StatusCode,
			input.Latency.Milliseconds(),
			input.Error,
			input.CheckedAt.UTC(),
		).
		Suffix(`
		RETURNING
			id,
			monitor_id,
			status,
			status_code,
			latency_milliseconds,
			error_message,
			checked_at`).
		ToSql()
	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"build create check result query: %w",
			err,
		)
	}

	result, err := scanCheckResult(
		r.pool.QueryRow(ctx, query, args...),
	)
	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"create check result: %w",
			err,
		)
	}

	return result, nil
}

func scanCheckResult(
	row rowScanner,
) (checkresult.CheckResult, error) {
	var (
		result              checkresult.CheckResult
		latencyMilliseconds int64
	)

	err := row.Scan(
		&result.ID,
		&result.MonitorID,
		&result.Status,
		&result.StatusCode,
		&latencyMilliseconds,
		&result.Error,
		&result.CheckedAt,
	)
	if err != nil {
		return checkresult.CheckResult{}, err
	}

	result.Latency =
		time.Duration(latencyMilliseconds) * time.Millisecond

	result.CheckedAt = result.CheckedAt.UTC()

	return result, nil
}

func (r *CheckResultRepository) ListByMonitorID(
	ctx context.Context,
	monitorID uuid.UUID,
	limit int,
) ([]checkresult.CheckResult, error) {
	if limit <= 0 {
		return nil, fmt.Errorf(
			"list check results: limit must be greater than zero",
		)
	}
	query, args, err := r.builder.
		Select(
			"id",
			"monitor_id",
			"status",
			"status_code",
			"latency_milliseconds",
			"error_message",
			"checked_at",
		).
		From("check_results").
		Where(sq.Eq{
			"monitor_id": monitorID,
		}).
		OrderBy(
			"checked_at DESC",
			"id DESC",
		).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"build list check results query: %w",
			err,
		)
	}

	rows, err := r.pool.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list check results: %w",
			err,
		)
	}
	defer rows.Close()

	results := make(
		[]checkresult.CheckResult,
		0,
		limit,
	)

	for rows.Next() {
		result, err := scanCheckResult(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan check result: %w",
				err,
			)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate check results: %w",
			err,
		)
	}

	return results, nil

}
