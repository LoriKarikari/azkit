package azurepim

import (
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
)

func TestToDomain_mapsFields(t *testing.T) {
	displayName := "Contributor"
	roleDefID := "/subscriptions/abc/providers/Microsoft.Authorization/roleDefinitions/111"
	scopeName := "sub-prod"
	scopeID := "/subscriptions/abc"
	scopeType := "subscription"
	schedID := "/subscriptions/abc/providers/Microsoft.Authorization/roleEligibilitySchedules/sched1"

	sched := &armauthorization.RoleEligibilitySchedule{
		ID: &schedID,
		Properties: &armauthorization.RoleEligibilityScheduleProperties{
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
}

func TestToDomain_resourceGroupScope(t *testing.T) {
	scopeType := "Microsoft.Authorization/roleEligibilitySchedules/resourceGroups"

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

func TestListForSubscription_liveIntegration(t *testing.T) {
	if os.Getenv("PIMCTL_LIVE_TESTS") != "1" {
		t.Skip("set PIMCTL_LIVE_TESTS=1 to run")
	}
	if os.Getenv("PIMCTL_LIVE_SUBSCRIPTION") == "" {
		t.Skip("set PIMCTL_LIVE_SUBSCRIPTION to a subscription ID")
	}

	t.Log("building credential...")
	adapter, err := NewEligibleAssignments()
	if err != nil {
		t.Fatalf("NewEligibleAssignments: %v", err)
	}

	scope := os.Getenv("PIMCTL_LIVE_SUBSCRIPTION")
	t.Logf("listing eligible assignments for subscription %s...", scope)

	as, err := adapter.listForSubscription(t.Context(), scope)
	if err != nil {
		t.Fatalf("listForSubscription: %v", err)
	}
	t.Logf("got %d assignments", len(as))
	for _, a := range as {
		t.Logf("  %s %s %s %s", a.ID, a.Role, a.ScopeType, a.ScopeName)
	}
}
