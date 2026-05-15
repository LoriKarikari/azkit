package azurepim

import (
	"context"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/google/uuid"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

var _ app.DeactivationStore = (*DeactivationStore)(nil)

type DeactivationStore struct {
	requests       roleAssignmentRequests
	newRequestName func() string
	log            *slog.Logger
}

func NewDeactivationStore(cred azcore.TokenCredential, log *slog.Logger) *DeactivationStore {
	return newDeactivationStore(
		azureRoleAssignmentRequests{cred: cred},
		uuid.NewString,
		log,
	)
}

func newDeactivationStore(
	requests roleAssignmentRequests,
	newRequestName func() string,
	log *slog.Logger,
) *DeactivationStore {
	return &DeactivationStore{
		requests:       requests,
		newRequestName: newRequestName,
		log:            logger(log),
	}
}

func (d *DeactivationStore) Deactivate(
	ctx context.Context,
	assignment domain.ActiveAssignment,
	reason string,
) (*domain.DeactivationResult, error) {
	reqType := armauthorization.RequestTypeSelfDeactivate
	requestName := d.newRequestName()

	d.log.Debug(
		"creating deactivation request",
		slog.String("request_name", requestName),
		slog.String("scope", assignment.ScopeID),
		slog.String("role", assignment.Role),
	)

	parameters := armauthorization.RoleAssignmentScheduleRequest{
		Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
			RoleDefinitionID: &assignment.RoleDefID,
			PrincipalID:      &assignment.PrincipalID,
			RequestType:      &reqType,
		},
	}
	if reason != "" {
		parameters.Properties.Justification = &reason
	}

	_, err := d.requests.Create(ctx, assignment.ScopeID, requestName, parameters)
	if err != nil {
		d.log.Debug(
			"deactivation request failed",
			slog.String("request_name", requestName),
			slog.String("scope", assignment.ScopeID),
			slog.Any("error", err),
		)
		return nil, azurePIMOperationError(err)
	}

	d.log.Debug(
		"deactivation request created",
		slog.String("request_name", requestName),
		slog.String("scope", assignment.ScopeID),
	)

	return &domain.DeactivationResult{
		Role:         assignment.Role,
		ScopeID:      assignment.ScopeID,
		ScopeName:    assignment.ScopeName,
		ScopeType:    assignment.ScopeType,
		AssignmentID: assignment.ID,
		Status:       domain.DeactivationRequested,
	}, nil
}
