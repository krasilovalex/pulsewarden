package monitor

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	ID                 uuid.UUID
	Name               string
	URL                string
	Method             string
	Interval           time.Duration
	Timeout            time.Duration
	ExpectedStatusFrom int
	ExpectedStatusTo   int
	Enabled            bool
	NextCheckAt        time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type NewMonitor struct {
	Name               string
	URL                string
	Method             string
	Interval           time.Duration
	Timeout            time.Duration
	ExpectedStatusFrom int
	ExpectedStatusTo   int
	Enabled            bool
	NextCheckAt        time.Time
}

func (m NewMonitor) WithDefaults() NewMonitor {
	if m.Method == "" {
		m.Method = http.MethodGet
	}

	if m.ExpectedStatusFrom == 0 {
		m.ExpectedStatusFrom = http.StatusOK
	}

	if m.ExpectedStatusTo == 0 {
		m.ExpectedStatusTo = 299
	}

	return m
}
