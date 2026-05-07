package azurepim

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestAssignmentFromExpanded_mapsSharedRoleAndScopeFields(t *testing.T) {
	role := "Contributor"
	roleDefID := "/roleDefs/contributor"
	scopeName := "rg-prod"
	scopeID := "/subscriptions/abc/resourceGroups/rg-prod"
	scopeType := "Microsoft.Resources/subscriptions/resourceGroups"

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
