package azurepim

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActiveAssignments struct {
	subscriptions subscriptionSource
	instances     activeInstanceSource
	now           func() time.Time
}

type activeInstanceSource interface {
	ListForScope(context.Context, string) ([]*armauthorization.RoleAssignmentScheduleInstance, error)
}

func NewActiveAssignments(cred azcore.TokenCredential) *ActiveAssignments {
	return newActiveAssignments(azureSubscriptions{cred: cred}, azureActiveInstances{cred: cred}, time.Now)
}

func newActiveAssignments(subscriptions subscriptionSource, instances activeInstanceSource, now func() time.Time) *ActiveAssignments {
	return &ActiveAssignments{subscriptions: subscriptions, instances: instances, now: now}
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
		instances, err := a.instances.ListForScope(ctx, scope)
		if err != nil {
			return nil, err
		}
		for _, instance := range instances {
			assignment, ok := activeInstanceToDomain(instance, a.now().UTC())
			if !ok {
				continue
			}
			assignment.SubscriptionID = sub.ID
			assignment.SubscriptionName = sub.Name
			all = append(all, assignment)
		}
	}
	return all, nil
}

type azureActiveInstances struct {
	cred azcore.TokenCredential
}

func (a azureActiveInstances) ListForScope(ctx context.Context, scope string) ([]*armauthorization.RoleAssignmentScheduleInstance, error) {
	client, err := armauthorization.NewRoleAssignmentScheduleInstancesClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	filter := "asTarget()"
	pager := client.NewListForScopePager(scope, &armauthorization.RoleAssignmentScheduleInstancesClientListForScopeOptions{
		Filter: &filter,
	})

	var instances []*armauthorization.RoleAssignmentScheduleInstance
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, azurePIMOperationError(err)
		}
		instances = append(instances, page.Value...)
	}
	return instances, nil
}

func activeInstanceToDomain(s *armauthorization.RoleAssignmentScheduleInstance, now time.Time) (domain.ActiveAssignment, bool) {
	if !currentActiveInstance(s, now) {
		return domain.ActiveAssignment{}, false
	}

	a := domain.ActiveAssignment{ScopeType: domain.ScopeSubscription, Status: domain.ActiveAssignmentActive}
	if s.ID != nil {
		a.ID = *s.ID
	}
	if s.Properties == nil {
		return a, true
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
	mapping := assignmentFromExpanded(s.Properties.ExpandedProperties)
	a.Role = mapping.Role
	a.RoleDefID = mapping.RoleDefID
	a.ScopeID = mapping.ScopeID
	a.ScopeName = mapping.ScopeName
	a.ScopeType = mapping.ScopeType
	return a, true
}

func currentActiveInstance(s *armauthorization.RoleAssignmentScheduleInstance, now time.Time) bool {
	if s == nil || s.Properties == nil || s.Properties.Status == nil {
		return false
	}
	switch *s.Properties.Status {
	case armauthorization.StatusGranted, armauthorization.StatusProvisioned:
	default:
		return false
	}
	if s.Properties.StartDateTime != nil && now.Before(s.Properties.StartDateTime.UTC()) {
		return false
	}
	if s.Properties.EndDateTime != nil && !now.Before(s.Properties.EndDateTime.UTC()) {
		return false
	}
	return true
}
