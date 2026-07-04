package app

import (
	"context"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type ActiveAssignments interface {
	ListActive(context.Context) ([]domain.ActiveAssignment, error)
}

type ActiveAssignmentStore interface {
	ActiveAssignments
	ActiveAssignmentLookup
}

type StatusService struct {
	store ActiveAssignmentStore
}

func NewStatusService(store ActiveAssignmentStore) *StatusService {
	return &StatusService{store: store}
}

func (s *StatusService) Status(ctx context.Context) ([]domain.ActiveAssignment, error) {
	return s.store.ListActive(ctx)
}

func (s *StatusService) StatusForScope(ctx context.Context, scope string) ([]domain.ActiveAssignment, error) {
	return s.store.ListActiveForScope(ctx, scope)
}
