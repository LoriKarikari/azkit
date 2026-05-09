package azurepim

import (
	"context"
	"fmt"
	"log/slog"

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
	log           *slog.Logger
}

type subscription struct {
	ID   string
	Name string
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
	return NewEligibleAssignmentsFromCred(cred, nil), nil
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
	subs, err := a.subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	a.log.Debug("listed subscriptions", slog.Int("count", len(subs)))

	var all []domain.EligibleAssignment
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		a.log.Debug("listing eligible assignments", slog.String("subscription_id", sub.ID))
		as, err := a.schedules.ListForSubscription(ctx, sub.ID)
		if err != nil {
			a.log.Debug(
				"eligible assignment listing failed",
				slog.String("subscription_id", sub.ID),
				slog.Any("error", err),
			)
			return nil, err
		}
		a.log.Debug(
			"listed eligible assignments",
			slog.String("subscription_id", sub.ID),
			slog.Int("count", len(as)),
		)
		for i := range as {
			as[i].SubscriptionID = sub.ID
			as[i].SubscriptionName = sub.Name
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

	var assignments []domain.EligibleAssignment
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
