package app

import (
	"context"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

const (
	CodeEligibleNotFound      Code = "eligible_not_found"
	CodeActiveNotFound        Code = "active_not_found"
	CodeMissingScope          Code = "missing_scope"
	CodeMissingRole           Code = "missing_role"
	CodeMissingReason         Code = "missing_reason"
	CodeInvalidDuration       Code = "invalid_duration"
	CodeConflictingSelectors  Code = "conflicting_selectors"
	CodeUnknownSubscription   Code = "unknown_subscription"
	CodeAmbiguousSubscription Code = "ambiguous_subscription"
	CodeUnknownResourceGroup  Code = "unknown_resource_group"
)

var (
	ErrEligibleNotFound     = &Error{Code: CodeEligibleNotFound, Message: "No matching eligible assignment found."}
	ErrActiveNotFound       = &Error{Code: CodeActiveNotFound, Message: "No matching active assignment found."}
	ErrMissingScope         = &Error{Code: CodeMissingScope, Message: "Activation scope is required."}
	ErrMissingRole          = &Error{Code: CodeMissingRole, Message: "Activation role is required."}
	ErrMissingReason        = &Error{Code: CodeMissingReason, Message: "Activation reason is required."}
	ErrConflictingSelectors = &Error{Code: CodeConflictingSelectors, Message: "Use --scope or --subscription, not both."}
)

type ActivationStore interface {
	Activate(context.Context, domain.ActivationTarget) (*domain.ActivationResult, error)
}

type ActivationService struct {
	store     EligibleAssignments
	activator ActivationStore
	resolver  activationResolver
}

func NewActivationService(store EligibleAssignments, activator ActivationStore) *ActivationService {
	return &ActivationService{store: store, activator: activator, resolver: activationResolver{}}
}

func (s *ActivationService) ActivateResolved(
	ctx context.Context,
	target domain.ActivationTarget,
) (*domain.ActivationResult, error) {
	target.Reason = strings.TrimSpace(target.Reason)
	if target.Reason == "" {
		return nil, ErrMissingReason
	}
	if target.Duration <= 0 {
		return nil, &Error{
			Code:    CodeInvalidDuration,
			Message: "Invalid activation duration.",
		}
	}
	return s.activator.Activate(ctx, target)
}

func (s *ActivationService) Activate(
	ctx context.Context,
	req domain.ActivationRequest,
) (*domain.ActivationResult, error) {
	if err := validateActivation(req); err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)

	eligible, err := s.store.ListEligible(ctx)
	if err != nil {
		return nil, err
	}

	target, err := s.resolver.Resolve(req, eligible)
	if err != nil {
		return nil, err
	}
	return s.activator.Activate(ctx, target)
}
