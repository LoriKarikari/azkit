package app

import (
	"context"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
)

const DefaultSubscriptionCacheTTL = 24 * time.Hour

type SubscriptionSource interface {
	ListSubscriptions(context.Context) ([]domain.Subscription, error)
}

type SubscriptionSourceFactory func(domain.TenantContext) (SubscriptionSource, error)

type SubscriptionCache interface {
	Load(context.Context, domain.TenantContext) (domain.SubscriptionCache, bool, error)
	Save(context.Context, domain.TenantContext, domain.SubscriptionCache) error
	Invalidate(context.Context, domain.TenantContext) error
}

type SubscriptionService struct {
	cache  SubscriptionCache
	source SubscriptionSourceFactory
	now    func() time.Time
	ttl    time.Duration
}

func NewSubscriptionService(
	cache SubscriptionCache,
	source SubscriptionSourceFactory,
	now func() time.Time,
) *SubscriptionService {
	if now == nil {
		now = time.Now
	}
	return &SubscriptionService{
		cache:  cache,
		source: source,
		now:    now,
		ttl:    DefaultSubscriptionCacheTTL,
	}
}

func (s *SubscriptionService) List(
	ctx context.Context,
	active domain.TenantContext,
	refresh bool,
) ([]domain.Subscription, error) {
	if !refresh {
		cached, ok, err := s.cache.Load(ctx, active)
		if err != nil {
			return nil, err
		}
		if ok && s.cacheFresh(cached) {
			return cached.Subscriptions, nil
		}
	}
	if active.Status != domain.ContextReady {
		return nil, ContextNeedsLogin(active)
	}
	source, err := s.source(active)
	if err != nil {
		return nil, err
	}
	subscriptions, err := source.ListSubscriptions(ctx)
	if err != nil {
		return nil, err
	}
	if refresh {
		if err := s.cache.Invalidate(ctx, active); err != nil {
			return nil, err
		}
	}
	cached := domain.SubscriptionCache{
		FetchedAt:     s.now().UTC(),
		Subscriptions: subscriptions,
	}
	if err := s.cache.Save(ctx, active, cached); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *SubscriptionService) cacheFresh(cached domain.SubscriptionCache) bool {
	if cached.FetchedAt.IsZero() {
		return false
	}
	return s.now().UTC().Sub(cached.FetchedAt.UTC()) < s.ttl
}
