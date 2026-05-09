package azurepim

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestActivationStore_createsSelfActivateRequest(t *testing.T) {
	roleName := "Contributor"
	scopeName := "sub-prod"
	requests := &fakeRoleAssignmentRequests{
		response: armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse{
			RoleAssignmentScheduleRequest: armauthorization.RoleAssignmentScheduleRequest{
				Properties: &armauthorization.RoleAssignmentScheduleRequestProperties{
					ExpandedProperties: &armauthorization.ExpandedProperties{
						RoleDefinition: &armauthorization.ExpandedPropertiesRoleDefinition{DisplayName: &roleName},
						Scope:          &armauthorization.ExpandedPropertiesScope{DisplayName: &scopeName},
					},
				},
			},
		},
	}
	store := newActivationStore(requests, func() time.Time { return activationTime("2026-05-07T18:00:00Z") }, func() string { return "request-id" }, nil)

	result, err := store.Activate(t.Context(), domain.ActivationTarget{
		Assignment: domain.EligibleAssignment{
			ID:          "eligibility-schedule-id",
			PrincipalID: "user-1",
			RoleDefID:   "/roleDefs/111",
			Role:        "Contributor",
			ScopeID:     "/subscriptions/abc",
			ScopeName:   "sub-prod",
		},
		Reason:   "Deploy",
		Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requests.scope != "/subscriptions/abc" {
		t.Fatalf("want scope /subscriptions/abc, got %s", requests.scope)
	}
	if requests.name != "request-id" {
		t.Fatalf("want request name request-id, got %s", requests.name)
	}
	props := requests.parameters.Properties
	if props == nil {
		t.Fatal("missing properties")
	}
	if *props.PrincipalID != "user-1" {
		t.Fatalf("want principal user-1, got %s", *props.PrincipalID)
	}
	if *props.RoleDefinitionID != "/roleDefs/111" {
		t.Fatalf("want role definition /roleDefs/111, got %s", *props.RoleDefinitionID)
	}
	if *props.RequestType != armauthorization.RequestTypeSelfActivate {
		t.Fatalf("want SelfActivate, got %s", *props.RequestType)
	}
	if *props.Justification != "Deploy" {
		t.Fatalf("want reason Deploy, got %s", *props.Justification)
	}
	if *props.LinkedRoleEligibilityScheduleID != "eligibility-schedule-id" {
		t.Fatalf("want linked schedule eligibility-schedule-id, got %s", *props.LinkedRoleEligibilityScheduleID)
	}
	if props.ScheduleInfo == nil || props.ScheduleInfo.Expiration == nil {
		t.Fatal("missing schedule info")
	}
	if *props.ScheduleInfo.Expiration.Duration != "PT2H" {
		t.Fatalf("want duration PT2H, got %s", *props.ScheduleInfo.Expiration.Duration)
	}
	if result.Role != "Contributor" || result.ScopeName != "sub-prod" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestActivationStore_mapsPermissionDenied(t *testing.T) {
	requests := &fakeRoleAssignmentRequests{err: errors.New("403 AuthorizationFailed")}
	store := newActivationStore(requests, time.Now, func() string { return "request-id" }, nil)

	_, err := store.Activate(t.Context(), domain.ActivationTarget{
		Assignment: domain.EligibleAssignment{PrincipalID: "user-1", RoleDefID: "/roleDefs/111", ScopeID: "/subscriptions/abc"},
		Reason:     "Deploy",
		Duration:   2 * time.Hour,
	})
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != app.CodePermissionDenied {
		t.Fatalf("want permission denied, got %s", appErr.Code)
	}
}

func activationTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

func TestDurationToISO8601(t *testing.T) {
	cases := map[time.Duration]string{
		2 * time.Hour:                           "PT2H",
		90 * time.Minute:                        "PT1H30M",
		(time.Hour + time.Minute + time.Second): "PT1H1M1S",
	}
	for input, want := range cases {
		if got := durationToISO8601(input); got != want {
			t.Fatalf("%s: want %s, got %s", input, want, got)
		}
	}
}

type fakeRoleAssignmentRequests struct {
	scope      string
	name       string
	parameters armauthorization.RoleAssignmentScheduleRequest
	response   armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse
	err        error
}

func (f *fakeRoleAssignmentRequests) Create(_ context.Context, scope string, name string, parameters armauthorization.RoleAssignmentScheduleRequest) (armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse, error) {
	f.scope = scope
	f.name = name
	f.parameters = parameters
	if f.err != nil {
		return armauthorization.RoleAssignmentScheduleRequestsClientCreateResponse{}, f.err
	}
	return f.response, nil
}
