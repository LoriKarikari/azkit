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
	subscriptions subscriptionSource
	schedules     eligibilityScheduleSource
}

type subscription struct {
	ID string
}

type subscriptionSource interface {
	ListSubscriptions(context.Context) ([]subscription, error)
}

type eligibilityScheduleSource interface {
	ListForSubscription(context.Context, string) ([]domain.EligibleAssignment, error)
}

func NewEligibleAssignments() (*EligibleAssignments, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return newEligibleAssignments(azureSubscriptions{cred: cred}, azureEligibilitySchedules{cred: cred}), nil
}

func newEligibleAssignments(subscriptions subscriptionSource, schedules eligibilityScheduleSource) *EligibleAssignments {
	return &EligibleAssignments{subscriptions: subscriptions, schedules: schedules}
}

func (a *EligibleAssignments) ListEligible(ctx context.Context) ([]domain.EligibleAssignment, error) {
	subs, err := a.subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	var all []domain.EligibleAssignment
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		as, err := a.schedules.ListForSubscription(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		all = append(all, as...)
	}
	return all, nil
}

type azureSubscriptions struct {
	cred azcore.TokenCredential
}

func (a azureSubscriptions) ListSubscriptions(ctx context.Context) ([]subscription, error) {
	client, err := armsubscriptions.NewClient(a.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	var subs []subscription
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sub := range page.Value {
			if sub.SubscriptionID != nil {
				subs = append(subs, subscription{ID: *sub.SubscriptionID})
			}
		}
	}
	return subs, nil
}

type azureEligibilitySchedules struct {
	cred azcore.TokenCredential
}

func (a azureEligibilitySchedules) ListForSubscription(ctx context.Context, subscriptionID string) ([]domain.EligibleAssignment, error) {
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
	a := domain.EligibleAssignment{ScopeType: domain.ScopeSubscription}
	if s.ID != nil {
		a.ID = *s.ID
	}
	if s.Properties == nil {
		return a
	}
	if s.Properties.EndDateTime != nil {
		a.EligibleUntil = *s.Properties.EndDateTime
	}
	if s.Properties.ExpandedProperties == nil {
		return a
	}
	if s.Properties.ExpandedProperties.RoleDefinition != nil && s.Properties.ExpandedProperties.RoleDefinition.DisplayName != nil {
		a.Role = *s.Properties.ExpandedProperties.RoleDefinition.DisplayName
	}
	if s.Properties.ExpandedProperties.Scope == nil {
		return a
	}
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
	return a
}
