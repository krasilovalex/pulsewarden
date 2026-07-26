package monitor

import (
	"context"
	"fmt"
	"time"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type CreatorRepository interface {
	Create(
		ctx context.Context,
		input domainmonitor.NewMonitor,
	) (domainmonitor.Monitor, error)
}

type Create struct {
	repository CreatorRepository
	now        func() time.Time
}

func NewCreate(
	repository CreatorRepository,
) *Create {
	return &Create{
		repository: repository,
		now:        time.Now,
	}
}

func (uc *Create) Execute(
	ctx context.Context,
	input domainmonitor.NewMonitor,
) (domainmonitor.Monitor, error) {
	normalized, err := domainmonitor.NormalizeAndValidate(
		input,
		uc.now(),
	)
	if err != nil {
		return domainmonitor.Monitor{}, err
	}

	created, err := uc.repository.Create(ctx, normalized)
	if err != nil {
		return domainmonitor.Monitor{}, fmt.Errorf(
			"create monitor: %w",
			err,
		)
	}

	return created, nil
}
