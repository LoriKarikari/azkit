package runtime_test

import (
	"errors"
	"testing"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/runtime"
)

func TestSubscriptionServiceRejectsMismatchedTenantEnv(t *testing.T) {
	t.Setenv("AZURE_TENANT_ID", "tenant-other")
	rt := runtime.New()
	svc, err := rt.Services().Subscriptions(nil)
	if err != nil {
		t.Fatalf("build subscription service: %v", err)
	}

	_, err = svc.List(t.Context(), domain.TenantContext{
		Name:     "prod",
		TenantID: "tenant-prod",
		Dir:      t.TempDir(),
		Status:   domain.ContextReady,
	}, false)
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeContextEnvMismatch {
		t.Fatalf("want %s, got %s", domain.CodeContextEnvMismatch, appErr.Code)
	}
}
