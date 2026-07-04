package cli_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRenderHuman_populated(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", EligibleUntil: at("2026-05-07T20:00:00Z")},
		{ID: "a2", Role: "Reader", ScopeType: domain.ScopeResourceGroup, ScopeName: "rg-dev-app", EligibleUntil: at("2026-05-08T09:30:00Z")},
	}
	got := cli.RenderHuman(as, false)
	want := "ROLE         TYPE            NAME        ELIGIBLE UNTIL\nContributor  subscription    sub-prod    2026-05-07 20:00 UTC\nReader       resource_group  rg-dev-app  2026-05-08 09:30 UTC\n"
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

func TestRenderHuman_extended(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", ScopeID: "/subscriptions/abc", EligibleUntil: at("2026-05-07T20:00:00Z")},
	}
	got := cli.RenderHuman(as, true)
	want := "ROLE         TYPE          NAME      ELIGIBLE UNTIL        ASSIGNMENT ID  SCOPE ID\nContributor  subscription  sub-prod  2026-05-07 20:00 UTC  a1             /subscriptions/abc\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderJSON_populated(t *testing.T) {
	as := []domain.EligibleAssignment{
		{ID: "a1", Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "sub-prod", ScopeID: "/subscriptions/abc", SubscriptionID: "abc", SubscriptionName: "sub-prod", EligibleUntil: at("2026-05-07T20:00:00Z")},
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
		"role":              "Contributor",
		"scope_type":        "subscription",
		"scope_id":          "/subscriptions/abc",
		"scope_name":        "sub-prod",
		"subscription_id":   "abc",
		"subscription_name": "sub-prod",
		"eligible_until":    "2026-05-07T20:00:00Z",
		"assignment_id":     "a1",
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

func at(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
