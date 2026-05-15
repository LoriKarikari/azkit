package interactive_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
	"github.com/LoriKarikari/pimctl/internal/interactive"
)

func TestDeactivateReturnsNotFoundForEmptyAssignments(t *testing.T) {
	svc := app.NewDeactivationService(&inmemory.ActiveAssignments{}, &stubDeactivator{})
	_, err := interactive.Deactivate(t.Context(), nil, svc, interactive.DeactivationInput{AutoConfirm: true})
	if !errors.Is(err, app.ErrActiveAssignmentNotFound) {
		t.Fatalf("want ErrActiveAssignmentNotFound, got %v", err)
	}
}

func TestDeactivateAutoSelectsSingleAssignment(t *testing.T) {
	assignment := domain.ActiveAssignment{
		ID:        "inst-1",
		Role:      "Contributor",
		ScopeID:   "/subscriptions/abc",
		ScopeName: "sub-prod",
	}
	deactivator := &stubDeactivator{
		result: &domain.DeactivationResult{
			Role:         "Contributor",
			ScopeID:      "/subscriptions/abc",
			ScopeName:    "sub-prod",
			AssignmentID: "inst-1",
			Status:       domain.DeactivationRequested,
		},
	}
	svc := app.NewDeactivationService(
		&inmemory.ActiveAssignments{Assignments: []domain.ActiveAssignment{assignment}},
		deactivator,
	)
	result, err := interactive.Deactivate(t.Context(), []domain.ActiveAssignment{assignment}, svc, interactive.DeactivationInput{
		Reason:      "done",
		AutoConfirm: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AssignmentID != "inst-1" {
		t.Fatalf("want inst-1, got %s", result.AssignmentID)
	}
}

type stubDeactivator struct {
	result *domain.DeactivationResult
	err    error
}

func (s *stubDeactivator) Deactivate(_ context.Context, _ domain.ActiveAssignment, _ string) (*domain.DeactivationResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}
