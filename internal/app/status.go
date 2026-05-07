package app

import (
	"context"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActiveAssignments interface {
	ListActive(context.Context) ([]domain.ActiveAssignment, error)
}

type StatusService struct {
	store ActiveAssignments
}

func NewStatusService(store ActiveAssignments) *StatusService {
	return &StatusService{store: store}
}

func (s *StatusService) Status(ctx context.Context) ([]domain.ActiveAssignment, error) {
	return s.store.ListActive(ctx)
}
