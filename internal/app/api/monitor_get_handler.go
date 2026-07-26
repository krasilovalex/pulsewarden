package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/wayzzoo/pulsewarden/internal/app/api/middleware"
	"github.com/wayzzoo/pulsewarden/internal/app/api/response"
	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

type MonitorGetter interface {
	Execute(
		context.Context,
		uuid.UUID,
	) (domainmonitor.Monitor, error)
}

func getMonitorHandler(
	logger *slog.Logger,
	getter MonitorGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := middleware.RequestIDFromContext(r.Context())

		id, err := uuid.Parse(r.PathValue("id"))

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

		item, err := getter.Execute(r.Context(), id)
		if err != nil {
			if errors.Is(err, repositorypostgres.ErrMonitorNotFound) {
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
				"get monitor failed",
				slog.Any("error", err),
				slog.String("monitor_id", id.String()),
				slog.String("request_id", requestID),
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

		_ = response.WriteJSON(
			w,
			http.StatusOK,
			monitorToResponse(item),
		)
	}
}
