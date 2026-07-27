package checkresult

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidCheckResult = errors.New("invalid check result")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s:%s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidCheckResult
}

func New(
	monitorID uuid.UUID,
	status Status,
	statusCode *int,
	latency time.Duration,
	errorMessage *string,
	checkedAt time.Time,
) (CheckResult, error) {
	result := CheckResult{
		MonitorID:  monitorID,
		Status:     status,
		StatusCode: statusCode,
		Latency:    latency,
		Error:      errorMessage,
		CheckedAt:  checkedAt.UTC(),
	}

	if err := validate(result); err != nil {
		return CheckResult{}, err
	}

	return result, nil
}

func validate(result CheckResult) error {
	if result.MonitorID == uuid.Nil {
		return validationError(
			"monitor_id",
			"must not be nil",
		)
	}

	if result.Status != StatusUp &&
		result.Status != StatusDown {
		return validationError(
			"status",
			"must be up or down",
		)
	}

	if result.Latency < 0 {
		return validationError(
			"latency",
			"must not be negative",
		)
	}

	if result.StatusCode != nil {
		if *result.StatusCode < 100 || *result.StatusCode > 599 {
			return validationError(
				"status_code",
				"must be between 100 and 599",
			)
		}
	}

	if result.Status == StatusUp && result.StatusCode == nil {
		return validationError(
			"status_code",
			"is required when status is up",
		)
	}

	if result.Status == StatusUp && result.Error != nil {
		return validationError(
			"error",
			"must be empty when status is up",
		)
	}

	if result.CheckedAt.IsZero() {
		return validationError(
			"checked_at",
			"must not be zero",
		)
	}

	if result.Status == StatusDown && result.StatusCode == nil && result.Error == nil {
		return validationError(
			"status",
			"down result must contain status code or error",
		)
	}

	return nil

}

func validationError(
	field string,
	message string,
) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
