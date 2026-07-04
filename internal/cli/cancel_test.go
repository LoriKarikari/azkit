package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/config"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/inmemory"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

func TestCancelActivationInteractive(t *testing.T) {
	originalTerm := interactive.IsTerminalFn
	t.Cleanup(func() { interactive.IsTerminalFn = originalTerm })
	interactive.IsTerminalFn = func() bool { return true }

	var stdout, stderr bytes.Buffer
	eligible := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{Role: "Contributor", RoleDefID: "/roleDefs/111", ScopeID: "/sub/abc", ScopeName: "sub-prod"},
		},
	}
	runner := cli.NewRunner(cli.Services{
		List: func(_ *slog.Logger) (*app.ListService, error) {
			return app.NewListService(eligible), nil
		},
		Activate: func(_ *slog.Logger) (*app.ActivationService, error) {
			return app.NewActivationService(nil, nil, nil), nil
		},
		Status: func(_ *slog.Logger) (*app.StatusService, error) {
			return app.NewStatusService(&inmemory.ActiveAssignments{}), nil
		},
		Deactivate: func(_ *slog.Logger) (*app.DeactivationService, error) {
			return nil, nil
		},
		ActivateInteractive: func(_ context.Context, _ []domain.EligibleAssignment, _ *app.ActivationService, _ *config.Config, _ interactive.ActivationInput) (*domain.ActivationResult, error) {
			return nil, interactive.ErrCanceled
		},
	}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"pim", "activate"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Activation canceled") {
		t.Fatalf("want cancel message on stderr, got: %q", stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
}

func TestCancelDeactivationInteractive(t *testing.T) {
	originalTerm := interactive.IsTerminalFn
	t.Cleanup(func() { interactive.IsTerminalFn = originalTerm })
	interactive.IsTerminalFn = func() bool { return true }

	var stdout, stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{
		List: func(_ *slog.Logger) (*app.ListService, error) {
			return app.NewListService(&inmemory.EligibleAssignments{}), nil
		},
		Activate: func(_ *slog.Logger) (*app.ActivationService, error) {
			return nil, nil
		},
		Status: func(_ *slog.Logger) (*app.StatusService, error) {
			return app.NewStatusService(&inmemory.ActiveAssignments{}), nil
		},
		Deactivate: func(_ *slog.Logger) (*app.DeactivationService, error) {
			return app.NewDeactivationService(&inmemory.ActiveAssignments{}, nil), nil
		},
		DeactivateInteractive: func(_ context.Context, _ []domain.ActiveAssignment, _ *app.DeactivationService, _ interactive.DeactivationInput) (*domain.DeactivationResult, error) {
			return nil, interactive.ErrCanceled
		},
	}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"pim", "deactivate"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Deactivation canceled") {
		t.Fatalf("want cancel message on stderr, got: %q", stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
}
