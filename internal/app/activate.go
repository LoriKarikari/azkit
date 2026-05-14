package app

import (
	"context"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

var (
	ErrEligibleNotFound         = &Error{Code: domain.CodeEligibleNotFound, Message: "No matching eligible assignment found."}
	ErrActiveAssignmentNotFound = &Error{Code: domain.CodeActiveAssignmentNotFound, Message: "No matching active assignment found."}
	ErrMissingScope             = &Error{Code: domain.CodeMissingScope, Message: "Activation scope is required."}
	ErrMissingRole              = &Error{Code: domain.CodeMissingRole, Message: "Activation role is required."}
	ErrMissingReason            = &Error{Code: domain.CodeMissingReason, Message: "Activation reason is required."}
	ErrConflictingSelectors     = &Error{Code: domain.CodeConflictingSelectors, Message: "Use --scope or --subscription, not both."}
)

type ActivationStore interface {
	Activate(context.Context, domain.ActivationTarget) (*domain.ActivationResult, error)
}

type ActiveAssignmentLookup interface {
	ListActiveForScope(context.Context, string) ([]domain.ActiveAssignment, error)
}

type ActivationService struct {
	store     EligibleAssignments
	active    ActiveAssignmentLookup
	activator ActivationStore
	resolver  activationResolver
}

func NewActivationService(
	store EligibleAssignments,
	active ActiveAssignmentLookup,
	activator ActivationStore,
) *ActivationService {
	return &ActivationService{store: store, active: active, activator: activator, resolver: activationResolver{}}
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
			Code:    domain.CodeInvalidDuration,
			Message: "Invalid activation duration.",
		}
	}
	result, ok, err := s.findActive(ctx, target)
	if err != nil {
		return nil, err
	}
	if ok {
		return result, nil
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
	result, ok, err := s.findActive(ctx, target)
	if err != nil {
		return nil, err
	}
	if ok {
		return result, nil
	}
	return s.activator.Activate(ctx, target)
}

func (s *ActivationService) findActive(
	ctx context.Context,
	target domain.ActivationTarget,
) (*domain.ActivationResult, bool, error) {
	if s.active == nil {
		return nil, false, nil
	}
	active, err := s.active.ListActiveForScope(ctx, target.Assignment.ScopeID)
	if err != nil {
		return nil, false, err
	}
	for _, assignment := range active {
		if !matchesActiveAssignment(assignment, target.Assignment) {
			continue
		}
		duration := assignment.EndTime.Sub(assignment.StartTime)
		if duration <= 0 {
			duration = target.Duration
		}
		return &domain.ActivationResult{
			Role:          assignment.Role,
			RoleDefID:     assignment.RoleDefID,
			ScopeID:       assignment.ScopeID,
			ScopeName:     assignment.ScopeName,
			Duration:      duration,
			StartedAt:     assignment.StartTime,
			ExpiresAt:     assignment.EndTime,
			AlreadyActive: true,
		}, true, nil
	}
	return nil, false, nil
}

func matchesActiveAssignment(active domain.ActiveAssignment, eligible domain.EligibleAssignment) bool {
	if active.Status != domain.ActiveAssignmentActive || !strings.EqualFold(active.ScopeID, eligible.ScopeID) {
		return false
	}
	if eligible.RoleDefID != "" && strings.EqualFold(active.RoleDefID, eligible.RoleDefID) {
		return true
	}
	return active.Role == eligible.Role
}
