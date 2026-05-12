package app_test

import (
	"errors"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func TestActivateResolved_validTarget(t *testing.T) {
	assignment := domain.EligibleAssignment{
		ID:      "sched-1",
		Role:    "Contributor",
		ScopeID: "/sub/abc",
	}
	activator := &testActivator{
		result: &domain.ActivationResult{
			Role:      "Contributor",
			ScopeID:   "/sub/abc",
			ScopeName: "sub-prod",
			Duration:  2 * time.Hour,
			Reason:    "deploy",
		},
	}
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, activator)

	got, err := svc.ActivateResolved(t.Context(), domain.ActivationTarget{
		Assignment: assignment,
		Reason:     "deploy",
		Duration:   2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("want Contributor, got %s", got.Role)
	}
}

func TestActivateResolved_missingReason(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.ActivateResolved(t.Context(), domain.ActivationTarget{
		Assignment: domain.EligibleAssignment{Role: "Reader", ScopeID: "/sub/1"},
		Reason:     "   ",
		Duration:   2 * time.Hour,
	})
	if !errors.Is(err, app.ErrMissingReason) {
		t.Fatalf("want ErrMissingReason, got %v", err)
	}
}

func TestActivateResolved_invalidDuration(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, &testActivator{})
	_, err := svc.ActivateResolved(t.Context(), domain.ActivationTarget{
		Assignment: domain.EligibleAssignment{Role: "Reader", ScopeID: "/sub/1"},
		Reason:     "deploy",
		Duration:   -1,
	})
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeInvalidDuration {
		t.Fatalf("want invalid duration error, got %v", err)
	}
}
