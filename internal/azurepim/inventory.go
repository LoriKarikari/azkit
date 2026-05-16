package azurepim

import (
	"context"
	"log/slog"

	"github.com/LoriKarikari/pimctl/internal/app"
)

func listAcrossSubscriptions[T any](
	ctx context.Context,
	subscriptions subscriptionSource,
	log *slog.Logger,
	operation string,
	list func(context.Context, subscription) ([]T, error),
	enrich func(*T, subscription),
) ([]T, error) {
	subs, err := subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, app.AuthFailed(err)
	}

	log.Debug("listed subscriptions", slog.Int("count", len(subs)))

	all := []T{}
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
		}
		all = append(all, items...)
	}
	return all, nil
}
