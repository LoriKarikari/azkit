package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

const (
	CodeEligibleNotFound      Code = "eligible_not_found"
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
	ErrMissingScope         = &Error{Code: CodeMissingScope, Message: "Activation scope is required."}
	ErrMissingRole          = &Error{Code: CodeMissingRole, Message: "Activation role is required."}
	ErrMissingReason        = &Error{Code: CodeMissingReason, Message: "Activation reason is required."}
	ErrConflictingSelectors = &Error{Code: CodeConflictingSelectors, Message: "Use --scope or --subscription, not both."}
)

var guidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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

	scopeID, err := resolveScope(req, eligible)
	if err != nil {
		return nil, err
	}

	match := findMatch(eligible, scopeID, req.Role)
	if match == nil {
		return nil, ErrEligibleNotFound
	}

	return s.activator.Activate(ctx, domain.ActivationTarget{
		Assignment: *match,
		Reason:     req.Reason,
		Duration:   req.Duration,
	})
}

func findMatch(eligible []domain.EligibleAssignment, scopeID, role string) *domain.EligibleAssignment {
	for i := range eligible {
		if eligible[i].ScopeID == scopeID &&
			(eligible[i].Role == role || eligible[i].RoleDefID == role) {
			return &eligible[i]
		}
	}
	return nil
}

func resolveScope(req domain.ActivationRequest, eligible []domain.EligibleAssignment) (string, error) {
	if req.ScopeID != "" {
		if req.Subscription != "" || req.ResourceGroup != "" {
			return "", ErrConflictingSelectors
		}
		return req.ScopeID, nil
	}
	if req.Subscription == "" {
		return "", ErrMissingScope
	}

	subID, err := resolveSubscription(req.Subscription, eligible)
	if err != nil {
		return "", err
	}

	if req.ResourceGroup == "" {
		return "/subscriptions/" + subID, nil
	}

	return resolveResourceGroup(subID, req.ResourceGroup, eligible)
}

func resolveSubscription(input string, eligible []domain.EligibleAssignment) (string, error) {
	if guidRE.MatchString(input) {
		return resolveByID(input, eligible)
	}
	return resolveByName(input, eligible)
}

func resolveByID(id string, eligible []domain.EligibleAssignment) (string, error) {
	prefix := "/subscriptions/" + id
	for _, a := range eligible {
		if strings.HasPrefix(a.ScopeID, prefix) {
			return id, nil
		}
	}
	return "", &Error{
		Code:    CodeUnknownSubscription,
		Message: fmt.Sprintf("Subscription %s not found among your eligible assignments.", id),
	}
}

func resolveByName(name string, eligible []domain.EligibleAssignment) (string, error) {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)

	seen := map[string]string{}
	for _, a := range eligible {
		if a.ScopeType == domain.ScopeSubscription && strings.ToLower(a.ScopeName) == lower {
			id := strings.TrimPrefix(a.ScopeID, "/subscriptions/")
			if idx := strings.Index(id, "/"); idx != -1 {
				id = id[:idx]
			}
			seen[id] = a.ScopeName
		}
	}

	if len(seen) == 0 {
		return "", &Error{
			Code:    CodeUnknownSubscription,
			Message: fmt.Sprintf("Subscription %q not found among your eligible assignments.", name),
		}
	}
	if len(seen) > 1 {
		var matches []string
		for _, displayName := range seen {
			matches = append(matches, displayName)
		}
		return "", &Error{
			Code:    CodeAmbiguousSubscription,
			Message: fmt.Sprintf("Subscription name %q is ambiguous. Matches: %s. Use the subscription ID instead.", name, strings.Join(matches, ", ")),
		}
	}
	for id := range seen {
		return id, nil
	}
	return "", nil
}

func resolveResourceGroup(subscriptionID, name string, eligible []domain.EligibleAssignment) (string, error) {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)
	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/", subscriptionID)

	for _, a := range eligible {
		if a.ScopeType == domain.ScopeResourceGroup &&
			strings.HasPrefix(a.ScopeID, prefix) &&
			strings.ToLower(a.ScopeName) == lower {
			return a.ScopeID, nil
		}
	}
	return "", &Error{
		Code:    CodeUnknownResourceGroup,
		Message: fmt.Sprintf("Resource group %q not found in subscription.", name),
	}
}

func validateActivation(req domain.ActivationRequest) error {
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
