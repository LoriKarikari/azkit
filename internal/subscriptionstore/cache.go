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
		return domain.SubscriptionCache{}, false, nil
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
	if err := writeFileAtomic(cachePath(active), contents); err != nil {
		return fmt.Errorf("writing subscription cache: %w", err)
	}
	return nil
}

func cachePath(active domain.TenantContext) string {
	return filepath.Join(active.Dir, cacheFileName)
}

func writeFileAtomic(target string, contents []byte) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, filepath.Base(target)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(contents); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, target)
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
