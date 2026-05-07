package azurepim

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/google/uuid"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivationStore struct {
	cred azcore.TokenCredential
}

func NewActivationStore(cred azcore.TokenCredential) *ActivationStore {
	return &ActivationStore{cred: cred}
}

func (a *ActivationStore) Activate(ctx context.Context, principalID, roleDefID, scope, reason string, duration time.Duration) (*domain.ActivationResult, error) {
	client, err := armauthorization.NewRoleAssignmentScheduleRequestsClient(a.cred, nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	reqName := uuid.NewString()
	reqType := armauthorization.RequestTypeSelfActivate
	now := time.Now().UTC()
	expType := armauthorization.TypeAfterDuration
	isoDur := durationToISO8601(duration)

	parameters := armauthorization.RoleAssignmentScheduleRequest{
		Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
			PrincipalID:                     &principalID,
			RoleDefinitionID:                &roleDefID,
			RequestType:                     &reqType,
			Justification:                   &reason,
			LinkedRoleEligibilityScheduleID: nil,
			ScheduleInfo: &armauthorization.RoleAssignmentScheduleRequestPropertiesScheduleInfo{
				StartDateTime: &now,
				Expiration: &armauthorization.RoleAssignmentScheduleRequestPropertiesScheduleInfoExpiration{
					Type:     &expType,
					Duration: &isoDur,
				},
			},
		},
	}

	resp, err := client.Create(ctx, scope, reqName, parameters, nil)
	if err != nil {
		if strings.Contains(err.Error(), "AuthorizationFailed") || strings.Contains(err.Error(), "403") {
			return nil, app.PermissionDenied(err)
		}
		return nil, app.AzureAPIError(err)
	}

	result := &domain.ActivationResult{
		Role:      "",
		ScopeID:   scope,
		ScopeName: "",
		Duration:  duration,
		StartedAt: now,
		ExpiresAt: now.Add(duration),
		Reason:    reason,
	}
	if resp.Properties != nil {
		if resp.Properties.ExpandedProperties != nil {
			if resp.Properties.ExpandedProperties.RoleDefinition != nil && resp.Properties.ExpandedProperties.RoleDefinition.DisplayName != nil {
				result.Role = *resp.Properties.ExpandedProperties.RoleDefinition.DisplayName
			}
			if resp.Properties.ExpandedProperties.Scope != nil && resp.Properties.ExpandedProperties.Scope.DisplayName != nil {
				result.ScopeName = *resp.Properties.ExpandedProperties.Scope.DisplayName
			}
		}
	}
	return result, nil
}

func durationToISO8601(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if s > 0 {
		return fmt.Sprintf("PT%dH%dM%dS", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("PT%dH%dM", h, m)
	}
	return fmt.Sprintf("PT%dH", h)
}
