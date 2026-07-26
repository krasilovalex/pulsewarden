package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainmonitor "github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

type UpdaterRepository interface {
	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domainmonitor.Monitor, error)

	Update(
		ctx context.Context,
		input domainmonitor.Monitor,
	) (domainmonitor.Monitor, error)
}

type Update struct {
	repository UpdaterRepository
	now        func() time.Time
}

func NewUpdate(repository UpdaterRepository) *Update {
	return &Update{
		repository: repository,
		now:        time.Now,
	}
}

func (uc *Update) Execute(
	ctx context.Context,
	id uuid.UUID,
	update domainmonitor.UpdateMonitor,
) (domainmonitor.Monitor, error) {
	current, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return domainmonitor.Monitor{}, fmt.Errorf(
			"get monitor for update: %w",
			err,
		)
	}

	normalized, err := domainmonitor.NormalizeAndValidateUpdate(
		current,
		update,
		uc.now(),
	)
	if err != nil {
		return domainmonitor.Monitor{}, fmt.Errorf(
			"validate monitor update: %w",
			err,
		)
	}

	result, err := uc.repository.Update(ctx, normalized)
	if err != nil {
		return domainmonitor.Monitor{}, fmt.Errorf(
			"update monitor: %w",
			err,
		)
	}

	return result, nil
}
