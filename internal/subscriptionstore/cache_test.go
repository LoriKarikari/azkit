package subscriptionstore_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

func TestCache_SaveAndLoadRoundTrip(t *testing.T) {
	cache := subscriptionstore.New()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}
	fetchedAt := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	want := domain.SubscriptionCache{
		FetchedAt:     fetchedAt,
		Subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}},
	}

	if err := cache.Save(t.Context(), active, want); err != nil {
		t.Fatalf("save cache: %v", err)
	}
	got, ok, err := cache.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if !ok {
		t.Fatal("want cache hit")
	}
	if !got.FetchedAt.Equal(fetchedAt) || len(got.Subscriptions) != 1 || got.Subscriptions[0].ID != "sub-a" {
		t.Fatalf("unexpected cache: %+v", got)
	}
	info, err := os.Stat(filepath.Join(active.Dir, "subscriptions.json"))
	if err != nil {
		t.Fatalf("stat cache: %v", err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0600 {
		t.Fatalf("want 0600 cache file, got %v", gotMode)
	}
}

func TestCache_SaveOverwritesPreviousCache(t *testing.T) {
	cache := subscriptionstore.New()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}

	if err := cache.Save(t.Context(), active, domain.SubscriptionCache{
		FetchedAt:     time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
		Subscriptions: []domain.Subscription{{ID: "sub-old", Name: "Old"}},
	}); err != nil {
		t.Fatalf("save old cache: %v", err)
	}
	if err := cache.Save(t.Context(), active, domain.SubscriptionCache{
		FetchedAt:     time.Date(2026, 7, 4, 13, 0, 0, 0, time.UTC),
		Subscriptions: []domain.Subscription{{ID: "sub-new", Name: "New"}},
	}); err != nil {
		t.Fatalf("save new cache: %v", err)
	}
	got, ok, err := cache.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if !ok || len(got.Subscriptions) != 1 || got.Subscriptions[0].ID != "sub-new" {
		t.Fatalf("want overwritten cache, got %+v", got)
	}
}

func TestCache_MissingAndCorruptFilesAreMisses(t *testing.T) {
	cache := subscriptionstore.New()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}

	_, ok, err := cache.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("missing cache should not error: %v", err)
	}
	if ok {
		t.Fatal("missing cache should be a miss")
	}

	if err := os.WriteFile(filepath.Join(active.Dir, "subscriptions.json"), []byte("{"), 0600); err != nil {
		t.Fatalf("write corrupt cache: %v", err)
	}
	_, ok, err = cache.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("corrupt cache should not error: %v", err)
	}
	if ok {
		t.Fatal("corrupt cache should be a miss")
	}
}
