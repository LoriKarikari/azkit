package app

import (
	"context"
	"regexp"
	"strings"
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
}

type AliasStore interface {
	Load(context.Context, domain.TenantContext) (map[string]domain.Subscription, error)
	Save(context.Context, domain.TenantContext, map[string]domain.Subscription) error
}

type SubscriptionService struct {
	cache      SubscriptionCache
	aliasStore AliasStore
	source     SubscriptionSourceFactory
	now        func() time.Time
	ttl        time.Duration
}

func NewSubscriptionService(
	cache SubscriptionCache,
	aliasStore AliasStore,
	source SubscriptionSourceFactory,
	now func() time.Time,
) *SubscriptionService {
	if now == nil {
		now = time.Now
	}
	return &SubscriptionService{
		cache:      cache,
		aliasStore: aliasStore,
		source:     source,
		now:        now,
		ttl:        DefaultSubscriptionCacheTTL,
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
	cached := domain.SubscriptionCache{
		FetchedAt:     s.now().UTC(),
		Subscriptions: subscriptions,
	}
	if err := s.cache.Save(ctx, active, cached); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *SubscriptionService) Resolve(
	ctx context.Context,
	active domain.TenantContext,
	selector string,
) (domain.Subscription, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return domain.Subscription{}, MissingSubscriptionCommand()
	}
	aliases, err := s.aliasStore.Load(ctx, active)
	if err != nil {
		return domain.Subscription{}, err
	}
	if sub, ok := aliases[aliasKey(selector)]; ok {
		return sub, nil
	}
	subs, err := s.List(ctx, active, false)
	if err != nil {
		return domain.Subscription{}, err
	}
	var byID []domain.Subscription
	var byName []domain.Subscription
	for _, sub := range subs {
		if strings.EqualFold(sub.ID, selector) {
			byID = append(byID, sub)
		}
		if strings.EqualFold(sub.Name, selector) {
			byName = append(byName, sub)
		}
	}
	if len(byID) == 1 {
		return byID[0], nil
	}
	if len(byName) == 1 {
		return byName[0], nil
	}
	if len(byName) > 1 || len(byID) > 1 {
		return domain.Subscription{}, AmbiguousSubscription(selector)
	}
	return domain.Subscription{}, UnknownSubscription(selector)
}

func (s *SubscriptionService) CreateAlias(
	ctx context.Context,
	active domain.TenantContext,
	alias string,
	selector string,
) error {
	alias = strings.TrimSpace(alias)
	if err := validateAliasName(alias); err != nil {
		return err
	}
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return MissingAliasSelector()
	}
	sub, err := s.Resolve(ctx, active, selector)
	if err != nil {
		return err
	}
	subs, err := s.List(ctx, active, false)
	if err != nil {
		return err
	}
	for _, existing := range subs {
		if strings.EqualFold(existing.Name, alias) || strings.EqualFold(existing.ID, alias) {
			return AliasNameCollision(alias, existing)
		}
	}
	aliases, err := s.aliasStore.Load(ctx, active)
	if err != nil {
		return err
	}
	key := aliasKey(alias)
	if _, exists := aliases[key]; exists {
		return AliasAlreadyExists(alias)
	}
	aliases[key] = sub
	return s.aliasStore.Save(ctx, active, aliases)
}

func (s *SubscriptionService) RemoveAlias(
	ctx context.Context,
	active domain.TenantContext,
	alias string,
) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return MissingAliasName()
	}
	aliases, err := s.aliasStore.Load(ctx, active)
	if err != nil {
		return err
	}
	key := aliasKey(alias)
	if _, ok := aliases[key]; !ok {
		return AliasNotFound(alias)
	}
	delete(aliases, key)
	return s.aliasStore.Save(ctx, active, aliases)
}

func (s *SubscriptionService) cacheFresh(cached domain.SubscriptionCache) bool {
	if cached.FetchedAt.IsZero() {
		return false
	}
	return s.now().UTC().Sub(cached.FetchedAt.UTC()) < s.ttl
}

var aliasNameRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,62}$`)

var reservedAliasNames = map[string]struct{}{
	"-":       {},
	"alias":   {},
	"current": {},
	"help":    {},
	"list":    {},
	"refresh": {},
	"unalias": {},
}

func validateAliasName(alias string) error {
	if !aliasNameRE.MatchString(alias) {
		return InvalidAliasName(alias)
	}
	if _, ok := reservedAliasNames[aliasKey(alias)]; ok {
		return InvalidAliasName(alias)
	}
	return nil
}

func aliasKey(alias string) string {
	return strings.ToLower(strings.TrimSpace(alias))
}
