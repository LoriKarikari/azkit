package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/inmemory"
	"github.com/LoriKarikari/azkit/internal/interactive"
	"github.com/stretchr/testify/assert"
)

type fakeActivator struct {
	result *domain.ActivationResult
	target domain.ActivationTarget
	err    error
}

func (a *fakeActivator) Activate(_ context.Context, target domain.ActivationTarget) (*domain.ActivationResult, error) {
	a.target = target
	if a.err != nil {
		return nil, a.err
	}
	if a.result != nil {
		return a.result, nil
	}
	return &domain.ActivationResult{
		Role:      target.Assignment.Role,
		ScopeID:   target.Assignment.ScopeID,
		ScopeName: target.Assignment.ScopeName,
		Duration:  target.Duration,
		Reason:    target.Reason,
	}, nil
}

type activateRunnerFixture struct {
	stdout         *bytes.Buffer
	stderr         *bytes.Buffer
	eligibleErr    error
	activator      *fakeActivator
	activateCalled *bool
	statusCalled   *bool
}

func TestActivate_missingRoleNonInteractive(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	activateCalled := false
	runner := activateRunner(t, activateRunnerFixture{
		stdout:         &stdout,
		stderr:         &stderr,
		activator:      &fakeActivator{},
		activateCalled: &activateCalled,
	})

	code := runner.Run(t.Context(), []string{"pim", "activate", "--scope", "/sub/abc", "--reason", "deploy"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if activateCalled {
		t.Fatal("activate service should not be built for local validation errors")
	}
	if !strings.Contains(stderr.String(), "role is required") {
		t.Fatalf("want missing role error, got: %s", stderr.String())
	}
}

func TestActivate_missingReasonNonInteractive(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := activateRunner(t, activateRunnerFixture{
		stdout:    &stdout,
		stderr:    &stderr,
		activator: &fakeActivator{},
	})

	code := runner.Run(t.Context(), []string{"pim", "activate", "--scope", "/sub/abc", "--role", "Contributor"})
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
	statusCalled := false
	activator := &fakeActivator{}
	runner := activateRunner(t, activateRunnerFixture{
		stdout:       &stdout,
		stderr:       &stderr,
		activator:    activator,
		statusCalled: &statusCalled,
	})

	code := runner.Run(t.Context(), []string{
		"pim", "activate", "--scope", "/subscriptions/abc",
		"--role", "Contributor", "--reason", "deploy",
	})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Contributor") {
		t.Fatalf("want activation output, got: %s", stdout.String())
	}
	if statusCalled {
		t.Fatal("status should not be polled without --wait")
	}
}

func TestActivate_waitPollsStatus(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	statusCalled := false
	activator := &fakeActivator{}
	runner := activateRunner(t, activateRunnerFixture{
		stdout:       &stdout,
		stderr:       &stderr,
		activator:    activator,
		statusCalled: &statusCalled,
	})

	code := runner.Run(t.Context(), []string{
		"pim", "activate", "--scope", "/subscriptions/abc",
		"--role", "Contributor", "--reason", "deploy", "--wait", "1s",
	})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !statusCalled {
		t.Fatal("status should be polled with --wait")
	}
}

func TestActivate_configDefaultDuration(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	activator := &fakeActivator{}
	runner := activateRunner(t, activateRunnerFixture{
		stdout:    &stdout,
		stderr:    &stderr,
		activator: activator,
	})
	configPath := writeConfig(t, "default_duration: 30m\n")

	code := runner.Run(t.Context(), []string{
		"--config", configPath,
		"pim", "activate", "--scope", "/subscriptions/abc",
		"--role", "Contributor", "--reason", "deploy",
	})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if activator.target.Duration != 30*time.Minute {
		t.Fatalf("want 30m duration, got %s", activator.target.Duration)
	}
}

func activateRunner(t *testing.T, f activateRunnerFixture) *cli.Runner {
	t.Helper()
	interactive.IsTerminalFn = func() bool { return false }
	t.Cleanup(func() { interactive.IsTerminalFn = interactive.IsTerminal })
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
		Err: f.eligibleErr,
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
	return cli.NewRunner(cli.Services{
		List: func(*slog.Logger) (*app.ListService, error) {
			return app.NewListService(eligibleStore), nil
		},
		Status: func(*slog.Logger) (*app.StatusService, error) {
			if f.statusCalled != nil {
				*f.statusCalled = true
			}
			return app.NewStatusService(activeStore), nil
		},
		Activate: func(*slog.Logger) (*app.ActivationService, error) {
			if f.activateCalled != nil {
				*f.activateCalled = true
			}
			return app.NewActivationService(eligibleStore, nil, f.activator), nil
		},
		Deactivate: func(*slog.Logger) (*app.DeactivationService, error) {
			return nil, assert.AnError
		},
	}, f.stdout, f.stderr)
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
