package azurepim

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActiveAssignments struct {
	subscriptions subscriptionSource
	schedules     activeScheduleSource
}

type activeScheduleSource interface {
	ListForScope(context.Context, string) ([]domain.ActiveAssignment, error)
}

func NewActiveAssignments(cred azcore.TokenCredential) *ActiveAssignments {
	return newActiveAssignments(azureSubscriptions{cred: cred}, azureActiveSchedules{cred: cred})
}

func newActiveAssignments(subscriptions subscriptionSource, schedules activeScheduleSource) *ActiveAssignments {
	return &ActiveAssignments{subscriptions: subscriptions, schedules: schedules}
}

func (a *ActiveAssignments) ListActive(ctx context.Context) ([]domain.ActiveAssignment, error) {
	subs, err := a.subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	var all []domain.ActiveAssignment
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		scope := fmt.Sprintf("/subscriptions/%s", sub.ID)
		as, err := a.schedules.ListForScope(ctx, scope)
		if err != nil {
			return nil, err
		}
		for i := range as {
			as[i].SubscriptionID = sub.ID
			as[i].SubscriptionName = sub.Name
		}
		all = append(all, as...)
	}
	return all, nil
}

type azureActiveSchedules struct {
	cred azcore.TokenCredential
}

func (a azureActiveSchedules) ListForScope(ctx context.Context, scope string) ([]domain.ActiveAssignment, error) {
	client, err := armauthorization.NewRoleAssignmentSchedulesClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	filter := "asTarget()"
	pager := client.NewListForScopePager(scope, &armauthorization.RoleAssignmentSchedulesClientListForScopeOptions{
		Filter: &filter,
	})

	var as []domain.ActiveAssignment
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, azurePIMOperationError(err)
		}
		for _, sched := range page.Value {
			as = append(as, activeToDomain(sched))
		}
	}
	return as, nil
}

func activeToDomain(s *armauthorization.RoleAssignmentSchedule) domain.ActiveAssignment {
	a := domain.ActiveAssignment{ScopeType: domain.ScopeSubscription}
	if s.ID != nil {
		a.ID = *s.ID
	}
	if s.Properties == nil {
		return a
	}
	if s.Properties.PrincipalID != nil {
		a.PrincipalID = *s.Properties.PrincipalID
	}
	if s.Properties.StartDateTime != nil {
		a.StartTime = *s.Properties.StartDateTime
	}
	if s.Properties.EndDateTime != nil {
		a.EndTime = *s.Properties.EndDateTime
	}
	if s.Properties.Status != nil {
		a.Status = domain.ActiveAssignmentStatus(strings.ToLower(string(*s.Properties.Status)))
	}
	if s.Properties.ExpandedProperties == nil {
		return a
	}
	if s.Properties.ExpandedProperties.RoleDefinition != nil {
		if s.Properties.ExpandedProperties.RoleDefinition.DisplayName != nil {
			a.Role = *s.Properties.ExpandedProperties.RoleDefinition.DisplayName
		}
		if s.Properties.ExpandedProperties.RoleDefinition.ID != nil {
			a.RoleDefID = *s.Properties.ExpandedProperties.RoleDefinition.ID
		}
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
