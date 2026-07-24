package api

import (
	"log/slog"
	"net/http"

	"github.com/krasilovalex/pulsewarden/internal/app/api/middleware"
	"github.com/krasilovalex/pulsewarden/internal/app/api/response"
)

type listMonitorsResponse struct {
	Items []createMonitorResponse `json:"items"`
}

func listMonitorsHandler(
	logger *slog.Logger,
	lister MonitorLister,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := middleware.RequestIDFromContext(r.Context())

		monitors, err := lister.Execute(r.Context())
		if err != nil {
			logger.Error(
				"list monitors failed",
				slog.Any("error", err),
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

		items := make([]createMonitorResponse, 0, len(monitors))

		for _, item := range monitors {
			items = append(items, monitorToResponse(item))
		}

		_ = response.WriteJSON(
			w,
			http.StatusOK,
			listMonitorsResponse{
				Items: items,
			},
		)
	}
}
