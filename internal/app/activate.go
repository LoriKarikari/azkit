package app

import (
	"context"
	"fmt"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

const CodeEligibleNotFound Code = "eligible_not_found"

var ErrEligibleNotFound = &Error{
	Code:    "eligible_not_found",
	Message: "No matching eligible assignment found.",
}

var ErrMissingReason = &Error{
	Code:    "missing_reason",
	Message: "Activation reason is required.",
}

type ActivationStore interface {
	Activate(ctx context.Context, principalID, roleDefID, scope, reason string, duration time.Duration) (*domain.ActivationResult, error)
}

type ActivationService struct {
	store     EligibleAssignments
	activator ActivationStore
}

func NewActivationService(store EligibleAssignments, activator ActivationStore) *ActivationService {
	return &ActivationService{store: store, activator: activator}
}

func (s *ActivationService) Activate(ctx context.Context, req domain.ActivationRequest) (*domain.ActivationResult, error) {
	if req.Reason == "" {
		return nil, ErrMissingReason
	}
	if req.Duration <= 0 {
		return nil, &Error{
			Code:    "invalid_duration",
			Message: fmt.Sprintf("Invalid duration: %s", req.Duration),
		}
	}

	eligible, err := s.store.ListEligible(ctx)
	if err != nil {
		return nil, err
	}

	var match *domain.EligibleAssignment
	for i := range eligible {
		if eligible[i].ScopeID == req.ScopeID &&
			(eligible[i].Role == req.Role || eligible[i].RoleDefID == req.Role) {
			match = &eligible[i]
			break
		}
	}
	if match == nil {
		return nil, ErrEligibleNotFound
	}

	return s.activator.Activate(ctx, match.PrincipalID, match.RoleDefID, match.ScopeID, req.Reason, req.Duration)
}
