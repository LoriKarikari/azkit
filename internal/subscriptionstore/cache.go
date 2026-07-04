package subscriptionstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

const cacheFileName = "subscriptions.json"

type Cache struct{}

type cacheFile struct {
	FetchedAt     time.Time            `json:"fetched_at"`
	Subscriptions []subscriptionRecord `json:"subscriptions"`
}

type subscriptionRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var _ app.SubscriptionCache = (*Cache)(nil)

func New() *Cache {
	return &Cache{}
}

func (c *Cache) Load(
	ctx context.Context,
	active domain.TenantContext,
) (domain.SubscriptionCache, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.SubscriptionCache{}, false, err
	}
	path := cachePath(active)
	contents, err := os.ReadFile(path) // #nosec G304 -- fixed cache file under the selected context credential cache dir.
	if errors.Is(err, os.ErrNotExist) {
		return domain.SubscriptionCache{}, false, nil
	}
	if err != nil {
		return domain.SubscriptionCache{}, false, fmt.Errorf("reading subscription cache: %w", err)
	}
	var data cacheFile
	if err := json.Unmarshal(contents, &data); err != nil {
		return domain.SubscriptionCache{}, false, fmt.Errorf("parsing subscription cache: %w", err)
	}
	return domain.SubscriptionCache{
		FetchedAt:     data.FetchedAt,
		Subscriptions: toDomain(data.Subscriptions),
	}, true, nil
}

func (c *Cache) Save(
	ctx context.Context,
	active domain.TenantContext,
	cache domain.SubscriptionCache,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(active.Dir, 0700); err != nil {
		return fmt.Errorf("creating subscription cache dir: %w", err)
	}
	data := cacheFile{
		FetchedAt:     cache.FetchedAt.UTC(),
		Subscriptions: fromDomain(cache.Subscriptions),
	}
	contents, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding subscription cache: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(cachePath(active), contents, 0600); err != nil {
		return fmt.Errorf("writing subscription cache: %w", err)
	}
	return nil
}

func (c *Cache) Invalidate(ctx context.Context, active domain.TenantContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Remove(cachePath(active)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing subscription cache: %w", err)
	}
	return nil
}

func cachePath(active domain.TenantContext) string {
	return filepath.Join(active.Dir, cacheFileName)
}

func toDomain(records []subscriptionRecord) []domain.Subscription {
	subscriptions := make([]domain.Subscription, 0, len(records))
	for _, record := range records {
		subscriptions = append(subscriptions, domain.Subscription{ID: record.ID, Name: record.Name})
	}
	return subscriptions
}

func fromDomain(subscriptions []domain.Subscription) []subscriptionRecord {
	records := make([]subscriptionRecord, 0, len(subscriptions))
	for _, sub := range subscriptions {
		records = append(records, subscriptionRecord{ID: sub.ID, Name: sub.Name})
	}
	return records
}
