package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

const (
	CodeEligibleNotFound Code = "eligible_not_found"
	CodeMissingScope     Code = "missing_scope"
	CodeMissingRole      Code = "missing_role"
	CodeMissingReason    Code = "missing_reason"
	CodeInvalidDuration  Code = "invalid_duration"
)

var ErrEligibleNotFound = &Error{
	Code:    CodeEligibleNotFound,
	Message: "No matching eligible assignment found.",
}

var ErrMissingScope = &Error{
	Code:    CodeMissingScope,
	Message: "Activation scope is required.",
}

var ErrMissingRole = &Error{
	Code:    CodeMissingRole,
	Message: "Activation role is required.",
}

var ErrMissingReason = &Error{
	Code:    CodeMissingReason,
	Message: "Activation reason is required.",
}

type ActivationStore interface {
	Activate(context.Context, domain.ActivationTarget) (*domain.ActivationResult, error)
}

type ActivationService struct {
	store     EligibleAssignments
	activator ActivationStore
}

func NewActivationService(store EligibleAssignments, activator ActivationStore) *ActivationService {
	return &ActivationService{store: store, activator: activator}
}

func (s *ActivationService) Activate(ctx context.Context, req domain.ActivationRequest) (*domain.ActivationResult, error) {
	if err := validateActivation(req); err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)

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

	return s.activator.Activate(ctx, domain.ActivationTarget{
		Assignment: *match,
		Reason:     req.Reason,
		Duration:   req.Duration,
	})
}

func validateActivation(req domain.ActivationRequest) error {
	if strings.TrimSpace(req.ScopeID) == "" {
		return ErrMissingScope
	}
	if strings.TrimSpace(req.Role) == "" {
		return ErrMissingRole
	}
	if strings.TrimSpace(req.Reason) == "" {
		return ErrMissingReason
	}
	if req.Duration <= 0 {
		return &Error{
			Code:    CodeInvalidDuration,
			Message: fmt.Sprintf("Invalid activation duration: %s.", req.Duration),
		}
	}
	return nil
}
