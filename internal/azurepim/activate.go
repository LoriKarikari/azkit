package azurepim

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/google/uuid"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

var (
	_ app.ActivationStore    = (*ActivationStore)(nil)
	_ roleAssignmentRequests = (*azureRoleAssignmentRequests)(nil)
)

type roleAssignmentRequests interface {
	Create(
		context.Context,
		string,
		string,
		armauthorization.RoleAssignmentScheduleRequest,
	) (armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse, error)
}

type ActivationStore struct {
	requests       roleAssignmentRequests
	now            func() time.Time
	newRequestName func() string
	log            *slog.Logger
}

func NewActivationStore(cred azcore.TokenCredential, log *slog.Logger) *ActivationStore {
	return newActivationStore(
		azureRoleAssignmentRequests{cred: cred},
		time.Now,
		uuid.NewString,
		log,
	)
}

func newActivationStore(
	requests roleAssignmentRequests,
	now func() time.Time,
	newRequestName func() string,
	log *slog.Logger,
) *ActivationStore {
	return &ActivationStore{
		requests:       requests,
		now:            now,
		newRequestName: newRequestName,
		log:            logger(log),
	}
}

func (a *ActivationStore) Activate(
	ctx context.Context,
	target domain.ActivationTarget,
) (*domain.ActivationResult, error) {
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

	requestName := a.newRequestName()
	a.log.Debug(
		"creating activation request",
		slog.String("request_name", requestName),
		slog.String("scope", target.Assignment.ScopeID),
		slog.String("role", target.Assignment.Role),
	)
	resp, err := a.requests.Create(
		ctx,
		target.Assignment.ScopeID,
		requestName,
		parameters,
	)
	if err != nil {
		a.log.Debug(
			"activation request failed",
			slog.String("request_name", requestName),
			slog.String("scope", target.Assignment.ScopeID),
			slog.Any("error", err),
		)
		return nil, azurePIMOperationError(err)
	}
	a.log.Debug(
		"activation request created",
		slog.String("request_name", requestName),
		slog.String("scope", target.Assignment.ScopeID),
	)

	result := &domain.ActivationResult{
		Role:      target.Assignment.Role,
		RoleDefID: target.Assignment.RoleDefID,
		ScopeID:   target.Assignment.ScopeID,
		ScopeName: target.Assignment.ScopeName,
		Duration:  target.Duration,
		StartedAt: now,
		ExpiresAt: now.Add(target.Duration),
		Reason:    target.Reason,
		Outcome:   domain.ActivationRequested,
	}
	if resp.Properties == nil || resp.Properties.ExpandedProperties == nil {
		return result, nil
	}
	if role := resp.Properties.ExpandedProperties.RoleDefinition; role != nil && role.DisplayName != nil {
		result.Role = *role.DisplayName
	}
	if scope := resp.Properties.ExpandedProperties.Scope; scope != nil && scope.DisplayName != nil {
		result.ScopeName = *scope.DisplayName
	}
	return result, nil
}

type azureRoleAssignmentRequests struct {
	cred azcore.TokenCredential
}

func (a azureRoleAssignmentRequests) Create(
	ctx context.Context,
	scope string,
	requestName string,
	parameters armauthorization.RoleAssignmentScheduleRequest,
) (armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse, error) {
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
