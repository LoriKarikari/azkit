package app

import (
	"context"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type EligibleAssignments interface {
	ListEligible(context.Context) ([]domain.EligibleAssignment, error)
}

type ListService struct {
	store EligibleAssignments
}

func NewListService(store EligibleAssignments) *ListService {
	return &ListService{store: store}
}

func (s *ListService) List(ctx context.Context) ([]domain.EligibleAssignment, error) {
	return s.store.ListEligible(ctx)
}
