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

func TestActiveAssignments_listsCurrentActiveInstancesAcrossSubscriptions(t *testing.T) {
	now := activeTime("2026-05-07T18:00:00Z")
	adapter := newActiveAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a", Name: "Sub A"}, {ID: "sub-b", Name: "Sub B"}}},
		fakeActiveInstances{instances: map[string][]*armauthorization.RoleAssignmentScheduleInstance{
			"/subscriptions/sub-a": {
				activeInstance("a1", "Contributor", "/roleDefs/contributor", "/subscriptions/sub-a", "Sub A", armauthorization.StatusProvisioned, "2026-05-07T17:00:00Z", "2026-05-07T19:00:00Z"),
				activeInstance("old", "Reader", "/roleDefs/reader", "/subscriptions/sub-a", "Sub A", armauthorization.StatusProvisioned, "2026-05-07T15:00:00Z", "2026-05-07T17:00:00Z"),
			},
			"/subscriptions/sub-b": {
				activeInstance("a2", "Reader", "/roleDefs/reader", "/subscriptions/sub-b", "Sub B", armauthorization.StatusGranted, "2026-05-07T17:30:00Z", "2026-05-07T20:00:00Z"),
				activeInstance("pending", "Owner", "/roleDefs/owner", "/subscriptions/sub-b", "Sub B", armauthorization.StatusPendingProvisioning, "2026-05-07T17:30:00Z", "2026-05-07T20:00:00Z"),
			},
		}},
		func() time.Time { return now },
		nil,
	)

	got, err := adapter.ListActive(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 assignments, got %d: %+v", len(got), got)
	}
	if got[0].ID != "a1" || got[1].ID != "a2" {
		t.Fatalf("unexpected assignments: %+v", got)
	}
	if got[0].Status != domain.ActiveAssignmentActive || got[1].Status != domain.ActiveAssignmentActive {
		t.Fatalf("want active statuses, got %+v", got)
	}
	if got[0].SubscriptionName != "Sub A" || got[1].SubscriptionName != "Sub B" {
		t.Fatalf("unexpected subscription names: %+v", got)
	}
}

func TestActiveAssignments_returnsInstanceError(t *testing.T) {
	adapter := newActiveAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a"}}},
		fakeActiveInstances{err: app.PermissionDenied(errors.New("denied"))},
		func() time.Time { return activeTime("2026-05-07T18:00:00Z") },
		nil,
	)

	_, err := adapter.ListActive(t.Context())
	if err == nil {
		t.Fatalf("want error, got nil")
	}
}

type fakeActiveInstances struct {
	instances map[string][]*armauthorization.RoleAssignmentScheduleInstance
	err       error
}

func (f fakeActiveInstances) ListForScope(_ context.Context, scope string) ([]*armauthorization.RoleAssignmentScheduleInstance, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.instances[scope], nil
}

func activeInstance(id, role, roleDefID, scopeID, scopeName string, status armauthorization.Status, start, end string) *armauthorization.RoleAssignmentScheduleInstance {
	scopeType := "subscription"
	startTime := activeTime(start)
	endTime := activeTime(end)
	return &armauthorization.RoleAssignmentScheduleInstance{
		ID: &id,
		Properties: &armauthorization.RoleAssignmentScheduleInstanceProperties{
			PrincipalID:   new("user-1"),
			StartDateTime: &startTime,
			EndDateTime:   &endTime,
			Status:        &status,
			ExpandedProperties: &armauthorization.ExpandedProperties{
				RoleDefinition: &armauthorization.ExpandedPropertiesRoleDefinition{
					DisplayName: &role,
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
}

func activeTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
