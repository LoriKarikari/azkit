package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/config"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

type fakeActivator struct {
	result *domain.ActivationResult
	err    error
}

func (a *fakeActivator) Activate(_ context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.result, nil
}

func TestActivate_missingRoleNonInteractive(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := activateRunner(t, &stdout, &stderr, nil, &fakeActivator{}, nil)

	code := runner.Run(t.Context(), []string{"activate", "--scope", "/sub/abc", "--reason", "deploy"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "role is required") {
		t.Fatalf("want missing role error, got: %s", stderr.String())
	}
}

func TestActivate_missingReasonNonInteractive(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := activateRunner(t, &stdout, &stderr, nil, &fakeActivator{}, nil)

	code := runner.Run(t.Context(), []string{"activate", "--scope", "/sub/abc", "--role", "Contributor"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "reason is required") {
		t.Fatalf("want missing reason error, got: %s", stderr.String())
	}
}

func TestActivate_nonInteractiveWithAllFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	activator := &fakeActivator{
		result: &domain.ActivationResult{
			Role:      "Contributor",
			ScopeID:   "/subscriptions/abc",
			ScopeName: "sub-prod",
			Duration:  2 * time.Hour,
			Reason:    "deploy",
		},
	}
	runner := activateRunner(t, &stdout, &stderr, nil, activator, nil)

	code := runner.Run(t.Context(), []string{
		"activate", "--scope", "/subscriptions/abc",
		"--role", "Contributor", "--reason", "deploy",
	})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Contributor") {
		t.Fatalf("want activation output, got: %s", stdout.String())
	}
}

func TestActivate_configDefaultDuration(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	activator := &fakeActivator{
		result: &domain.ActivationResult{
			Role:      "Contributor",
			ScopeID:   "/subscriptions/abc",
			ScopeName: "sub-prod",
			Duration:  30 * time.Minute,
			Reason:    "deploy",
		},
	}
	cfg := &config.Config{
		DefaultDuration: 30 * time.Minute,
	}
	runner := activateRunner(t, &stdout, &stderr, nil, activator, cfg)

	code := runner.Run(t.Context(), []string{
		"activate", "--scope", "/subscriptions/abc",
		"--role", "Contributor", "--reason", "deploy",
	})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Contributor") {
		t.Fatalf("want activation output, got: %s", stdout.String())
	}
}

func activateRunner(
	t *testing.T,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	eligibleErr error,
	activator *fakeActivator,
	cfg *config.Config,
) *cli.Runner {
	t.Helper()
	eligibleStore := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{
				ID:            "a1",
				Role:          "Contributor",
				ScopeType:     domain.ScopeSubscription,
				ScopeID:       "/subscriptions/abc",
				ScopeName:     "sub-prod",
				EligibleUntil: time.Now().Add(24 * time.Hour),
			},
		},
		Err: eligibleErr,
	}
	activeStore := &testActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{
				Role:      "Contributor",
				ScopeID:   "/subscriptions/abc",
				ScopeName: "sub-prod",
				EndTime:   time.Now().Add(2 * time.Hour),
				Status:    domain.ActiveAssignmentActive,
			},
		},
	}
	runner := cli.NewRunner(cli.Services{
		List: func(*slog.Logger) (*app.ListService, error) {
			return app.NewListService(eligibleStore), nil
		},
		Status: func(*slog.Logger) (*app.StatusService, error) {
			return app.NewStatusService(activeStore), nil
		},
		Activate: func(*slog.Logger) (*app.ActivationService, error) {
			return app.NewActivationService(eligibleStore, activator), nil
		},
	}, stdout, stderr)
	return runner
}
