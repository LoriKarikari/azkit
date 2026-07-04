package azurepim

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type assignmentMapping struct {
	Role      string
	RoleDefID string
	ScopeID   string
	ScopeName string
	ScopeType domain.ScopeType
}

func assignmentFromExpanded(expanded *armauthorization.ExpandedProperties) assignmentMapping {
	m := assignmentMapping{ScopeType: domain.ScopeSubscription}
	if expanded == nil {
		return m
	}
	if expanded.RoleDefinition != nil {
		if expanded.RoleDefinition.DisplayName != nil {
			m.Role = *expanded.RoleDefinition.DisplayName
		}
		if expanded.RoleDefinition.ID != nil {
			m.RoleDefID = *expanded.RoleDefinition.ID
		}
	}
	if expanded.Scope != nil {
		if expanded.Scope.DisplayName != nil {
			m.ScopeName = *expanded.Scope.DisplayName
		}
		if expanded.Scope.ID != nil {
			m.ScopeID = *expanded.Scope.ID
		}
		if expanded.Scope.Type != nil {
			m.ScopeType = scopeTypeFromAzure(*expanded.Scope.Type)
		}
	}
	return m
}

func scopeTypeFromAzure(typeStr string) domain.ScopeType {
	lower := strings.ToLower(typeStr)
	if strings.Contains(lower, "managementgroup") || strings.Contains(lower, "managementgroups") {
		return domain.ScopeManagementGroup
	}
	if strings.Contains(lower, "resourcegroup") || strings.Contains(lower, "resourcegroups") {
		return domain.ScopeResourceGroup
	}
	return domain.ScopeSubscription
}
