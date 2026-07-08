package monitor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MinInterval = 10 * time.Second
	MaxInterval = 24 * time.Hour

	MinTimeout = 100 * time.Millisecond
	MaxTimeout = 30 * time.Second

	MaxNameLength = 200
	MaxURLLength  = 2048
)

var ErrInvalidMonitor = errors.New("invalid monitor")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidMonitor
}

func NormalizeAndValidate(
	input NewMonitor,
	now time.Time,
) (NewMonitor, error) {
	input = input.WithDefaults()

	input.Name = strings.TrimSpace(input.Name)
	input.URL = strings.TrimSpace(input.URL)
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))

	if input.NextCheckAt.IsZero() {
		input.NextCheckAt = now.UTC()
	} else {
		input.NextCheckAt = input.NextCheckAt.UTC()
	}

	if err := validateName(input.Name); err != nil {
		return NewMonitor{}, err
	}

	if err := validateURL(input.URL); err != nil {
		return NewMonitor{}, err
	}

	if input.Method != http.MethodGet {
		return NewMonitor{}, validationError(
			"method",
			"only GET is supported",
		)
	}

	if input.Interval < MinInterval || input.Interval > MaxInterval {
		return NewMonitor{}, validationError(
			"interval",
			fmt.Sprintf(
				"must be between %s and %s",
				MinInterval,
				MaxInterval,
			),
		)
	}

	if input.Timeout < MinTimeout || input.Timeout > MaxTimeout {
		return NewMonitor{}, validationError(
			"timeout",
			fmt.Sprintf(
				"must be between %s and %s",
				MinTimeout,
				MaxTimeout,
			),
		)
	}

	if input.Timeout >= input.Interval {
		return NewMonitor{}, validationError(
			"timeout",
			"must be less than interval",
		)
	}

	if input.ExpectedStatusFrom < 100 ||
		input.ExpectedStatusFrom > 599 {
		return NewMonitor{}, validationError(
			"expected_status_from",
			"must be between 100 and 599",
		)
	}

	if input.ExpectedStatusTo < 100 ||
		input.ExpectedStatusTo > 599 {
		return NewMonitor{}, validationError(
			"expected_status_to",
			"must be between 100 and 599",
		)
	}

	if input.ExpectedStatusFrom > input.ExpectedStatusTo {
		return NewMonitor{}, validationError(
			"expected_status_to",
			"must be greater than or equal to expected_status_from",
		)
	}

	return input, nil
}

func validateName(name string) error {
	if name == "" {
		return validationError("name", "must not be blank")
	}

	if len(name) > MaxNameLength {
		return validationError(
			"name",
			fmt.Sprintf(
				"must not exceed %d bytes",
				MaxNameLength,
			),
		)
	}

	return nil
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return validationError("url", "must not be blank")
	}

	if len(rawURL) > MaxURLLength {
		return validationError(
			"url",
			fmt.Sprintf(
				"must not exceed %d bytes",
				MaxURLLength,
			),
		)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return validationError("url", "must be a valid URL")
	}

	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		return validationError(
			"url",
			"scheme must be http or https",
		)
	}

	if parsedURL.Opaque != "" {
		return validationError(
			"url",
			"must be an absolute hierarchical URL",
		)
	}

	if parsedURL.Host == "" || parsedURL.Hostname() == "" {
		return validationError(
			"url",
			"host must not be empty",
		)
	}

	if parsedURL.User != nil {
		return validationError(
			"url",
			"embedded credentials are not allowed",
		)
	}

	if parsedURL.Fragment != "" {
		return validationError(
			"url",
			"fragment is not allowed",
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
