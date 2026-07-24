package monitor

import (
	"context"

	domainmonitor "github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

type ListerRepository interface {
	List(context.Context) ([]domainmonitor.Monitor, error)
}

type List struct {
	repository ListerRepository
}

func NewList(repository ListerRepository) *List {
	return &List{
		repository: repository,
	}
}

func (uc *List) Execute(
	ctx context.Context,
) ([]domainmonitor.Monitor, error) {
	return uc.repository.List(ctx)
}
