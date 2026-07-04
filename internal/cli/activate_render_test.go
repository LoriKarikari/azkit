package cli

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRenderActivationJSON(t *testing.T) {
	result := &domain.ActivationResult{
		Role:      "Contributor",
		ScopeID:   "/subscriptions/abc",
		ScopeName: "sub-prod",
		Duration:  2 * time.Hour,
		StartedAt: mustTime("2026-05-07T18:00:00Z"),
		ExpiresAt: mustTime("2026-05-07T20:00:00Z"),
		Reason:    "Deploy",
	}

	got := renderActivationJSON(result)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if decoded["role"] != "Contributor" {
		t.Fatalf("want role Contributor, got %v", decoded["role"])
	}
	if decoded["duration"] != "2h0m0s" {
		t.Fatalf("want duration 2h0m0s, got %v", decoded["duration"])
	}
	if decoded["started_at"] != "2026-05-07T18:00:00Z" {
		t.Fatalf("want started_at, got %v", decoded["started_at"])
	}
	if decoded["expires_at"] != "2026-05-07T20:00:00Z" {
		t.Fatalf("want expires_at, got %v", decoded["expires_at"])
	}
	if decoded["reason"] != "Deploy" {
		t.Fatalf("want reason Deploy, got %v", decoded["reason"])
	}

	scope, ok := decoded["scope"].(map[string]any)
	if !ok {
		t.Fatalf("want scope object, got %T", decoded["scope"])
	}
	if scope["id"] != "/subscriptions/abc" || scope["name"] != "sub-prod" {
		t.Fatalf("unexpected scope: %+v", scope)
	}
}

func TestRenderActivationJSONAlreadyActive(t *testing.T) {
	result := &domain.ActivationResult{
		Role:      "Contributor",
		ScopeID:   "/subscriptions/abc",
		ScopeName: "sub-prod",
		Duration:  2 * time.Hour,
		StartedAt: mustTime("2026-05-07T18:00:00Z"),
		ExpiresAt: mustTime("2026-05-07T20:00:00Z"),
		Outcome:   domain.ActivationAlreadyActive,
	}

	got := renderActivationJSON(result)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded["outcome"] != "already_active" {
		t.Fatalf("want outcome already_active, got %v", decoded["outcome"])
	}
}

func TestRenderActivationHuman(t *testing.T) {
	result := &domain.ActivationResult{
		Role:      "Contributor",
		ScopeName: "sub-prod",
		Duration:  2 * time.Hour,
		ExpiresAt: mustTime("2026-05-07T20:00:00Z"),
		Reason:    "Deploy",
	}

	got := renderActivationHuman(result)
	want := "✓ Activated Contributor on sub-prod\nDuration:  2h0m0s\nExpires:   2026-05-07 20:00 UTC\nReason:    Deploy\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderActivationHumanPending(t *testing.T) {
	result := &domain.ActivationResult{
		Role:      "Contributor",
		ScopeName: "sub-prod",
		Duration:  2 * time.Hour,
		ExpiresAt: mustTime("2026-05-07T20:00:00Z"),
		Reason:    "Deploy",
		Outcome:   domain.ActivationPending,
	}

	got := renderActivationHuman(result)
	want := "✓ Activation requested for Contributor on sub-prod\nDuration:  2h0m0s\nExpires:   2026-05-07 20:00 UTC\nReason:    Deploy\nStatus:    Waiting for Azure to report this assignment as active\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderActivationHumanAlreadyActive(t *testing.T) {
	result := &domain.ActivationResult{
		Role:      "Contributor",
		ScopeName: "sub-prod",
		Duration:  2 * time.Hour,
		ExpiresAt: mustTime("2026-05-07T20:00:00Z"),
		Outcome:   domain.ActivationAlreadyActive,
	}

	got := renderActivationHuman(result)
	want := "✓ Already active: Contributor on sub-prod\nDuration:  2h0m0s\nExpires:   2026-05-07 20:00 UTC\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func mustTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
