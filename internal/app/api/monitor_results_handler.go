package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wayzzoo/pulsewarden/internal/app/api/middleware"
	"github.com/wayzzoo/pulsewarden/internal/app/api/response"
	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

const (
	defaultMonitorResultsLimit = 100
	maxMonitorResultLimit      = 500
)

type MonitorResultsLister interface {
	Execute(
		ctx context.Context,
		monitorID uuid.UUID,
		limit int,
	) ([]checkresult.CheckResult, error)
}

type listMonitorResultsResponse struct {
	Items []checkResultResponse `json:"items"`
}

type checkResultResponse struct {
	ID                  string  `json:"id"`
	MonitorID           string  `json:"monitor_id"`
	Status              string  `json:"status"`
	StatusCode          *int    `json:"status_code"`
	LatencyMilliseconds int64   `json:"latency_milliseconds"`
	ErrorMessage        *string `json:"error_message"`
	CheckedAt           string  `json:"checked_at"`
}

func listMonitorResultsHandler(
	logger *slog.Logger,
	lister MonitorResultsLister,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ :=
			middleware.RequestIDFromContext(r.Context())

		monitorID, err := uuid.Parse(
			r.PathValue("id"),
		)
		if err != nil {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_monitor_id",
				"monitor id must be a valid UUID",
				requestID,
			)
			return
		}

		limit, err := parseMonitorResultsLimit(r)
		if err != nil {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_limit",
				err.Error(),
				requestID,
			)
			return
		}

		results, err := lister.Execute(
			r.Context(),
			monitorID,
			limit,
		)
		if err != nil {
			if errors.Is(
				err,
				repositorypostgres.ErrMonitorNotFound,
			) {
				_ = response.WriteError(
					w,
					http.StatusNotFound,
					"monitor_not_found",
					"monitor not found",
					requestID,
				)
				return
			}

			logger.Error(
				"list monitor results failed",
				slog.Any("error", err),
				slog.String(
					"monitor_id",
					monitorID.String(),
				),
				slog.String(
					"request_id",
					requestID,
				),
			)

			_ = response.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				"internal server error",
				requestID,
			)
			return
		}

		items := make(
			[]checkResultResponse,
			0,
			len(results),
		)

		for _, result := range results {
			items = append(items, checkResultToResponse(result))
		}

		_ = response.WriteJSON(
			w,
			http.StatusOK,
			listMonitorResultsResponse{
				Items: items,
			},
		)
	}
}

func parseMonitorResultsLimit(
	r *http.Request,
) (int, error) {
	rawLimit := r.URL.Query().Get("limit")
	if rawLimit == "" {
		return defaultMonitorResultsLimit, nil
	}

	limit, err := strconv.Atoi(rawLimit)
	if err != nil ||
		limit <= 0 ||
		limit > maxMonitorResultLimit {
		return 0, fmt.Errorf(
			"limit must be an integer betweem 1 and %d",
			maxMonitorResultLimit,
		)
	}

	return limit, nil
}

func checkResultToResponse(
	result checkresult.CheckResult,
) checkResultResponse {
	return checkResultResponse{
		ID:                  result.ID.String(),
		MonitorID:           result.MonitorID.String(),
		Status:              string(result.Status),
		StatusCode:          result.StatusCode,
		LatencyMilliseconds: result.Latency.Milliseconds(),
		ErrorMessage:        result.Error,
		CheckedAt:           result.CheckedAt.UTC().Format(time.RFC3339Nano),
	}
}
