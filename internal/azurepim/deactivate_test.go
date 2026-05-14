package azurepim

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestDeactivationStore_createsSelfDeactivateRequest(t *testing.T) {
	requests := &fakeRoleAssignmentRequests{
		response: armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse{
			RoleAssignmentScheduleRequest: armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{},
			},
		},
	}
	store := newDeactivationStore(requests, func() string { return "deact-request-1" }, nil)

	result, err := store.Deactivate(t.Context(), domain.ActiveAssignment{
		ID:          "inst-1",
		PrincipalID: "user-1",
		Role:        "Contributor",
		RoleDefID:   "/roleDefs/111",
		ScopeID:     "/subscriptions/abc",
		ScopeName:   "sub-prod",
	}, "incident resolved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requests.scope != "/subscriptions/abc" {
		t.Fatalf("want scope /subscriptions/abc, got %s", requests.scope)
	}
	if requests.name != "deact-request-1" {
		t.Fatalf("want request name deact-request-1, got %s", requests.name)
	}
	props := requests.parameters.Properties
	if props == nil {
		t.Fatal("missing properties")
		return
	}
	if *props.RequestType != armauthorization.RequestTypeSelfDeactivate {
		t.Fatalf("want SelfDeactivate, got %s", *props.RequestType)
	}
	if *props.Justification != "incident resolved" {
		t.Fatalf("want reason 'incident resolved', got %s", *props.Justification)
	}
	if *props.RoleDefinitionID != "/roleDefs/111" {
		t.Fatalf("want role definition /roleDefs/111, got %s", *props.RoleDefinitionID)
	}
	if *props.PrincipalID != "user-1" {
		t.Fatalf("want principal user-1, got %s", *props.PrincipalID)
	}
	if *props.TargetRoleAssignmentScheduleInstanceID != "inst-1" {
		t.Fatalf("want target schedule instance inst-1, got %s", *props.TargetRoleAssignmentScheduleInstanceID)
	}

	if result.Role != "Contributor" {
		t.Fatalf("want Contributor, got %s", result.Role)
	}
	if result.ScopeName != "sub-prod" {
		t.Fatalf("want sub-prod, got %s", result.ScopeName)
	}
	if result.AssignmentID != "inst-1" {
		t.Fatalf("want inst-1, got %s", result.AssignmentID)
	}
	if result.Status != domain.DeactivationRequested {
		t.Fatalf("want deactivation_requested, got %s", result.Status)
	}
}

func TestDeactivationStore_omitsBlankJustification(t *testing.T) {
	requests := &fakeRoleAssignmentRequests{
		response: okDeactivationResponse(),
	}
	store := newDeactivationStore(requests, func() string { return "req-1" }, nil)

	_, err := store.Deactivate(t.Context(), domain.ActiveAssignment{
		ID: "inst-1", PrincipalID: "user-1", RoleDefID: "/roleDefs/111", ScopeID: "/sub/abc",
	}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	props := requests.parameters.Properties
	if props == nil {
		t.Fatal("missing properties")
		return
	}
	if props.Justification != nil {
		t.Fatalf("want no justification, got %s", *props.Justification)
	}
}

func TestDeactivationStore_mapsPermissionDenied(t *testing.T) {
	requests := &fakeRoleAssignmentRequests{err: errors.New("403 AuthorizationFailed")}
	store := newDeactivationStore(requests, func() string { return "req-1" }, nil)

	_, err := store.Deactivate(t.Context(), domain.ActiveAssignment{
		ID: "inst-1", PrincipalID: "user-1", RoleDefID: "/roleDefs/111", ScopeID: "/sub/abc",
	}, "")
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodePermissionDenied {
		t.Fatalf("want permission denied, got %s", appErr.Code)
	}
}

func okDeactivationResponse() armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse {
	return armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse{
		RoleAssignmentScheduleRequest: armauthorization.RoleAssignmentScheduleRequest{
			Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{},
		},
	}
}
