package subscriptionstore_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

func TestCache_SaveLoadAndInvalidate(t *testing.T) {
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

	if err := cache.Invalidate(t.Context(), active); err != nil {
		t.Fatalf("invalidate cache: %v", err)
	}
	_, ok, err = cache.Load(t.Context(), active)
	if err != nil {
		t.Fatalf("load after invalidate: %v", err)
	}
	if ok {
		t.Fatal("want cache miss after invalidate")
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

func TestCache_InvalidateMissingFileSucceeds(t *testing.T) {
	cache := subscriptionstore.New()
	active := domain.TenantContext{Name: "prod", TenantID: "tenant", Dir: t.TempDir()}

	if err := cache.Invalidate(t.Context(), active); err != nil {
		t.Fatalf("invalidate missing cache: %v", err)
	}
}
