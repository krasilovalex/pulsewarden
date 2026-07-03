package api

import (
	"net/http"
	"time"
)

type ServerConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func NewServer(cfg ServerConfig) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readinessHandler)

	return &http.Server{
		Addr:              cfg.Address,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w, http.StatusOK, `{"status":"ok"}`)
}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w, http.StatusOK, `{"status":"ready"}`)
}

func writeStatus(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write([]byte(body + "\n"))
}
