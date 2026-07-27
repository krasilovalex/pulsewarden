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
