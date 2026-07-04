package cli

import (
	"testing"

	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestActivateFilterEligibleByRole(t *testing.T) {
	cmd := &ActivateCmd{Role: "Contributor"}
	eligible := []domain.EligibleAssignment{
		{Role: "Contributor", ScopeID: "/subscriptions/abc"},
		{Role: "Reader", ScopeID: "/subscriptions/abc"},
	}

	got := cmd.filterEligible(eligible)
	if len(got) != 1 || got[0].Role != "Contributor" {
		t.Fatalf("want only Contributor, got %+v", got)
	}
}

func TestActivateFilterEligibleBySubscription(t *testing.T) {
	cmd := &ActivateCmd{Subscription: "Production"}
	eligible := []domain.EligibleAssignment{
		{Role: "Contributor", ScopeID: "/subscriptions/abc", SubscriptionName: "Production"},
		{Role: "Contributor", ScopeID: "/subscriptions/def", SubscriptionName: "Sandbox"},
	}

	got := cmd.filterEligible(eligible)
	if len(got) != 1 || got[0].SubscriptionName != "Production" {
		t.Fatalf("want only Production subscription, got %+v", got)
	}
}

func TestActivateFilterEligibleByScope(t *testing.T) {
	cmd := &ActivateCmd{Scope: "/subscriptions/abc/resourceGroups/prod-rg"}
	eligible := []domain.EligibleAssignment{
		{Role: "Contributor", ScopeID: "/subscriptions/abc/resourceGroups/prod-rg"},
		{Role: "Contributor", ScopeID: "/subscriptions/abc/resourceGroups/dev-rg"},
	}

	got := cmd.filterEligible(eligible)
	if len(got) != 1 || got[0].ScopeID != cmd.Scope {
		t.Fatalf("want only matching scope, got %+v", got)
	}
}

func TestActivateFilterEligibleByResourceGroup(t *testing.T) {
	cmd := &ActivateCmd{ResourceGroup: "prod-rg"}
	eligible := []domain.EligibleAssignment{
		{Role: "Contributor", ScopeType: domain.ScopeResourceGroup, ScopeName: "prod-rg"},
		{Role: "Contributor", ScopeType: domain.ScopeResourceGroup, ScopeName: "dev-rg"},
		{Role: "Contributor", ScopeType: domain.ScopeSubscription, ScopeName: "prod-rg"},
	}

	got := cmd.filterEligible(eligible)
	if len(got) != 1 || got[0].ScopeName != "prod-rg" || got[0].ScopeType != domain.ScopeResourceGroup {
		t.Fatalf("want only matching resource group, got %+v", got)
	}
}
