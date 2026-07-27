package checkresult

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

type CheckResult struct {
	ID         uuid.UUID
	MonitorID  uuid.UUID
	Status     Status
	StatusCode *int
	Latency    time.Duration
	Error      *string
	CheckedAt  time.Time
}

func (result CheckResult) IsUp() bool {
	return result.Status == StatusUp
}
