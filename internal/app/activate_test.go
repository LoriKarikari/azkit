package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func TestActivation_succeeds(t *testing.T) {
	dur := 2 * time.Hour
	start := time.Now().UTC().Truncate(time.Second)
	end := start.Add(dur)

	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1", ScopeID: "/sub/abc", ScopeName: "sub-prod"},
		},
	}
	act := &testActivator{result: &domain.ActivationResult{
		Role: "Contributor", ScopeID: "/sub/abc", ScopeName: "sub-prod",
		Duration: dur, StartedAt: start, ExpiresAt: end, Reason: "Deploy",
	}}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(context.Background(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: " Deploy ", Duration: dur,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" || got.Reason != "Deploy" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if act.calledPrincipalID != "user-1" {
		t.Fatalf("want principalID user-1, got %s", act.calledPrincipalID)
	}
	if act.calledReason != "Deploy" {
		t.Fatalf("want trimmed reason Deploy, got %q", act.calledReason)
	}
}

func TestActivation_missingScope(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(context.Background(), domain.ActivationRequest{
		Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingScope) {
		t.Fatalf("want missing scope error, got %v", err)
	}
}

func TestActivation_missingRole(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(context.Background(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingRole) {
		t.Fatalf("want missing role error, got %v", err)
	}
}

func TestActivation_missingReason(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(context.Background(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "   ", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingReason) {
		t.Fatalf("want missing reason error, got %v", err)
	}
}

func TestActivation_noMatchingAssignment(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{Role: "Reader", ScopeID: "/sub/def"},
		},
	}
	svc := app.NewActivationService(store, &testActivator{})
	_, err := svc.Activate(context.Background(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrEligibleNotFound) {
		t.Fatalf("want not found error, got %v", err)
	}
}

func TestActivation_invalidDuration(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(context.Background(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "Deploy", Duration: -1,
	})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != app.CodeInvalidDuration {
		t.Fatalf("want invalid duration, got %s", appErr.Code)
	}
}

type testActivator struct {
	result            *domain.ActivationResult
	err               error
	calledPrincipalID string
	calledReason      string
}

func (a *testActivator) Activate(_ context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	a.calledPrincipalID = target.Assignment.PrincipalID
	a.calledReason = target.Reason
	if a.err != nil {
		return nil, a.err
	}
	return a.result, nil
}
