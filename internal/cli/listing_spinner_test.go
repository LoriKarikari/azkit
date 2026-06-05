package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/config"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/inmemory"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

type delayEligibleStore struct {
	inner app.EligibleAssignments
	delay time.Duration
}

func (s *delayEligibleStore) ListEligible(ctx context.Context) ([]domain.EligibleAssignment, error) {
	select {
	case <-time.After(s.delay):
		return s.inner.ListEligible(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type delayActiveStore struct {
	inner app.ActiveAssignments
	delay time.Duration
}

func (s *delayActiveStore) ListActive(ctx context.Context) ([]domain.ActiveAssignment, error) {
	select {
	case <-time.After(s.delay):
		return s.inner.ListActive(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestActivateShowsListingSpinner(t *testing.T) {
	originalTerm := interactive.IsTerminalFn
	t.Cleanup(func() { interactive.IsTerminalFn = originalTerm })
	interactive.IsTerminalFn = func() bool { return true }

	var stdout, stderr bytes.Buffer

	store := &delayEligibleStore{
		inner: &inmemory.EligibleAssignments{
			Assignments: []domain.EligibleAssignment{
				{Role: "Contributor", RoleDefID: "/roleDefs/111", ScopeID: "/sub/abc", ScopeName: "sub-prod"},
			},
		},
		delay: 150 * time.Millisecond,
	}

	runner := cli.NewRunner(cli.Services{
		List: func(_ *slog.Logger) (*app.ListService, error) {
			return app.NewListService(store), nil
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

	runner.Run(t.Context(), []string{"pim", "activate"})

	if !strings.Contains(stderr.String(), "Loading eligible assignments...") {
		t.Fatalf("want loading message on stderr, got: %q", stderr.String())
	}
}

func TestDeactivateShowsListingSpinner(t *testing.T) {
	originalTerm := interactive.IsTerminalFn
	t.Cleanup(func() { interactive.IsTerminalFn = originalTerm })
	interactive.IsTerminalFn = func() bool { return true }

	var stdout, stderr bytes.Buffer

	store := &delayActiveStore{
		inner: &inmemory.ActiveAssignments{
			Assignments: []domain.ActiveAssignment{
				{Role: "Contributor", ScopeID: "/sub/abc", ScopeName: "sub-prod", ID: "1"},
			},
		},
		delay: 150 * time.Millisecond,
	}

	runner := cli.NewRunner(cli.Services{
		List: func(_ *slog.Logger) (*app.ListService, error) {
			return app.NewListService(&inmemory.EligibleAssignments{}), nil
		},
		Activate: func(_ *slog.Logger) (*app.ActivationService, error) {
			return nil, nil
		},
		Status: func(_ *slog.Logger) (*app.StatusService, error) {
			return app.NewStatusService(store), nil
		},
		Deactivate: func(_ *slog.Logger) (*app.DeactivationService, error) {
			return app.NewDeactivationService(&inmemory.ActiveAssignments{}, nil), nil
		},
		DeactivateInteractive: func(_ context.Context, _ []domain.ActiveAssignment, _ *app.DeactivationService, _ interactive.DeactivationInput) (*domain.DeactivationResult, error) {
			return nil, interactive.ErrCanceled
		},
	}, &stdout, &stderr)

	runner.Run(t.Context(), []string{"pim", "deactivate"})

	if !strings.Contains(stderr.String(), "Loading active assignments...") {
		t.Fatalf("want loading message on stderr, got: %q", stderr.String())
	}
}
