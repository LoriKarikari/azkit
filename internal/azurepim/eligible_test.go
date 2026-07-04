package azurepim

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestEligibleAssignments_listsAcrossSubscriptions(t *testing.T) {
	adapter := newEligibleAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a", Name: "Sub A"}, {ID: "sub-b", Name: "Sub B"}}},
		fakeSchedules{assignments: map[string][]domain.EligibleAssignment{
			"sub-a": {{ID: "a1", Role: "Contributor"}},
			"sub-b": {{ID: "a2", Role: "Reader"}},
		}},
		nil,
	)

	got, err := adapter.ListEligible(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 assignments, got %d", len(got))
	}
	if got[0].ID != "a1" || got[1].ID != "a2" {
		t.Fatalf("unexpected assignments: %+v", got)
	}
	if got[0].SubscriptionName != "Sub A" || got[1].SubscriptionName != "Sub B" {
		t.Fatalf("unexpected subscription names: %+v", got)
	}
}

func TestEligibleAssignments_wrapsSubscriptionErrorAsAzureAPIError(t *testing.T) {
	adapter := newEligibleAssignments(
		fakeSubscriptions{err: errors.New("network failed")},
		fakeSchedules{},
		nil,
	)

	_, err := adapter.ListEligible(t.Context())
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeAzureAPIError {
		t.Fatalf("want %s, got %s", domain.CodeAzureAPIError, appErr.Code)
	}
}

func TestEligibleAssignments_returnsScheduleError(t *testing.T) {
	want := app.PermissionDenied(errors.New("denied"), "PIM")
	adapter := newEligibleAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a"}}},
		fakeSchedules{err: want},
		nil,
	)

	_, err := adapter.ListEligible(t.Context())
	if !errors.Is(err, want) {
		t.Fatalf("want schedule error, got %v", err)
	}
}

func TestEligibleAssignments_failsWholeListWhenSubscriptionScheduleFails(t *testing.T) {
	want := app.AzureAPIError(errors.New("sub-b failed"))
	adapter := newEligibleAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a"}, {ID: "sub-b"}}},
		fakeSchedules{
			assignments: map[string][]domain.EligibleAssignment{
				"sub-a": {{ID: "a1"}},
			},
			errors: map[string]error{
				"sub-b": want,
			},
		},
		nil,
	)

	got, err := adapter.ListEligible(t.Context())
	if !errors.Is(err, want) {
		t.Fatalf("want schedule error, got %v", err)
	}
	if got != nil {
		t.Fatalf("want no partial assignments, got %+v", got)
	}
}

func TestEligibleAssignments_skipsBlankSubscriptionID(t *testing.T) {
	schedules := fakeSchedules{assignments: map[string][]domain.EligibleAssignment{
		"sub-a": {{ID: "a1"}},
	}}
	adapter := newEligibleAssignments(
		fakeSubscriptions{subs: []subscription{{ID: ""}, {ID: "sub-a"}}},
		schedules,
		nil,
	)

	got, err := adapter.ListEligible(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "a1" {
		t.Fatalf("unexpected assignments: %+v", got)
	}
}

func TestEligibleAssignments_deduplicatesInheritedAssignments(t *testing.T) {
	adapter := newEligibleAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a", Name: "Sub A"}, {ID: "sub-b", Name: "Sub B"}}},
		fakeSchedules{assignments: map[string][]domain.EligibleAssignment{
			"sub-a": {{ID: "mg-assignment", Role: "Reader", ScopeType: domain.ScopeManagementGroup}},
			"sub-b": {{ID: "mg-assignment", Role: "Reader", ScopeType: domain.ScopeManagementGroup}},
		}},
		nil,
	)

	got, err := adapter.ListEligible(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 assignment, got %d: %+v", len(got), got)
	}
	if got[0].SubscriptionName != "Sub A" {
		t.Fatalf("want first subscription enrichment kept, got %+v", got[0])
	}
}

func TestToDomain_mapsFields(t *testing.T) {
	displayName := "Contributor"
	roleDefID := "/subscriptions/abc/providers/Microsoft.Authorization/roleDefinitions/111"
	scopeName := "sub-prod"
	scopeID := "/subscriptions/abc"
	scopeType := "subscription"
	schedID := "/subscriptions/abc/providers/Microsoft.Authorization/roleEligibilitySchedules/sched1"
	end := time.Date(2026, 5, 7, 20, 0, 0, 0, time.UTC)

	sched := &armauthorization.RoleEligibilitySchedule{
		ID: &schedID,
		Properties: &armauthorization.RoleEligibilityScheduleProperties{
			EndDateTime: &end,
			ExpandedProperties: &armauthorization.ExpandedProperties{
				RoleDefinition: &armauthorization.ExpandedPropertiesRoleDefinition{
					DisplayName: &displayName,
					ID:          &roleDefID,
				},
				Scope: &armauthorization.ExpandedPropertiesScope{
					DisplayName: &scopeName,
					ID:          &scopeID,
					Type:        &scopeType,
				},
			},
		},
	}

	a := toDomain(sched)

	if a.ID != schedID {
		t.Fatalf("want ID %q, got %q", schedID, a.ID)
	}
	if a.Role != "Contributor" {
		t.Fatalf("want Role Contributor, got %q", a.Role)
	}
	if a.ScopeType != "subscription" {
		t.Fatalf("want ScopeType subscription, got %q", a.ScopeType)
	}
	if a.ScopeName != "sub-prod" {
		t.Fatalf("want ScopeName sub-prod, got %q", a.ScopeName)
	}
	if a.ScopeID != "/subscriptions/abc" {
		t.Fatalf("want ScopeID /subscriptions/abc, got %q", a.ScopeID)
	}
	if !a.EligibleUntil.Equal(end) {
		t.Fatalf("want EligibleUntil %s, got %s", end, a.EligibleUntil)
	}
}

func TestToDomain_resourceGroupScope(t *testing.T) {
	scopeType := "resourcegroup"

	sched := &armauthorization.RoleEligibilitySchedule{
		Properties: &armauthorization.RoleEligibilityScheduleProperties{
			ExpandedProperties: &armauthorization.ExpandedProperties{
				Scope: &armauthorization.ExpandedPropertiesScope{
					Type: &scopeType,
				},
			},
		},
	}

	a := toDomain(sched)
	if a.ScopeType != "resource_group" {
		t.Fatalf("want ScopeType resource_group, got %q", a.ScopeType)
	}
}

func TestToDomain_nilSafe(t *testing.T) {
	sched := &armauthorization.RoleEligibilitySchedule{}
	a := toDomain(sched)
	if a.ID != "" || a.Role != "" || a.ScopeName != "" || a.ScopeID != "" {
		t.Fatalf("want empty assignment for nil fields, got %+v", a)
	}
}

type fakeSubscriptions struct {
	subs []subscription
	err  error
}

func (f fakeSubscriptions) ListSubscriptions(context.Context) ([]subscription, error) {
	return f.subs, f.err
}

type fakeSchedules struct {
	assignments map[string][]domain.EligibleAssignment
	errors      map[string]error
	err         error
}

func (f fakeSchedules) ListForSubscription(_ context.Context, subscriptionID string) ([]domain.EligibleAssignment, error) {
	if f.err != nil {
		return nil, f.err
	}
	if err := f.errors[subscriptionID]; err != nil {
		return nil, err
	}
	return f.assignments[subscriptionID], nil
}
