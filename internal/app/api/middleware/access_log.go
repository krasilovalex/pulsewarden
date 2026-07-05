package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type clock func() time.Time

func AccessLog(log *slog.Logger, next http.Handler) http.Handler {
	return accessLog(log, next, time.Now)
}

func accessLog(
	log *slog.Logger,
	next http.Handler,
	now clock,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := now()

		tracker := trackResponse(w)

		next.ServeHTTP(tracker, r)

		requestID, _ := RequestIDFromContext(r.Context())

		duration := now().Sub(startedAt)

		log.InfoContext(
			r.Context(),
			"HTTP request completed",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", tracker.StatusCode()),
			slog.Int("response_bytes", tracker.BytesWritten()),
			slog.Float64(
				"duration_ms",
				float64(duration.Microseconds())/1000,
			),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}
