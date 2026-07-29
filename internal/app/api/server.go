package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/wayzzoo/pulsewarden/internal/app/api/middleware"
	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type ReadinessChecker interface {
	Ping(context.Context) error
}

type MonitorCreator interface {
	Execute(
		context.Context,
		domainmonitor.NewMonitor,
	) (domainmonitor.Monitor, error)
}

type ServerConfig struct {
	Address              string
	ReadHeaderTimeout    time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	ReadinessTimeout     time.Duration
	Logger               *slog.Logger
	Postgres             ReadinessChecker
	MonitorCreator       MonitorCreator
	MonitorLister        MonitorLister
	MonitorGetter        MonitorGetter
	MonitorUpdater       MonitorUpdater
	MonitorResultsLister MonitorResultsLister
}

type MonitorLister interface {
	Execute(context.Context) ([]domainmonitor.Monitor, error)
}

func NewServer(cfg ServerConfig) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc(
		"GET /readyz",
		readinessHandler(cfg.Postgres, cfg.ReadinessTimeout),
	)
	mux.HandleFunc("POST /api/v1/monitors", createMonitorHandler(cfg.Logger, cfg.MonitorCreator))
	mux.HandleFunc(
		"GET /api/v1/monitors",
		listMonitorsHandler(cfg.Logger, cfg.MonitorLister),
	)

	mux.HandleFunc(
		"GET /api/v1/monitors/{id}",
		getMonitorHandler(cfg.Logger, cfg.MonitorGetter),
	)

	mux.HandleFunc(
		"PATCH /api/v1/monitors/{id}",
		updateMonitorHandler(cfg.Logger, cfg.MonitorUpdater),
	)

	mux.HandleFunc(
		"GET /api/v1/monitors/{id}/results",
		listMonitorResultsHandler(
			cfg.Logger,
			cfg.MonitorResultsLister,
		),
	)

	var handler http.Handler = mux

	handler = middleware.Recovery(cfg.Logger, handler)
	handler = middleware.AccessLog(cfg.Logger, handler)
	handler = middleware.RequestID(handler)

	return &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w, http.StatusOK, `{"status":"ok"}`)
}

func readinessHandler(checker ReadinessChecker, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if checker == nil {
			writeStatus(w, http.StatusServiceUnavailable, `{"status":"unavailable"}`)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			writeStatus(w, http.StatusServiceUnavailable, `{"status":"unavailable"}`)
			return
		}

		writeStatus(w, http.StatusOK, `{"status":"ready"}`)
	}
}

func writeStatus(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write([]byte(body + "\n"))
}
