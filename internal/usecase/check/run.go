package check

import (
	"context"
	"fmt"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type Checker interface {
	Check(
		ctx context.Context,
		input monitor.Monitor,
	) (checkresult.CheckResult, error)
}

type ResultRepository interface {
	Create(
		ctx context.Context,
		input checkresult.CheckResult,
	) (checkresult.CheckResult, error)
}

type Run struct {
	checker    Checker
	repository ResultRepository
}

func NewRun(
	checker Checker,
	repository ResultRepository,
) *Run {
	return &Run{checker: checker, repository: repository}
}

func (uc *Run) Execute(
	ctx context.Context,
	input monitor.Monitor,
) (checkresult.CheckResult, error) {
	result, err := uc.checker.Check(ctx, input)
	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"check monitor: %w",
			err,
		)
	}

	created, err := uc.repository.Create(ctx, result)
	if err != nil {
		return checkresult.CheckResult{}, fmt.Errorf(
			"save check result: %w",
			err,
		)
	}

	return created, nil
}
