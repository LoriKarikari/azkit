package azurepim

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

type SubscriptionSource struct {
	subscriptions subscriptionSource
}

var _ app.SubscriptionSource = (*SubscriptionSource)(nil)

func NewSubscriptionSourceFromCred(cred azcore.TokenCredential) *SubscriptionSource {
	return &SubscriptionSource{subscriptions: azureSubscriptions{cred: cred}}
}

func (s *SubscriptionSource) ListSubscriptions(ctx context.Context) ([]domain.Subscription, error) {
	subs, err := s.subscriptions.ListSubscriptions(ctx)
	if err != nil {
		return nil, azurePIMOperationError(err)
	}
	out := make([]domain.Subscription, 0, len(subs))
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		out = append(out, domain.Subscription{ID: sub.ID, Name: sub.Name})
	}
	return out, nil
}
