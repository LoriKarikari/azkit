package cli_test

import (
	"encoding/json"
	"testing"

	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRenderStatusHuman_populated(t *testing.T) {
	as := []domain.ActiveAssignment{
		{ID: "s1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", EndTime: at("2026-05-07T20:00:00Z"), Status: domain.ActiveAssignmentActive},
		{ID: "s2", Role: "Reader", ScopeType: domain.ScopeResourceGroup, ScopeName: "rg-dev-app", EndTime: at("2026-05-08T09:30:00Z"), Status: domain.ActiveAssignmentActive},
	}
	got := cli.RenderStatusHuman(as, false)
	want := "ROLE         TYPE            NAME        STATUS  EXPIRES\nContributor  subscription    sub-prod    Active  2026-05-07 20:00 UTC\nReader       resource_group  rg-dev-app  Active  2026-05-08 09:30 UTC\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderStatusHuman_empty(t *testing.T) {
	got := cli.RenderStatusHuman(nil, false)
	if got != "No active assignments.\n" {
		t.Errorf("want empty message, got: %q", got)
	}
}

func TestRenderStatusHuman_extended(t *testing.T) {
	as := []domain.ActiveAssignment{
		{ID: "s1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", ScopeID: "/subscriptions/abc", StartTime: at("2026-05-07T18:00:00Z"), EndTime: at("2026-05-07T20:00:00Z"), Status: domain.ActiveAssignmentActive},
	}
	got := cli.RenderStatusHuman(as, true)
	want := "ROLE         TYPE          NAME      STATUS  STARTED               EXPIRES               ASSIGNMENT ID  SCOPE ID\nContributor  subscription  sub-prod  active  2026-05-07 18:00 UTC  2026-05-07 20:00 UTC  s1             /subscriptions/abc\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderStatusJSON_populated(t *testing.T) {
	as := []domain.ActiveAssignment{
		{ID: "s1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeID: "/subscriptions/abc", ScopeName: "sub-prod", StartTime: at("2026-05-07T18:00:00Z"), EndTime: at("2026-05-07T20:00:00Z"), Status: domain.ActiveAssignmentActive},
	}
	got := cli.RenderStatusJSON(as)
	var decoded []map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("want 1 assignment, got %d", len(decoded))
	}
	want := map[string]any{
		"role":          "Contributor",
		"scope_type":    "subscription",
		"scope_id":      "/subscriptions/abc",
		"scope_name":    "sub-prod",
		"status":        "active",
		"started_at":    "2026-05-07T18:00:00Z",
		"expires_at":    "2026-05-07T20:00:00Z",
		"assignment_id": "s1",
	}
	for key, value := range want {
		if decoded[0][key] != value {
			t.Fatalf("%s: want %v, got %v", key, value, decoded[0][key])
		}
	}
	if len(decoded[0]) != len(want) {
		t.Fatalf("unexpected JSON fields: %+v", decoded[0])
	}
}

func TestRenderStatusJSON_empty(t *testing.T) {
	got := cli.RenderStatusJSON(nil)
	if got != "[]\n" {
		t.Errorf("want empty array, got: %q", got)
	}
}
