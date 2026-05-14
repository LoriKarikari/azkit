package app

import (
	"context"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type DeactivationStore interface {
	Deactivate(context.Context, domain.ActiveAssignment, string) (*domain.DeactivationResult, error)
}

type DeactivationService struct {
	active ActiveAssignments
	store  DeactivationStore
}

func NewDeactivationService(active ActiveAssignments, store DeactivationStore) *DeactivationService {
	return &DeactivationService{active: active, store: store}
}

func (s *DeactivationService) Deactivate(
	ctx context.Context,
	assignmentID string,
	reason string,
) (*domain.DeactivationResult, error) {
	reason = strings.TrimSpace(reason)

	assignments, err := s.active.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	var found *domain.ActiveAssignment
	for i := range assignments {
		if assignments[i].ID == assignmentID {
			found = &assignments[i]
			break
		}
	}
	if found == nil {
		return nil, ErrActiveNotFound
	}

	return s.store.Deactivate(ctx, *found, reason)
}
