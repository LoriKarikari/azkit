package azurepim

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

type ActiveAssignments struct {
	subscriptions subscriptionSource
	instances     activeInstanceSource
	now           func() time.Time
	log           *slog.Logger
}

var (
	_ app.ActiveAssignments = (*ActiveAssignments)(nil)
	_ activeInstanceSource  = (*azureActiveInstances)(nil)
)

type activeInstanceSource interface {
	ListForScope(
		context.Context,
		string,
	) ([]*armauthorization.RoleAssignmentScheduleInstance, error)
}

func NewActiveAssignments(cred azcore.TokenCredential, log *slog.Logger) *ActiveAssignments {
	return newActiveAssignments(
		azureSubscriptions{cred: cred},
		azureActiveInstances{cred: cred},
		time.Now,
		log,
	)
}

func newActiveAssignments(
	subscriptions subscriptionSource,
	instances activeInstanceSource,
	now func() time.Time,
	log *slog.Logger,
) *ActiveAssignments {
	return &ActiveAssignments{
		subscriptions: subscriptions,
		instances:     instances,
		now:           now,
		log:           logger(log),
	}
}

func (a *ActiveAssignments) ListActive(ctx context.Context) ([]domain.ActiveAssignment, error) {
	return listAcrossSubscriptions(
		ctx,
		a.subscriptions,
		a.log,
		"listing active assignment instances",
		func(ctx context.Context, sub subscription) ([]domain.ActiveAssignment, error) {
			scope := fmt.Sprintf("/subscriptions/%s", sub.ID)
			return a.ListActiveForScope(ctx, scope)
		},
		func(assignment *domain.ActiveAssignment, sub subscription) {
			assignment.SubscriptionID = sub.ID
			assignment.SubscriptionName = sub.Name
		},
		func(assignment domain.ActiveAssignment) string {
			return assignment.ID
		},
	)
}

func (a *ActiveAssignments) ListActiveForScope(ctx context.Context, scope string) ([]domain.ActiveAssignment, error) {
	a.log.Debug("listing active assignment instances", slog.String("scope", scope))
	instances, err := a.instances.ListForScope(ctx, scope)
	if err != nil {
		a.log.Debug(
			"active assignment instance listing failed",
			slog.String("scope", scope),
			slog.Any("error", err),
		)
		return nil, err
	}
	a.log.Debug(
		"listed active assignment instances",
		slog.String("scope", scope),
		slog.Int("count", len(instances)),
	)
	assignments := []domain.ActiveAssignment{}
	for _, instance := range instances {
		assignment, ok := activeInstanceToDomain(instance, a.now().UTC())
		if !ok {
			continue
		}
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}

type azureActiveInstances struct {
	cred azcore.TokenCredential
}

func (a azureActiveInstances) ListForScope(
	ctx context.Context,
	scope string,
) ([]*armauthorization.RoleAssignmentScheduleInstance, error) {
	client, err := armauthorization.NewRoleAssignmentScheduleInstancesClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	filter := "asTarget()"
	pager := client.NewListForScopePager(scope, &armauthorization.RoleAssignmentScheduleInstancesClientListForScopeOptions{
		Filter: &filter,
	})

	instances := []*armauthorization.RoleAssignmentScheduleInstance{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, azurePIMOperationError(err)
		}
		instances = append(instances, page.Value...)
	}
	return instances, nil
}

func activeInstanceToDomain(
	s *armauthorization.RoleAssignmentScheduleInstance,
	now time.Time,
) (domain.ActiveAssignment, bool) {
	if !currentActiveInstance(s, now) {
		return domain.ActiveAssignment{}, false
	}

	a := domain.ActiveAssignment{
		ScopeType: domain.ScopeSubscription,
		Status:    domain.ActiveAssignmentActive,
	}
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
