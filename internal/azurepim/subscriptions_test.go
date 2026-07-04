package azurepim

import (
	"context"
	"errors"
	"testing"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestSubscriptionSource_listsSubscriptions(t *testing.T) {
	source := &SubscriptionSource{subscriptions: fakeSubscriptions{subs: []subscription{
		{ID: "sub-a", Name: "Prod A"},
		{ID: "", Name: "Blank"},
		{ID: "sub-b", Name: "Prod B"},
	}}}

	got, err := source.ListSubscriptions(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 subscriptions, got %d: %+v", len(got), got)
	}
	if got[0] != (domain.Subscription{ID: "sub-a", Name: "Prod A"}) {
		t.Fatalf("unexpected first subscription: %+v", got[0])
	}
	if got[1] != (domain.Subscription{ID: "sub-b", Name: "Prod B"}) {
		t.Fatalf("unexpected second subscription: %+v", got[1])
	}
}

func TestSubscriptionSource_wrapsErrors(t *testing.T) {
	source := &SubscriptionSource{subscriptions: fakeSubscriptions{err: errors.New("network failed")}}

	_, err := source.ListSubscriptions(context.Background())
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeAzureAPIError {
		t.Fatalf("want %s, got %s", domain.CodeAzureAPIError, appErr.Code)
	}
}
