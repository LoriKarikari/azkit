package azurepim

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

type EligibleAssignments struct {
	subscriptions subscriptionSource
	schedules     eligibilityScheduleSource
	log           *slog.Logger
}

type subscription struct {
	ID   string
	Name string
}

var (
	_ app.EligibleAssignments   = (*EligibleAssignments)(nil)
	_ subscriptionSource        = (*azureSubscriptions)(nil)
	_ eligibilityScheduleSource = (*azureEligibilitySchedules)(nil)
)

type subscriptionSource interface {
	ListSubscriptions(context.Context) ([]subscription, error)
}

type eligibilityScheduleSource interface {
	ListForSubscription(context.Context, string) ([]domain.EligibleAssignment, error)
}

func NewEligibleAssignmentsFromCred(cred azcore.TokenCredential, log *slog.Logger) *EligibleAssignments {
	return newEligibleAssignments(
		azureSubscriptions{cred: cred},
		azureEligibilitySchedules{cred: cred},
		log,
	)
}

func newEligibleAssignments(
	subscriptions subscriptionSource,
	schedules eligibilityScheduleSource,
	log *slog.Logger,
) *EligibleAssignments {
	return &EligibleAssignments{
		subscriptions: subscriptions,
		schedules:     schedules,
		log:           logger(log),
	}
}

func (a *EligibleAssignments) ListEligible(ctx context.Context) ([]domain.EligibleAssignment, error) {
	return listAcrossSubscriptions(
		ctx,
		a.subscriptions,
		a.log,
		"listing eligible assignments",
		func(ctx context.Context, sub subscription) ([]domain.EligibleAssignment, error) {
			return a.schedules.ListForSubscription(ctx, sub.ID)
		},
		func(assignment *domain.EligibleAssignment, sub subscription) {
			assignment.SubscriptionID = sub.ID
			assignment.SubscriptionName = sub.Name
		},
	)
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
	subs := []subscription{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sub := range page.Value {
			if sub.SubscriptionID != nil {
				s := subscription{ID: *sub.SubscriptionID}
				if sub.DisplayName != nil {
					s.Name = *sub.DisplayName
				}
				subs = append(subs, s)
			}
		}
	}
	return subs, nil
}

type azureEligibilitySchedules struct {
	cred azcore.TokenCredential
}

func (a azureEligibilitySchedules) ListForSubscription(
	ctx context.Context,
	subscriptionID string,
) ([]domain.EligibleAssignment, error) {
	client, err := armauthorization.NewRoleEligibilitySchedulesClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	filter := "asTarget()"
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	pager := client.NewListForScopePager(scope, &armauthorization.RoleEligibilitySchedulesClientListForScopeOptions{
		Filter: &filter,
	})

	assignments := []domain.EligibleAssignment{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, azurePIMOperationError(err)
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
	if s.Properties.PrincipalID != nil {
		a.PrincipalID = *s.Properties.PrincipalID
	}
	if s.Properties.EndDateTime != nil {
		a.EligibleUntil = *s.Properties.EndDateTime
	}
	mapping := assignmentFromExpanded(s.Properties.ExpandedProperties)
	a.Role = mapping.Role
	a.RoleDefID = mapping.RoleDefID
	a.ScopeID = mapping.ScopeID
	a.ScopeName = mapping.ScopeName
	a.ScopeType = mapping.ScopeType
	return a
}
