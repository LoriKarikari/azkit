package azurepim

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type EligibleAssignments struct {
	cred azcore.TokenCredential
}

func NewEligibleAssignments() (*EligibleAssignments, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return &EligibleAssignments{cred: cred}, nil
}

func (a *EligibleAssignments) ListEligible(ctx context.Context) ([]domain.EligibleAssignment, error) {
	subs, err := a.subscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	var all []domain.EligibleAssignment
	for _, sub := range subs {
		scope := *sub.SubscriptionID
		as, err := a.listForSubscription(ctx, scope)
		if err != nil {
			return nil, err
		}
		all = append(all, as...)
	}
	return all, nil
}

func (a *EligibleAssignments) subscriptions(ctx context.Context) ([]*armsubscriptions.Subscription, error) {
	client, err := armsubscriptions.NewClient(a.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	var subs []*armsubscriptions.Subscription
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		subs = append(subs, page.Value...)
	}
	return subs, nil
}

func (a *EligibleAssignments) listForSubscription(ctx context.Context, subscriptionID string) ([]domain.EligibleAssignment, error) {
	client, err := armauthorization.NewRoleEligibilitySchedulesClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	filter := "asTarget()"
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	pager := client.NewListForScopePager(scope, &armauthorization.RoleEligibilitySchedulesClientListForScopeOptions{
		Filter: &filter,
	})

	var assignments []domain.EligibleAssignment
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "AuthorizationFailed") || strings.Contains(err.Error(), "403") {
				return nil, app.PermissionDenied(err)
			}
			return nil, app.AzureAPIError(err)
		}
		for _, sched := range page.Value {
			assignments = append(assignments, toDomain(sched))
		}
	}
	return assignments, nil
}

func toDomain(s *armauthorization.RoleEligibilitySchedule) domain.EligibleAssignment {
	a := domain.EligibleAssignment{
		ScopeType: domain.ScopeSubscription,
	}
	if s.ID != nil {
		a.ID = *s.ID
	}
	if s.Properties != nil {
		if s.Properties.ExpandedProperties != nil {
			if s.Properties.ExpandedProperties.RoleDefinition != nil {
				if s.Properties.ExpandedProperties.RoleDefinition.DisplayName != nil {
					a.Role = *s.Properties.ExpandedProperties.RoleDefinition.DisplayName
				}
			}
			if s.Properties.ExpandedProperties.Scope != nil {
				if s.Properties.ExpandedProperties.Scope.DisplayName != nil {
					a.ScopeName = *s.Properties.ExpandedProperties.Scope.DisplayName
				}
				if s.Properties.ExpandedProperties.Scope.ID != nil {
					a.ScopeID = *s.Properties.ExpandedProperties.Scope.ID
				}

				typeStr := ""
				if s.Properties.ExpandedProperties.Scope.Type != nil {
					typeStr = *s.Properties.ExpandedProperties.Scope.Type
				}
				if strings.Contains(typeStr, "resourceGroup") || strings.Contains(typeStr, "resourceGroups") {
					a.ScopeType = domain.ScopeResourceGroup
				}
			}
		}
		if s.Properties.EndDateTime != nil {
			a.MaxDuration = "8h"
		}
	}
	return a
}
