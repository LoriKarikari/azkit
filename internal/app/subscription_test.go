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
	svc := app.NewSubscriptionService(cache, nil, func(active domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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
	svc := app.NewSubscriptionService(cache, nil, func(domain.TenantContext) (app.SubscriptionSource, error) {
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

func TestSubscriptionService_resolvePrefersAliasThenIDThenName(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt: now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{
				{ID: "sub-1", Name: "Production"},
				{ID: "sub-2", Name: "Development"},
			},
		},
		hasCached: true,
	}
	aliases := &fakeAliasStore{aliases: map[string]domain.Subscription{
		"prod": {ID: "sub-1", Name: "Production"},
	}}
	svc := app.NewSubscriptionService(cache, aliases, sourceFactory(nil), func() time.Time { return now })

	got, err := svc.Resolve(t.Context(), readyContext(), "prod")
	if err != nil {
		t.Fatalf("resolve alias: %v", err)
	}
	if got.ID != "sub-1" {
		t.Fatalf("want alias sub-1, got %+v", got)
	}

	got, err = svc.Resolve(t.Context(), readyContext(), "sub-2")
	if err != nil {
		t.Fatalf("resolve id: %v", err)
	}
	if got.ID != "sub-2" {
		t.Fatalf("want id sub-2, got %+v", got)
	}

	got, err = svc.Resolve(t.Context(), readyContext(), "Development")
	if err != nil {
		t.Fatalf("resolve name: %v", err)
	}
	if got.ID != "sub-2" {
		t.Fatalf("want name Development -> sub-2, got %+v", got)
	}
}

func TestSubscriptionService_resolveUnknownOrAmbiguous(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt: now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{
				{ID: "sub-1", Name: "Team A"},
				{ID: "sub-2", Name: "Team A"},
			},
		},
		hasCached: true,
	}
	svc := app.NewSubscriptionService(cache, nil, sourceFactory(nil), func() time.Time { return now })

	_, err := svc.Resolve(t.Context(), readyContext(), "missing")
	if err == nil {
		t.Fatal("want unknown subscription error")
	}
	_, err = svc.Resolve(t.Context(), readyContext(), "Team A")
	if err == nil {
		t.Fatal("want ambiguous subscription error")
	}
}

func TestSubscriptionService_createAlias(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{
		cached: domain.SubscriptionCache{
			FetchedAt:     now.Add(-time.Hour),
			Subscriptions: []domain.Subscription{{ID: "sub-1", Name: "Production"}},
		},
		hasCached: true,
	}
	aliases := &fakeAliasStore{aliases: map[string]domain.Subscription{}}
	svc := app.NewSubscriptionService(cache, aliases, sourceFactory(nil), func() time.Time { return now })

	if err := svc.CreateAlias(t.Context(), readyContext(), "prod", "sub-1"); err != nil {
		t.Fatalf("create alias: %v", err)
	}
	if aliases.aliases["prod"].ID != "sub-1" {
		t.Fatalf("alias not saved: %+v", aliases.aliases)
	}

	if err := svc.CreateAlias(t.Context(), readyContext(), "prod", "sub-1"); err == nil {
		t.Fatal("want duplicate alias error")
	}
	if err := svc.CreateAlias(t.Context(), readyContext(), "Production", "sub-1"); err == nil {
		t.Fatal("want collision with subscription name")
	}
}

func TestSubscriptionService_removeAlias(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cache := &fakeSubscriptionCache{hasCached: true, cached: domain.SubscriptionCache{FetchedAt: now.Add(-time.Hour)}}
	aliases := &fakeAliasStore{aliases: map[string]domain.Subscription{
		"prod": {ID: "sub-1", Name: "Production"},
	}}
	svc := app.NewSubscriptionService(cache, aliases, sourceFactory(nil), func() time.Time { return now })

	if err := svc.RemoveAlias(t.Context(), readyContext(), "prod"); err != nil {
		t.Fatalf("remove alias: %v", err)
	}
	if len(aliases.aliases) != 0 {
		t.Fatalf("want alias removed, got %+v", aliases.aliases)
	}
	if err := svc.RemoveAlias(t.Context(), readyContext(), "prod"); err == nil {
		t.Fatal("want alias not found")
	}
}

func sourceFactory(source *fakeSubscriptionSource) func(domain.TenantContext) (app.SubscriptionSource, error) {
	return func(domain.TenantContext) (app.SubscriptionSource, error) {
		if source == nil {
			return &fakeSubscriptionSource{}, nil
		}
		return source, nil
	}
}

type fakeAliasStore struct {
	aliases map[string]domain.Subscription
}

func (f *fakeAliasStore) Load(context.Context, domain.TenantContext) (map[string]domain.Subscription, error) {
	return mapsClone(f.aliases), nil
}

func (f *fakeAliasStore) Save(_ context.Context, _ domain.TenantContext, aliases map[string]domain.Subscription) error {
	f.aliases = mapsClone(aliases)
	return nil
}

func mapsClone(m map[string]domain.Subscription) map[string]domain.Subscription {
	if m == nil {
		return map[string]domain.Subscription{}
	}
	out := make(map[string]domain.Subscription, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
