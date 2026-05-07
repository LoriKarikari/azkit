package azurepim

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/google/uuid"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type roleAssignmentRequests interface {
	Create(context.Context, string, string, armauthorization.RoleAssignmentScheduleRequest) (armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse, error)
}

type ActivationStore struct {
	requests       roleAssignmentRequests
	now            func() time.Time
	newRequestName func() string
}

func NewActivationStore(cred azcore.TokenCredential) *ActivationStore {
	return newActivationStore(azureRoleAssignmentRequests{cred: cred}, time.Now, uuid.NewString)
}

func newActivationStore(requests roleAssignmentRequests, now func() time.Time, newRequestName func() string) *ActivationStore {
	return &ActivationStore{requests: requests, now: now, newRequestName: newRequestName}
}

func (a *ActivationStore) Activate(ctx context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	reqType := armauthorization.RequestTypeSelfActivate
	now := a.now().UTC()
	expType := armauthorization.TypeAfterDuration
	isoDur := durationToISO8601(target.Duration)
	linkedScheduleID := target.Assignment.ID

	parameters := armauthorization.RoleAssignmentScheduleRequest{
		Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
			PrincipalID:                     &target.Assignment.PrincipalID,
			RoleDefinitionID:                &target.Assignment.RoleDefID,
			RequestType:                     &reqType,
			Justification:                   &target.Reason,
			LinkedRoleEligibilityScheduleID: &linkedScheduleID,
			ScheduleInfo: &armauthorization.RoleAssignmentScheduleRequestPropertiesScheduleInfo{
				StartDateTime: &now,
				Expiration: &armauthorization.RoleAssignmentScheduleRequestPropertiesScheduleInfoExpiration{
					Type:     &expType,
					Duration: &isoDur,
				},
			},
		},
	}

	resp, err := a.requests.Create(ctx, target.Assignment.ScopeID, a.newRequestName(), parameters)
	if err != nil {
		return nil, azurePIMOperationError(err)
	}

	result := &domain.ActivationResult{
		Role:      target.Assignment.Role,
		ScopeID:   target.Assignment.ScopeID,
		ScopeName: target.Assignment.ScopeName,
		Duration:  target.Duration,
		StartedAt: now,
		ExpiresAt: now.Add(target.Duration),
		Reason:    target.Reason,
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

type azureRoleAssignmentRequests struct {
	cred azcore.TokenCredential
}

func (a azureRoleAssignmentRequests) Create(ctx context.Context, scope string, requestName string, parameters armauthorization.RoleAssignmentScheduleRequest) (armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse, error) {
	client, err := armauthorization.NewRoleAssignmentScheduleRequestsClient(a.cred, nil)
	if err != nil {
		return armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse{}, app.AuthFailed(err)
	}
	return client.Create(ctx, scope, requestName, parameters, nil)
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
