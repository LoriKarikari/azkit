package interactive_test

import (
	"errors"
	"testing"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

func TestPickContextEmpty(t *testing.T) {
	_, err := interactive.PickContext(t.Context(), nil)
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeContextNotFound {
		t.Fatalf("want context not found, got %s", appErr.Code)
	}
}

func TestPickSubscriptionEmpty(t *testing.T) {
	_, err := interactive.PickSubscription(t.Context(), nil)
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeNoSubscriptions {
		t.Fatalf("want no subscriptions, got %s", appErr.Code)
	}
}
