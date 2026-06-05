package app

import (
	"context"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type ActiveAssignments interface {
	ListActive(context.Context) ([]domain.ActiveAssignment, error)
}

type StatusService struct {
	store  ActiveAssignments
	scoped ActiveAssignmentLookup
}

func NewStatusService(store ActiveAssignments) *StatusService {
	svc := &StatusService{store: store}
	if scoped, ok := store.(ActiveAssignmentLookup); ok {
		svc.scoped = scoped
	}
	return svc
}

func (s *StatusService) Status(ctx context.Context) ([]domain.ActiveAssignment, error) {
	return s.store.ListActive(ctx)
}

func (s *StatusService) StatusForScope(ctx context.Context, scope string) ([]domain.ActiveAssignment, error) {
	if s.scoped != nil {
		return s.scoped.ListActiveForScope(ctx, scope)
	}
	assignments, err := s.store.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	matched := []domain.ActiveAssignment{}
	for _, assignment := range assignments {
		if assignment.ScopeID == scope {
			matched = append(matched, assignment)
		}
	}
	return matched, nil
}
