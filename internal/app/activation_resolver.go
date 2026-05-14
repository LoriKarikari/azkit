package app

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

var guidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type activationResolver struct{}

func (activationResolver) Resolve(req domain.ActivationRequest, eligible []domain.EligibleAssignment) (domain.ActivationTarget, error) {
	scopeID, err := resolveScope(req, eligible)
	if err != nil {
		return domain.ActivationTarget{}, err
	}

	match, ok := findMatch(eligible, scopeID, req.Role)
	if !ok {
		return domain.ActivationTarget{}, ErrEligibleNotFound
	}

	return domain.ActivationTarget{
		Assignment: match,
		Reason:     req.Reason,
		Duration:   req.Duration,
	}, nil
}

func findMatch(eligible []domain.EligibleAssignment, scopeID, role string) (domain.EligibleAssignment, bool) {
	return lo.Find(eligible, func(a domain.EligibleAssignment) bool {
		return a.ScopeID == scopeID && (a.Role == role || a.RoleDefID == role)
	})
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
		if a.SubscriptionID == id || strings.HasPrefix(a.ScopeID, prefix) {
			return id, nil
		}
	}
	return "", &Error{
		Code:    CodeUnknownSubscription,
		Message: fmt.Sprintf("Subscription %s not found among your eligible assignments.%s", id, selectorSuggestions(subscriptionSuggestions(eligible))),
	}
}

func resolveByName(name string, eligible []domain.EligibleAssignment) (string, error) {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)

	seen := make(map[string]string, len(eligible))
	for _, a := range eligible {
		if strings.ToLower(a.SubscriptionName) == lower {
			id := a.SubscriptionID
			if id == "" {
				id = subscriptionIDFromScope(a.ScopeID)
			}
			if id != "" {
				seen[id] = a.SubscriptionName
			}
			continue
		}
		if a.ScopeType == domain.ScopeSubscription && strings.ToLower(a.ScopeName) == lower {
			id := subscriptionIDFromScope(a.ScopeID)
			if id != "" {
				seen[id] = a.ScopeName
			}
		}
	}

	if len(seen) == 0 {
		return "", &Error{
			Code:    CodeUnknownSubscription,
			Message: fmt.Sprintf("Subscription %q not found among your eligible assignments.%s", name, selectorSuggestions(subscriptionSuggestions(eligible))),
		}
	}
	if len(seen) > 1 {
		matches := make([]string, 0, len(seen))
		for id, displayName := range seen {
			matches = append(matches, fmt.Sprintf("%s (%s)", displayName, id))
		}
		sort.Strings(matches)
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

func subscriptionIDFromScope(scopeID string) string {
	id := strings.TrimPrefix(scopeID, "/subscriptions/")
	if id == scopeID {
		return ""
	}
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	return id
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
		Message: fmt.Sprintf("Resource group %q not found in subscription.%s", name, selectorSuggestions(resourceGroupSuggestions(subscriptionID, eligible))),
	}
}

func selectorSuggestions(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return " Try one of: " + strings.Join(values, ", ") + "."
}

func subscriptionSuggestions(eligible []domain.EligibleAssignment) []string {
	seen := make(map[string]struct{}, len(eligible))
	for _, a := range eligible {
		name := a.SubscriptionName
		if name == "" && a.ScopeType == domain.ScopeSubscription {
			name = a.ScopeName
		}
		if name == "" {
			name = a.SubscriptionID
		}
		if name == "" {
			name = subscriptionIDFromScope(a.ScopeID)
		}
		if name != "" {
			seen[name] = struct{}{}
		}
	}
	return sortedKeys(seen)
}

func resourceGroupSuggestions(subscriptionID string, eligible []domain.EligibleAssignment) []string {
	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/", subscriptionID)
	seen := make(map[string]struct{}, len(eligible))
	for _, a := range eligible {
		if a.ScopeType == domain.ScopeResourceGroup && strings.HasPrefix(a.ScopeID, prefix) && a.ScopeName != "" {
			seen[a.ScopeName] = struct{}{}
		}
	}
	return sortedKeys(seen)
}

func sortedKeys(seen map[string]struct{}) []string {
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
