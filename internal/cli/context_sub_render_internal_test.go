package cli

import (
	"testing"

	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRenderCurrentContextHumanGolden(t *testing.T) {
	got := renderCurrentContextHuman(domain.TenantContext{
		Name:     "prod",
		TenantID: "tenant-prod",
		Dir:      "/state/contexts/prod",
		Status:   domain.ContextReady,
	}, true)
	want := "Context:     prod\nTenant:      tenant-prod\nStatus:      ready\nConfig dir:  /state/contexts/prod\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderContextsHumanGolden(t *testing.T) {
	got := renderContextsHuman([]domain.TenantContext{
		{Name: "dev", TenantID: "tenant-dev", Status: domain.ContextNeedsLogin},
		{Name: "prod", TenantID: "tenant-prod", Status: domain.ContextReady},
	}, "prod")
	want := "CURRENT  NAME  TENANT       STATUS\n         dev   tenant-dev   needs_login\n*        prod  tenant-prod  ready\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderCurrentSubscriptionHumanGolden(t *testing.T) {
	got := renderCurrentSubscriptionHuman("sub-prod", "Production")
	want := "Subscription:  sub-prod\nName:          Production\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderSubscriptionsHumanGolden(t *testing.T) {
	got := renderSubscriptionsHuman([]domain.Subscription{
		{ID: "sub-dev", Name: "Development"},
		{ID: "sub-prod", Name: "Production"},
	})
	want := "ID        NAME\nsub-dev   Development\nsub-prod  Production\n"
	if got != want {
		t.Fatalf("want:\n%s\ngot:\n%s", want, got)
	}
}
