package interactive_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/inmemory"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

func TestActivateReturnsNotFoundForEmptyAssignments(t *testing.T) {
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, nil, &fakeActivator{})

	_, err := interactive.Activate(
		t.Context(),
		[]domain.EligibleAssignment{},
		svc,
		nil,
		interactive.ActivationInput{
			Reason:      "deploy",
			Duration:    30 * time.Minute,
			AutoConfirm: true,
		},
	)
	if !errors.Is(err, app.ErrEligibleNotFound) {
		t.Fatalf("want ErrEligibleNotFound, got %v", err)
	}
}

func TestActivateSkipsFormWhenInputsAreComplete(t *testing.T) {
	activator := &fakeActivator{}
	svc := app.NewActivationService(&inmemory.EligibleAssignments{}, nil, activator)

	result, err := interactive.Activate(
		t.Context(),
		[]domain.EligibleAssignment{
			{
				Role:      "Contributor",
				ScopeID:   "/subscriptions/abc",
				ScopeName: "sub-prod",
			},
		},
		svc,
		nil,
		interactive.ActivationInput{
			Reason:      "deploy",
			Duration:    30 * time.Minute,
			AutoConfirm: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Duration != 30*time.Minute {
		t.Fatalf("want 30m duration, got %s", result.Duration)
	}
	if result.Reason != "deploy" {
		t.Fatalf("want deploy reason, got %q", result.Reason)
	}
}

type fakeActivator struct{}

func (a *fakeActivator) Activate(_ context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	return &domain.ActivationResult{
		Role:      target.Assignment.Role,
		ScopeID:   target.Assignment.ScopeID,
		ScopeName: target.Assignment.ScopeName,
		Duration:  target.Duration,
		Reason:    target.Reason,
	}, nil
}
