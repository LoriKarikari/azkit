package cli_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

func TestRunner_subListFetchesAndCachesForActiveContext(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	source := &cliSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := subscriptionRunner(&stdout, &stderr, source, now)
	addReadyContext(t, runner, &stdout, &stderr, stateRoot, "prod", "tenant-prod")

	code := runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sub-a") || !strings.Contains(stdout.String(), "Prod A") {
		t.Fatalf("missing subscription output:\n%s", stdout.String())
	}
	if source.calls != 1 {
		t.Fatalf("want one source call, got %d", source.calls)
	}

	stdout.Reset()
	stderr.Reset()
	source.subscriptions = []domain.Subscription{{ID: "sub-b", Name: "Prod B"}}
	code = runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sub-a") || strings.Contains(stdout.String(), "sub-b") {
		t.Fatalf("want cached subscription output, got:\n%s", stdout.String())
	}
	if source.calls != 1 {
		t.Fatalf("cache should avoid source call, got %d calls", source.calls)
	}
}

func TestRunner_subRefreshInvalidatesCache(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	source := &cliSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := subscriptionRunner(&stdout, &stderr, source, now)
	addReadyContext(t, runner, &stdout, &stderr, stateRoot, "prod", "tenant-prod")

	if code := runner.Run(t.Context(), []string{"sub", "-l"}); code != 0 {
		t.Fatalf("initial list: exit %d: %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	source.subscriptions = []domain.Subscription{{ID: "sub-b", Name: "Prod B"}}
	code := runner.Run(t.Context(), []string{"sub", "--refresh"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sub-b") || strings.Contains(stdout.String(), "sub-a") {
		t.Fatalf("want refreshed subscription output, got:\n%s", stdout.String())
	}
	if source.calls != 2 {
		t.Fatalf("want second source call after refresh, got %d", source.calls)
	}
}

func TestRunner_subListMissingActiveContext(t *testing.T) {
	setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := subscriptionRunner(&stdout, &stderr, &cliSubscriptionSource{}, time.Now())

	code := runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "No active context") {
		t.Fatalf("want missing active context error, got: %s", stderr.String())
	}
}

func TestRunner_subListContextNeedsLogin(t *testing.T) {
	setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	source := &cliSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}}}
	runner := subscriptionRunner(&stdout, &stderr, source, time.Now())
	if code := runner.Run(t.Context(), []string{"ctx", "add", "prod", "--tenant", "tenant-prod"}); code != 0 {
		t.Fatalf("add context: exit %d: %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	code := runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "az login --tenant tenant-prod") {
		t.Fatalf("want login guidance, got: %s", stderr.String())
	}
	if source.calls != 0 {
		t.Fatalf("source should not be called when context needs login, got %d", source.calls)
	}
}

func TestRunner_subBareCommandFailsWithoutStdout(t *testing.T) {
	setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	source := &cliSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}}}
	runner := subscriptionRunner(&stdout, &stderr, source, time.Now())

	code := runner.Run(t.Context(), []string{"--shell-env", "sub"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("bare shell-env sub must not print eval-able stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "azkit sub -l") {
		t.Fatalf("want list guidance, got: %s", stderr.String())
	}
}

func TestRunner_subListCredentialFailureIsActionable(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	source := &cliSubscriptionSource{factoryErr: app.AuthFailed(errors.New("missing credential"))}
	runner := subscriptionRunner(&stdout, &stderr, source, time.Now())
	addReadyContext(t, runner, &stdout, &stderr, stateRoot, "prod", "tenant-prod")

	code := runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Could not authenticate with Azure") {
		t.Fatalf("want authentication guidance, got: %s", stderr.String())
	}
}

func TestRunner_subCacheIsPerContext(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	source := &cliSubscriptionSource{subscriptions: []domain.Subscription{{ID: "sub-a", Name: "Prod A"}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := subscriptionRunner(&stdout, &stderr, source, now)

	t.Setenv("AZKIT_CONTEXT", "prod")
	addReadyContext(t, runner, &stdout, &stderr, stateRoot, "prod", "tenant-prod")
	code := runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 0 {
		t.Fatalf("prod list: exit %d: %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	source.subscriptions = []domain.Subscription{{ID: "sub-b", Name: "Dev B"}}
	t.Setenv("AZKIT_CONTEXT", "dev")
	addReadyContext(t, runner, &stdout, &stderr, stateRoot, "dev", "tenant-dev")
	code = runner.Run(t.Context(), []string{"sub", "-l"})
	if code != 0 {
		t.Fatalf("dev list: exit %d: %s", code, stderr.String())
	}
	if source.calls != 2 {
		t.Fatalf("want separate fetch per context, got %d source calls", source.calls)
	}
	if !strings.Contains(stdout.String(), "sub-b") || strings.Contains(stdout.String(), "sub-a") {
		t.Fatalf("want dev subscription output, got:\n%s", stdout.String())
	}
}

func subscriptionRunner(
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	source *cliSubscriptionSource,
	now time.Time,
) *cli.Runner {
	return cli.NewRunner(cli.Services{
		Subscriptions: func(*slog.Logger) (*app.SubscriptionService, error) {
			return app.NewSubscriptionService(
				subscriptionstore.New(),
				func(domain.TenantContext) (app.SubscriptionSource, error) {
					if source.factoryErr != nil {
						return nil, source.factoryErr
					}
					return source, nil
				},
				func() time.Time { return now },
			), nil
		},
	}, stdout, stderr)
}

func addReadyContext(
	t *testing.T,
	runner *cli.Runner,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	stateRoot string,
	name string,
	tenantID string,
) {
	t.Helper()
	if code := runner.Run(t.Context(), []string{"ctx", "add", name, "--tenant", tenantID}); code != 0 {
		t.Fatalf("add context: exit %d: %s", code, stderr.String())
	}
	contextDir := filepath.Join(stateRoot, "azkit", "contexts", name)
	if err := os.WriteFile(filepath.Join(contextDir, "azureProfile.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
}

type cliSubscriptionSource struct {
	subscriptions []domain.Subscription
	calls         int
	factoryErr    error
}

func (s *cliSubscriptionSource) ListSubscriptions(context.Context) ([]domain.Subscription, error) {
	s.calls++
	return s.subscriptions, nil
}
