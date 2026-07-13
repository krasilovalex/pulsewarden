package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	domainmonitor "github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

const maxRequestBodyBytes = 1 << 20

type createMonitorRequest struct {
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Method              string `json:"method"`
	IntervalSeconds     int64  `json:"interval_seconds"`
	TimeoutMilliseconds int64  `json:"timeout_milliseconds"`
	ExpectedStatusFrom  int    `json:"expected_status_from"`
	ExpectedStatusTo    int    `json:"expected_status_to"`
	Enabled             *bool  `json:"enabled"`
}

func (r createMonitorRequest) toDomain() domainmonitor.NewMonitor {
	enabled := true

	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	return domainmonitor.NewMonitor{
		Name:               r.Name,
		URL:                r.URL,
		Method:             r.Method,
		Interval:           time.Duration(r.IntervalSeconds) * time.Second,
		Timeout:            time.Duration(r.TimeoutMilliseconds) * time.Millisecond,
		ExpectedStatusFrom: r.ExpectedStatusFrom,
		ExpectedStatusTo:   r.ExpectedStatusTo,
		Enabled:            enabled,
	}
}

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode JSON body: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
