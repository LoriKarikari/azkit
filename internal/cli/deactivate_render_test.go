package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRenderDeactivationJSON(t *testing.T) {
	result := &domain.DeactivationResult{
		Role:         "Contributor",
		ScopeType:    domain.ScopeSubscription,
		ScopeID:      "/subscriptions/abc",
		ScopeName:    "sub-prod",
		AssignmentID: "inst-1",
		Status:       domain.DeactivationRequested,
	}

	got := renderDeactivationJSON(result)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if decoded["role"] != "Contributor" {
		t.Fatalf("want role Contributor, got %v", decoded["role"])
	}
	if decoded["scope_type"] != "subscription" {
		t.Fatalf("want scope_type subscription, got %v", decoded["scope_type"])
	}
	if decoded["scope_id"] != "/subscriptions/abc" {
		t.Fatalf("want scope_id, got %v", decoded["scope_id"])
	}
	if decoded["scope_name"] != "sub-prod" {
		t.Fatalf("want scope_name sub-prod, got %v", decoded["scope_name"])
	}
	if decoded["assignment_id"] != "inst-1" {
		t.Fatalf("want assignment_id inst-1, got %v", decoded["assignment_id"])
	}
	if decoded["status"] != "deactivation_requested" {
		t.Fatalf("want status deactivation_requested, got %v", decoded["status"])
	}
}

func TestRenderDeactivationHuman(t *testing.T) {
	result := &domain.DeactivationResult{
		Role:         "Contributor",
		ScopeName:    "sub-prod",
		ScopeType:    domain.ScopeSubscription,
		ScopeID:      "/subscriptions/abc",
		AssignmentID: "inst-1",
		Status:       domain.DeactivationRequested,
	}

	got := renderDeactivationHuman(result)
	if !strings.Contains(got, "Contributor") {
		t.Fatalf("want role in output, got: %s", got)
	}
	if !strings.Contains(got, "sub-prod") {
		t.Fatalf("want scope name in output, got: %s", got)
	}
	if !strings.Contains(got, "inst-1") {
		t.Fatalf("want assignment ID in output, got: %s", got)
	}
}
