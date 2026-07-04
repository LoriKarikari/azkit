package azurepim

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestAssignmentFromExpanded_mapsSharedRoleAndScopeFields(t *testing.T) {
	role := "Contributor"
	roleDefID := "/roleDefs/contributor"
	scopeName := "rg-prod"
	scopeID := "/subscriptions/abc/resourceGroups/rg-prod"
	scopeType := "resourcegroup"

	got := assignmentFromExpanded(&armauthorization.ExpandedProperties{
		RoleDefinition: &armauthorization.ExpandedPropertiesRoleDefinition{
			DisplayName: &role,
			ID:          &roleDefID,
		},
		Scope: &armauthorization.ExpandedPropertiesScope{
			DisplayName: &scopeName,
			ID:          &scopeID,
			Type:        &scopeType,
		},
	})

	if got.Role != role || got.RoleDefID != roleDefID || got.ScopeName != scopeName || got.ScopeID != scopeID {
		t.Fatalf("unexpected mapping: %+v", got)
	}
	if got.ScopeType != domain.ScopeResourceGroup {
		t.Fatalf("want resource group scope, got %s", got.ScopeType)
	}
}

func TestScopeTypeFromAzure(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  domain.ScopeType
	}{
		{name: "subscription", input: "subscription", want: domain.ScopeSubscription},
		{name: "resource group lowercase", input: "resourcegroup", want: domain.ScopeResourceGroup},
		{name: "resource group ARM type", input: "Microsoft.Resources/subscriptions/resourceGroups", want: domain.ScopeResourceGroup},
		{name: "management group lowercase", input: "managementgroup", want: domain.ScopeManagementGroup},
		{name: "management group ARM type", input: "Microsoft.Management/managementGroups", want: domain.ScopeManagementGroup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scopeTypeFromAzure(tt.input)
			if got != tt.want {
				t.Fatalf("want %s, got %s", tt.want, got)
			}
		})
	}
}
