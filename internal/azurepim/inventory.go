package azurepim

import (
	"context"
	"log/slog"

	"github.com/LoriKarikari/azkit/internal/app"
)

func listAcrossSubscriptions[T any](
	ctx context.Context,
	subscriptions subscriptionSource,
	log *slog.Logger,
	operation string,
	list func(context.Context, subscription) ([]T, error),
	enrich func(*T, subscription),
	key func(T) string,
) ([]T, error) {
	subs, err := subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	log.Debug("listed subscriptions", slog.Int("count", len(subs)))

	all := []T{}
	seen := map[string]struct{}{}
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		log.Debug(operation, slog.String("subscription_id", sub.ID))
		items, err := list(ctx, sub)
		if err != nil {
			log.Debug(
				operation+" failed",
				slog.String("subscription_id", sub.ID),
				slog.Any("error", err),
			)
			return nil, err
		}
		log.Debug(
			operation+" completed",
			slog.String("subscription_id", sub.ID),
			slog.Int("count", len(items)),
		)
		for i := range items {
			enrich(&items[i], sub)
			itemKey := key(items[i])
			if itemKey != "" {
				if _, ok := seen[itemKey]; ok {
					continue
				}
				seen[itemKey] = struct{}{}
			}
			all = append(all, items[i])
		}
	}
	return all, nil
}
