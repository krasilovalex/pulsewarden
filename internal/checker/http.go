package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type HTTPChecker struct {
	client *http.Client
	now    func() time.Time
}

func NewHTTPChecker(client *http.Client) *HTTPChecker {
	if client == nil {
		client = &http.Client{}
	}

	return &HTTPChecker{
		client: client,
		now:    time.Now,
	}
}

func (c *HTTPChecker) Check(
	ctx context.Context,
	input monitor.Monitor,
) (checkresult.CheckResult, error) {
	checkCtx, cancel := context.WithTimeout(
		ctx,
		input.Timeout,
	)
	defer cancel()

	request, err := http.NewRequestWithContext(
		checkCtx,
		input.Method,
		input.URL,
		nil,
	)

	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"create HTTP request: %w",
			err,
		)
	}

	checkedAt := c.now().UTC()

	response, requestErr := c.client.Do(request)

	latency := c.now().Sub(checkedAt)

	if requestErr != nil {
		errorMessage := requestErr.Error()

		result, err := checkresult.New(
			input.ID,
			checkresult.StatusDown,
			nil,
			latency,
			&errorMessage,
			checkedAt,
		)
		if err != nil {
			return checkresult.CheckResult{}, fmt.Errorf(
				"build failed HTTP check result: %w",
				err,
			)
		}

		return result, nil
	}

	defer response.Body.Close()

	statusCode := response.StatusCode
	status := checkresult.StatusDown

	if statusCode >= input.ExpectedStatusFrom &&
		statusCode <= input.ExpectedStatusTo {
		status = checkresult.StatusUp
	}

	result, err := checkresult.New(
		input.ID,
		status,
		&statusCode,
		latency,
		nil,
		checkedAt,
	)

	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"build HTTP check result: %w",
			err,
		)
	}

	return result, nil
}
