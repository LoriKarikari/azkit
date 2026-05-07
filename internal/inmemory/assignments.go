package inmemory

import (
	"context"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type EligibleAssignments struct {
	Assignments []domain.EligibleAssignment
	Err         error
}

func (s *EligibleAssignments) ListEligible(_ context.Context) ([]domain.EligibleAssignment, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Assignments, nil
}
