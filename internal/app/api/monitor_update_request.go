package api

import (
	"time"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type updateMonitorRequest struct {
	Name                *string `json:"name"`
	URL                 *string `json:"url"`
	Method              *string `json:"method"`
	IntervalSeconds     *int64  `json:"interval_seconds"`
	TimeoutMilliseconds *int64  `json:"timeout_milliseconds"`
	ExpectedStatusFrom  *int    `json:"expected_status_from"`
	ExpectedStatusTo    *int    `json:"expected_status_to"`
	Enabled             *bool   `json:"enabled"`
}

func (request updateMonitorRequest) toDomain() domainmonitor.UpdateMonitor {
	result := domainmonitor.UpdateMonitor{
		Name:               request.Name,
		URL:                request.URL,
		Method:             request.Method,
		ExpectedStatusFrom: request.ExpectedStatusFrom,
		ExpectedStatusTo:   request.ExpectedStatusTo,
		Enabled:            request.Enabled,
	}

	if request.IntervalSeconds != nil {
		value := time.Duration(*request.IntervalSeconds) * time.Second
		result.Interval = &value
	}

	if request.TimeoutMilliseconds != nil {
		value := time.Duration(*request.TimeoutMilliseconds) * time.Millisecond
		result.Timeout = &value
	}

	return result
}
