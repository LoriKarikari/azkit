package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/inmemory"
)

func TestListService_populated(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod"},
			{ID: "a2", Role: "Reader", ScopeType: domain.ScopeResourceGroup, ScopeName: "rg-dev-app"},
		},
	}
	svc := app.NewListService(store)
	got, err := svc.List(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 assignments, got %d", len(got))
	}
	if got[0].Role != "Contributor" || got[1].Role != "Reader" {
		t.Fatalf("roles mismatch: %+v", got)
	}
}

func TestListService_empty(t *testing.T) {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{},
	}
	svc := app.NewListService(store)
	got, err := svc.List(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 assignments, got %d", len(got))
	}
}

func TestListService_adapterError(t *testing.T) {
	store := &inmemory.EligibleAssignments{Err: assert.AnError}
	svc := app.NewListService(store)
	_, err := svc.List(t.Context())
	if err == nil {
		t.Fatal("want error, got nil")
	}
}
