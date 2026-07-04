package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestFilterEligibleAssignments(t *testing.T) {
	eligible := []domain.EligibleAssignment{
		{
			Role:             "Reader",
			RoleDefID:        "reader-id",
			ScopeType:        domain.ScopeSubscription,
			ScopeID:          "/subscriptions/sub-a",
			ScopeName:        "Prod",
			SubscriptionID:   "sub-a",
			SubscriptionName: "Prod",
		},
		{
			Role:             "Contributor",
			RoleDefID:        "contrib-id",
			ScopeType:        domain.ScopeResourceGroup,
			ScopeID:          "/subscriptions/sub-a/resourceGroups/rg-app",
			ScopeName:        "rg-app",
			SubscriptionID:   "sub-a",
			SubscriptionName: "Prod",
		},
	}

	tests := []struct {
		name     string
		req      domain.ActivationRequest
		expected []string
	}{
		{
			name: "role display name",
			req: domain.ActivationRequest{
				Role: "reader",
			},
			expected: []string{"Reader"},
		},
		{
			name: "role definition id",
			req: domain.ActivationRequest{
				Role: "contrib-id",
			},
			expected: []string{"Contributor"},
		},
		{
			name: "subscription selector",
			req: domain.ActivationRequest{
				Subscription: "Prod",
			},
			expected: []string{"Reader", "Contributor"},
		},
		{
			name: "resource group selector",
			req: domain.ActivationRequest{
				ResourceGroup: "RG-APP",
			},
			expected: []string{"Contributor"},
		},
		{
			name: "resource scope",
			req: domain.ActivationRequest{
				ScopeID: "/subscriptions/sub-a",
			},
			expected: []string{"Reader"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.FilterEligibleAssignments(eligible, tt.req)
			roles := make([]string, 0, len(got))
			for _, assignment := range got {
				roles = append(roles, assignment.Role)
			}
			require.Equal(t, tt.expected, roles)
		})
	}
}
