package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

func TestRunner_deactivateHuman(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := deactivateRunner(t, &stdout, &stderr, nil, &fakeDeactivator{
		result: &domain.DeactivationResult{
			Role:         "Contributor",
			ScopeID:      "/subscriptions/abc",
			ScopeName:    "sub-prod",
			ScopeType:    domain.ScopeSubscription,
			AssignmentID: "inst-1",
			Status:       domain.DeactivationRequested,
		},
	})

	code := runner.Run(t.Context(), []string{"pim", "deactivate", "inst-1"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}

	if stderr.String() != "" {
		t.Fatalf("want empty stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Deactivation requested for Contributor on sub-prod") {
		t.Fatalf("want deactivation message on stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Contributor") {
		t.Fatalf("want role in human output, got: %s", stdout.String())
	}
}

func TestRunner_deactivateJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := deactivateRunner(t, &stdout, &stderr, nil, &fakeDeactivator{
		result: &domain.DeactivationResult{
			Role:         "Contributor",
			ScopeID:      "/subscriptions/abc",
			ScopeName:    "sub-prod",
			ScopeType:    domain.ScopeSubscription,
			AssignmentID: "inst-1",
			Status:       domain.DeactivationRequested,
		},
	})

	code := runner.Run(t.Context(), []string{"pim", "deactivate", "inst-1", "--json"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, `"role": "Contributor"`) {
		t.Fatalf("want role in JSON output, got: %s", output)
	}
	if !strings.Contains(output, `"status": "deactivation_requested"`) {
		t.Fatalf("want status in JSON output, got: %s", output)
	}
	if !strings.Contains(output, `"assignment_id": "inst-1"`) {
		t.Fatalf("want assignment_id in JSON output, got: %s", output)
	}
}

func TestRunner_deactivateWithReason(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	deactivator := &fakeDeactivator{
		result: &domain.DeactivationResult{
			Role:      "Owner",
			ScopeID:   "/subscriptions/abc",
			ScopeName: "sub-prod",
			Status:    domain.DeactivationRequested,
		},
	}
	runner := deactivateRunner(t, &stdout, &stderr, nil, deactivator)

	code := runner.Run(t.Context(), []string{"pim", "deactivate", "inst-1", "--reason", "incident resolved"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if deactivator.reason != "incident resolved" {
		t.Fatalf("want reason 'incident resolved', got %q", deactivator.reason)
	}
}

func TestRunner_deactivateAssignmentNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := deactivateRunner(t, &stdout, &stderr, nil, &fakeDeactivator{})

	code := runner.Run(t.Context(), []string{"pim", "deactivate", "nonexistent-id"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "No matching active assignment found") {
		t.Fatalf("want not found error, got: %s", stderr.String())
	}
}

func TestRunner_deactivateAuthError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := deactivateRunner(t, &stdout, &stderr, app.AuthFailed(assert.AnError), &fakeDeactivator{})

	code := runner.Run(t.Context(), []string{"pim", "deactivate", "inst-1"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Could not authenticate with Azure") {
		t.Fatalf("want auth error, got: %s", stderr.String())
	}
}

func TestRunner_deactivateNoArgNonInteractive(t *testing.T) {
	interactive.IsTerminalFn = func() bool { return false }
	t.Cleanup(func() { interactive.IsTerminalFn = interactive.IsTerminal })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := deactivateRunner(t, &stdout, &stderr, nil, &fakeDeactivator{})

	code := runner.Run(t.Context(), []string{"pim", "deactivate"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Assignment ID is required") {
		t.Fatalf("want missing-assignment-id error, got: %s", stderr.String())
	}
}

func deactivateRunner(
	t *testing.T,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	err error,
	deactivator *fakeDeactivator,
) *cli.Runner {
	t.Helper()
	activeStore := &testActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{
				ID:        "inst-1",
				Role:      "Contributor",
				ScopeType: domain.ScopeSubscription,
				ScopeID:   "/subscriptions/abc",
				ScopeName: "sub-prod",
				EndTime:   runnerTime("2026-05-07T20:00:00Z"),
				Status:    domain.ActiveAssignmentActive,
			},
		},
		Err: err,
	}
	return cli.NewRunner(cli.Services{
		List: func(*slog.Logger) (*app.ListService, error) {
			return nil, assert.AnError
		},
		Status: func(*slog.Logger) (*app.StatusService, error) {
			return app.NewStatusService(activeStore), nil
		},
		Activate: func(*slog.Logger) (*app.ActivationService, error) {
			return nil, assert.AnError
		},
		Deactivate: func(*slog.Logger) (*app.DeactivationService, error) {
			return app.NewDeactivationService(activeStore, deactivator), nil
		},
	}, stdout, stderr)
}

type fakeDeactivator struct {
	result *domain.DeactivationResult
	reason string
	err    error
}

func (d *fakeDeactivator) Deactivate(
	_ context.Context,
	assignment domain.ActiveAssignment,
	reason string,
) (*domain.DeactivationResult, error) {
	d.reason = reason
	if d.err != nil {
		return nil, d.err
	}
	return d.result, nil
}
