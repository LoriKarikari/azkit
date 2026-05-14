package app

import (
	"context"
	"strings"

	"github.com/samber/lo"

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

	found, ok := lo.Find(assignments, func(a domain.ActiveAssignment) bool {
		return a.ID == assignmentID
	})
	if !ok {
		return nil, ErrActiveAssignmentNotFound
	}

	return s.store.Deactivate(ctx, found, reason)
}
