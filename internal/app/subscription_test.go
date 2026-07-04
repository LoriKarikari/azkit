package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestSubscriptionService_usesFreshCache(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{{ID: "sub-cached", Name: "Cached"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{}
	svc := app.NewSubscriptionService(cache, func(active domain.TenantContext) (app.SubscriptionSource, error) {
		if active.Name != "prod" {
			t.Fatalf("unexpected active context: %+v", active)
		}
		return source, nil
	}, func() time.Time { return now })

	got, err := svc.List(t.Context(), readyContext(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "sub-cached" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
	if source.calls != 0 {
		t.Fatalf("source should not be called, got %d calls", source.calls)
	}
}

func TestSubscriptionService_refreshOverwritesCache(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{{ID: "sub-cached", Name: "Cached"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-live", Name: "Live"}}}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	got, err := svc.List(t.Context(), readyContext(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "sub-live" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
	if source.calls != 1 {
		t.Fatalf("want one source call, got %d", source.calls)
	}
	if !cache.saved.FetchedAt.Equal(now) || cache.saved.Subscriptions[0].ID != "sub-live" {
		t.Fatalf("unexpected saved cache: %+v", cache.saved)
	}
}

func TestSubscriptionService_refreshKeepsCacheWhenFetchFails(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{{ID: "sub-cached", Name: "Cached"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{err: errors.New("network down")}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	_, err := svc.List(t.Context(), readyContext(), true)
	if err == nil {
		t.Fatal("want refresh fetch error")
	}
	if cache.saved.Subscriptions != nil {
		t.Fatal("cache should not be saved when refresh fetch fails")
	}
}

func TestSubscriptionService_fetchesWhenCacheExpired(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-(25 * time.Hour)),
			Subscriptions: []domain.Subscription{{ID: "sub-stale", Name: "Stale"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-live", Name: "Live"}}}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	got, err := svc.List(t.Context(), readyContext(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "sub-live" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
	if source.calls != 1 {
		t.Fatalf("want one source call, got %d", source.calls)
	}
}

func TestSubscriptionService_treatsExactlyTTLAsStale(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-app.DefaultSubscriptionCacheTTL),
			Subscriptions: []domain.Subscription{{ID: "sub-stale", Name: "Stale"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-live", Name: "Live"}}}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	got, err := svc.List(t.Context(), readyContext(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].ID != "sub-live" {
		t.Fatalf("want live subscription at TTL boundary, got %+v", got)
	}
}

func TestSubscriptionService_servesFreshCacheEvenWhenContextNeedsLogin(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{{ID: "sub-cached", Name: "Cached"}},
		},
		hasCached: true,
	}
	source := &fakeSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-live", Name: "Live"}}}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	got, err := svc.List(
		t.Context(),
		domain.TenantContext{Name: "prod", TenantID: "tenant", Status: domain.ContextNeedsLogin},
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].ID != "sub-cached" {
		t.Fatalf("want cached subscription, got %+v", got)
	}
	if source.calls != 0 {
		t.Fatalf("source should not be called, got %d calls", source.calls)
	}
}

func TestSubscriptionService_needsLoginWithoutFreshCache(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{}
	source := &fakeSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-live", Name: "Live"}}}
	svc := app.NewSubscriptionService(cache, func(domain.TenantContext) (app.SubscriptionSource, error) {
		return source, nil
	}, func() time.Time { return now })

	_, err := svc.List(t.Context(), domain.TenantContext{Name: "prod", TenantID: "tenant", Status: domain.ContextNeedsLogin}, false)
	if err == nil {
		t.Fatal("want context needs login error")
	}
	if source.calls != 0 {
		t.Fatalf("source should not be called, got %d calls", source.calls)
	}
}

func readyContext() domain.TenantContext {
	return domain.TenantContext{Name: "prod", TenantID: "tenant", Status: domain.ContextReady}
}

type fakeSubscriptionCache struct {
	cached    domain.SubscriptionCache
	saved     domain.SubscriptionCache
	hasCached bool
}

func (f *fakeSubscriptionCache) Load(context.Context, domain.TenantContext) (domain.SubscriptionCache, bool, error) {
	return f.cached, f.hasCached, nil
}

func (f *fakeSubscriptionCache) Save(_ context.Context, _ domain.TenantContext, cache domain.SubscriptionCache) error {
	f.saved = cache
	return nil
}

type fakeSubscriptionSource struct {
	subscriptions []domain.Subscription
	calls         int
	err           error
}

func (f *fakeSubscriptionSource) ListSubscriptions(context.Context) ([]domain.Subscription, error) {
	f.calls++
	return f.subscriptions, f.err
}
