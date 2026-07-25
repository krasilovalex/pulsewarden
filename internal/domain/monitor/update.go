package monitor

import "time"

type UpdateMonitor struct {
	Name               *string
	URL                *string
	Method             *string
	Interval           *time.Duration
	Timeout            *time.Duration
	ExpectedStatusFrom *int
	ExpectedStatusTo   *int
	Enabled            *bool
}

func (update UpdateMonitor) IsEmpty() bool {
	return update.Name == nil &&
		update.URL == nil &&
		update.Method == nil &&
		update.Interval == nil &&
		update.Timeout == nil &&
		update.ExpectedStatusFrom == nil &&
		update.ExpectedStatusTo == nil &&
		update.Enabled == nil
}

func (update UpdateMonitor) Apply(target *Monitor) {
	if update.Name != nil {
		target.Name = *update.Name
	}

	if update.URL != nil {
		target.URL = *update.URL
	}

	if update.Method != nil {
		target.Method = *update.Method
	}

	if update.Interval != nil {
		target.Interval = *update.Interval
	}

	if update.Timeout != nil {
		target.Timeout = *update.Timeout
	}

	if update.ExpectedStatusFrom != nil {
		target.ExpectedStatusFrom = *update.ExpectedStatusFrom
	}

	if update.ExpectedStatusTo != nil {
		target.ExpectedStatusTo = *update.ExpectedStatusTo
	}

	if update.Enabled != nil {
		target.Enabled = *update.Enabled
	}
}
