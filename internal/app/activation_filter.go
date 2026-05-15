package app

import (
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

// FilterEligibleAssignments returns eligible assignments matching the Activation selectors supplied by the caller.
func FilterEligibleAssignments(
	eligible []domain.EligibleAssignment,
	req domain.ActivationRequest,
) []domain.EligibleAssignment {
	matches := make([]domain.EligibleAssignment, 0, len(eligible))
	for _, assignment := range eligible {
		if eligibleAssignmentMatches(assignment, req) {
			matches = append(matches, assignment)
		}
	}
	return matches
}

func eligibleAssignmentMatches(a domain.EligibleAssignment, req domain.ActivationRequest) bool {
	role := strings.TrimSpace(req.Role)
	if role != "" && a.Role != role && a.RoleDefID != role {
		return false
	}
	if req.ScopeID != "" && a.ScopeID != req.ScopeID {
		return false
	}
	if req.Subscription != "" && !eligibleAssignmentMatchesSubscription(a, req.Subscription) {
		return false
	}
	if req.ResourceGroup != "" && !eligibleAssignmentMatchesResourceGroup(a, req.ResourceGroup) {
		return false
	}
	return true
}

func eligibleAssignmentMatchesSubscription(a domain.EligibleAssignment, input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return true
	}
	lower := strings.ToLower(input)
	if a.SubscriptionID == input || strings.EqualFold(a.SubscriptionName, input) {
		return true
	}
	if a.ScopeType == domain.ScopeSubscription && strings.EqualFold(a.ScopeName, input) {
		return true
	}
	return strings.HasPrefix(strings.ToLower(a.ScopeID), "/subscriptions/"+lower)
}

func eligibleAssignmentMatchesResourceGroup(a domain.EligibleAssignment, input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return true
	}
	return a.ScopeType == domain.ScopeResourceGroup && strings.EqualFold(a.ScopeName, input)
}
