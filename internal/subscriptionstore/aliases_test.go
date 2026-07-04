package subscriptionstore_test

import (
	"testing"

	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

func TestAliases_LoadMissingReturnsEmptyMap(t *testing.T) {
	store := subscriptionstore.NewAliases()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}

	aliases, err := store.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load missing aliases: %v", err)
	}
	if len(aliases) != 0 {
		t.Fatalf("want empty aliases, got %+v", aliases)
	}
}

func TestAliases_SaveAndLoadRoundTrip(t *testing.T) {
	store := subscriptionstore.NewAliases()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}
	want := map[string]domain.Subscription{
		"prod": {ID: "sub-1", Name: "Production"},
		"dev":  {ID: "sub-2", Name: "Development"},
	}

	if err := store.Save(t.Context(), active, want); err != nil {
		t.Fatalf("save aliases: %v", err)
	}
	got, err := store.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load aliases: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 aliases, got %+v", got)
	}
	if got["prod"].ID != "sub-1" || got["dev"].ID != "sub-2" {
		t.Fatalf("unexpected aliases: %+v", got)
	}
}

func TestAliases_SaveOverwrites(t *testing.T) {
	store := subscriptionstore.NewAliases()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}

	if err := store.Save(t.Context(), active, map[string]domain.Subscription{
		"prod": {ID: "sub-old", Name: "Old"},
	}); err != nil {
		t.Fatalf("save old: %v", err)
	}
	if err := store.Save(t.Context(), active, map[string]domain.Subscription{
		"prod": {ID: "sub-new", Name: "New"},
	}); err != nil {
		t.Fatalf("save new: %v", err)
	}
	got, err := store.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got["prod"].ID != "sub-new" {
		t.Fatalf("want overwritten alias, got %+v", got)
	}
}
