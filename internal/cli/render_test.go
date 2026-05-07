package cli_test

import (
	"encoding/json"
	"testing"

	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestRenderHuman_populated(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", MaxDuration: "8h"},
		{ID: "a2", Role: "Reader", ScopeType: domain.ScopeResourceGroup, ScopeName: "rg-dev-app", MaxDuration: "2h"},
	}
	got := cli.RenderHuman(as, false)
	want := "ROLE         TYPE            NAME        MAX DURATION\nContributor  subscription    sub-prod    8h\nReader       resource_group  rg-dev-app  2h\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderHuman_empty(t *testing.T) {
	got := cli.RenderHuman(nil, false)
	if got != "No eligible assignments.\n" {
		t.Errorf("want empty message, got: %q", got)
	}
}

func TestRenderHuman_verbose(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", ScopeID: "/subscriptions/abc", MaxDuration: "8h"},
	}
	got := cli.RenderHuman(as, true)
	want := "ROLE         TYPE          NAME      MAX DURATION  ASSIGNMENT ID  SCOPE ID\nContributor  subscription  sub-prod  8h            a1             /subscriptions/abc\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderJSON_populated(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", ScopeID: "/subscriptions/abc", MaxDuration: "8h"},
	}
	got := cli.RenderJSON(as)
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
		"max_duration":  "8h",
		"assignment_id": "a1",
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

func TestRenderJSON_empty(t *testing.T) {
	got := cli.RenderJSON(nil)
	if got != "[]\n" {
		t.Errorf("want empty array, got: %q", got)
	}
}
