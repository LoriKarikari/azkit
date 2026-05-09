package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func TestActivation_byScopeID(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1", ScopeID: "/sub/abc", ScopeName: "sub-prod"},
		},
	}
	act := &testActivator{result: okResult(t, 2*time.Hour)}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(t.Context(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestActivation_bySubscriptionID(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1",
				ScopeID: "/subscriptions/00000000-0000-0000-0000-000000000000", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod"},
		},
	}
	act := &testActivator{result: okResult(t, 2*time.Hour)}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "00000000-0000-0000-0000-000000000000", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestActivation_bySubscriptionName(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1",
				ScopeID: "/subscriptions/abc", ScopeType: domain.ScopeSubscription, ScopeName: "Production Platform"},
		},
	}
	act := &testActivator{result: okResult(t, 2*time.Hour)}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "Production Platform", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestActivation_byResourceGroup(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1",
				ScopeID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg", ScopeType: domain.ScopeResourceGroup, ScopeName: "prod-rg"},
		},
	}
	act := &testActivator{result: okResult(t, 2*time.Hour)}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "00000000-0000-0000-0000-000000000000", ResourceGroup: "prod-rg", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestActivation_byResourceGroupWithSubscriptionName(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "sched-1", Role: "Contributor", RoleDefID: "/roleDefs/111", PrincipalID: "user-1",
				ScopeID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg", ScopeType: domain.ScopeResourceGroup, ScopeName: "prod-rg",
				SubscriptionID: "00000000-0000-0000-0000-000000000000", SubscriptionName: "Production Platform"},
		},
	}
	act := &testActivator{result: okResult(t, 2*time.Hour)}
	svc := app.NewActivationService(store, act)

	got, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "production platform", ResourceGroup: "prod-rg", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestActivation_unknownSubscription(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ScopeID: "/subscriptions/abc", ScopeType: domain.ScopeSubscription, ScopeName: "Production Platform", SubscriptionID: "abc", SubscriptionName: "Production Platform"},
		},
	}
	svc := app.NewActivationService(store, &testActivator{})

	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "nonexistent", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeUnknownSubscription {
		t.Fatalf("want unknown subscription, got %v", err)
	}
	if !strings.Contains(appErr.Message, "Production Platform") {
		t.Fatalf("want subscription suggestion, got %q", appErr.Message)
	}
}

func TestActivation_unknownResourceGroup(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ScopeID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg", ScopeType: domain.ScopeResourceGroup, ScopeName: "prod-rg"},
		},
	}
	svc := app.NewActivationService(store, &testActivator{})

	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "00000000-0000-0000-0000-000000000000", ResourceGroup: "missing-rg", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeUnknownResourceGroup {
		t.Fatalf("want unknown resource group, got %v", err)
	}
	if !strings.Contains(appErr.Message, "prod-rg") {
		t.Fatalf("want resource group suggestion, got %q", appErr.Message)
	}
}

func TestActivation_ambiguousSubscription(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ScopeID: "/subscriptions/abc", ScopeType: domain.ScopeSubscription, ScopeName: "PROD"},
			{ScopeID: "/subscriptions/def", ScopeType: domain.ScopeSubscription, ScopeName: "prod"},
		},
	}
	svc := app.NewActivationService(store, &testActivator{})

	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Subscription: "prod", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeAmbiguousSubscription {
		t.Fatalf("want ambiguous subscription, got %v", err)
	}
}

func TestActivation_conflictingSelectors(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})

	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Subscription: "abc", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrConflictingSelectors) {
		t.Fatalf("want conflicting selectors error, got %v", err)
	}
}

func TestActivation_missingScope(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingScope) {
		t.Fatalf("want missing scope error, got %v", err)
	}
}

func TestActivation_missingRole(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingRole) {
		t.Fatalf("want missing role error, got %v", err)
	}
}

func TestActivation_missingReason(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
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
	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "Deploy", Duration: 2 * time.Hour,
	})
	if !errors.Is(err, app.ErrEligibleNotFound) {
		t.Fatalf("want not found error, got %v", err)
	}
}

func TestActivation_invalidDuration(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.Activate(t.Context(), domain.ActivationRequest{
		ScopeID: "/sub/abc", Role: "Contributor", Reason: "Deploy", Duration: -1,
	})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeInvalidDuration {
		t.Fatalf("want invalid duration, got %v", err)
	}
}

func okResult(t *testing.T, dur time.Duration) *domain.ActivationResult {
	t.Helper()
	start := time.Now().UTC().Truncate(time.Second)
	return &domain.ActivationResult{
		Role: "Contributor", ScopeName: "sub-prod", ScopeID: "/sub/abc",
		Duration: dur, StartedAt: start, ExpiresAt: start.Add(dur), Reason: "Deploy",
	}
}

type testActivator struct {
	result *domain.ActivationResult
	err    error
}

func (a *testActivator) Activate(_ context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.result, nil
}
