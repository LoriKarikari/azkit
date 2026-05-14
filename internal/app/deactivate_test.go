package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func TestDeactivation_foundAndDeactivated(t *testing.T) {
	active := &inmemory.ActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{
				ID:        "inst-1",
				Role:      "Contributor",
				ScopeID:   "/subscriptions/abc",
				ScopeName: "sub-prod",
			},
		},
	}
	deactivator := &testDeactivator{
		result: &domain.DeactivationResult{
			Role:         "Contributor",
			ScopeID:      "/subscriptions/abc",
			ScopeName:    "sub-prod",
			AssignmentID: "inst-1",
			Status:       domain.DeactivationRequested,
		},
	}
	svc := app.NewDeactivationService(active, deactivator)

	got, err := svc.Deactivate(t.Context(), "inst-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "Contributor" {
		t.Fatalf("want Contributor, got %s", got.Role)
	}
	if got.AssignmentID != "inst-1" {
		t.Fatalf("want inst-1, got %s", got.AssignmentID)
	}
	if got.Status != domain.DeactivationRequested {
		t.Fatalf("want deactivation_requested, got %s", got.Status)
	}
}

func TestDeactivation_withReason(t *testing.T) {
	active := &inmemory.ActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{ID: "inst-1", Role: "Contributor", ScopeID: "/sub/abc"},
		},
	}
	deactivator := &testDeactivator{
		result: &domain.DeactivationResult{
			Role:   "Contributor",
			Status: domain.DeactivationRequested,
		},
	}
	svc := app.NewDeactivationService(active, deactivator)

	_, err := svc.Deactivate(t.Context(), "inst-1", "incident resolved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deactivator.reason != "incident resolved" {
		t.Fatalf("want reason 'incident resolved', got %q", deactivator.reason)
	}
}

func TestDeactivation_assignmentNotFound(t *testing.T) {
	active := &inmemory.ActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{ID: "inst-1", Role: "Contributor"},
		},
	}
	svc := app.NewDeactivationService(active, &testDeactivator{})

	_, err := svc.Deactivate(t.Context(), "inst-nonexistent", "")
	if !errors.Is(err, app.ErrActiveNotFound) {
		t.Fatalf("want ErrActiveNotFound, got %v", err)
	}
}

func TestDeactivation_emptyAssignmentList(t *testing.T) {
	active := &inmemory.ActiveAssignments{}
	svc := app.NewDeactivationService(active, &testDeactivator{})

	_, err := svc.Deactivate(t.Context(), "inst-1", "")
	if !errors.Is(err, app.ErrActiveNotFound) {
		t.Fatalf("want ErrActiveNotFound, got %v", err)
	}
}

func TestDeactivation_listActiveError(t *testing.T) {
	active := &inmemory.ActiveAssignments{
		Err: errors.New("azure down"),
	}
	svc := app.NewDeactivationService(active, &testDeactivator{})

	_, err := svc.Deactivate(t.Context(), "inst-1", "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !errors.Is(err, active.Err) {
		t.Fatalf("want azure error, got %v", err)
	}
}

func TestDeactivation_deactivatorError(t *testing.T) {
	active := &inmemory.ActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{ID: "inst-1", Role: "Contributor"},
		},
	}
	deactivator := &testDeactivator{
		err: errors.New("azure rejected"),
	}
	svc := app.NewDeactivationService(active, deactivator)

	_, err := svc.Deactivate(t.Context(), "inst-1", "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

type testDeactivator struct {
	result *domain.DeactivationResult
	reason string
	err    error
}

func (d *testDeactivator) Deactivate(
	_ context.Context,
	assignment domain.ActiveAssignment,
	reason string,
) (*domain.DeactivationResult, error) {
	d.reason = reason
	if d.err != nil {
		return nil, d.err
	}
	return d.result, nil
}
