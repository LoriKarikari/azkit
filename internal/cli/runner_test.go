package cli_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func TestRunner_listHuman(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(context.Background(), []string{"list"}); code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}

	got := stdout.String()
	if !strings.Contains(got, "Contributor") || !strings.Contains(got, "sub-prod") {
		t.Fatalf("missing assignment output:\n%s", got)
	}
	if stderr.String() != "" {
		t.Fatalf("want empty stderr, got: %q", stderr.String())
	}
}

func TestRunner_listJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(context.Background(), []string{"list", "--json"}); code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}

	got := stdout.String()
	if !strings.Contains(got, `"assignment_id": "a1"`) {
		t.Fatalf("missing JSON assignment ID:\n%s", got)
	}
}

func TestRunner_listErrorHuman(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, app.AuthFailed(assert.AnError))

	code := runner.Run(context.Background(), []string{"list"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Error: Could not authenticate with Azure.") {
		t.Fatalf("missing human error:\n%s", stderr.String())
	}
}

func TestRunner_listErrorJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, app.AuthFailed(assert.AnError))

	code := runner.Run(context.Background(), []string{"list", "--json"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `"code": "authentication_failed"`) {
		t.Fatalf("missing JSON error code:\n%s", stderr.String())
	}
}

func TestRunner_statusHuman(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(context.Background(), []string{"status"}); code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}

	got := stdout.String()
	if !strings.Contains(got, "Contributor") || !strings.Contains(got, "Active") {
		t.Fatalf("missing status output:\n%s", got)
	}
	if stderr.String() != "" {
		t.Fatalf("want empty stderr, got: %q", stderr.String())
	}
}

func TestRunner_statusErrorJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, app.AuthFailed(assert.AnError))

	code := runner.Run(context.Background(), []string{"status", "--json"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `"code": "authentication_failed"`) {
		t.Fatalf("missing JSON error code:\n%s", stderr.String())
	}
}

func TestRunner_helpDoesNotBuildListService(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	called := false
	runner := cli.NewRunner(cli.Services{List: func(*slog.Logger) (*app.ListService, error) {
		called = true
		return nil, assert.AnError
	}, Status: func(*slog.Logger) (*app.StatusService, error) {
		return nil, assert.AnError
	}, Activate: func(*slog.Logger) (*app.ActivationService, error) {
		return nil, assert.AnError
	}}, &stdout, &stderr)

	code := runner.Run(context.Background(), []string{"--help"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if called {
		t.Fatal("list service should not be built for help")
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("missing help output:\n%s", stdout.String())
	}
}

func newRunner(stdout *bytes.Buffer, stderr *bytes.Buffer, err error) *cli.Runner {
	eligibleStore := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{
				ID:            "a1",
				Role:          "Contributor",
				ScopeType:     domain.ScopeSubscription,
				ScopeID:       "/subscriptions/abc",
				ScopeName:     "sub-prod",
				EligibleUntil: runnerTime("2026-05-07T20:00:00Z"),
			},
		},
		Err: err,
	}
	activeStore := &testActiveAssignments{
		Assignments: []domain.ActiveAssignment{
			{
				ID:        "s1",
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
	return cli.NewRunner(cli.Services{List: func(*slog.Logger) (*app.ListService, error) {
		return app.NewListService(eligibleStore), nil
	}, Status: func(*slog.Logger) (*app.StatusService, error) {
		return app.NewStatusService(activeStore), nil
	}, Activate: func(*slog.Logger) (*app.ActivationService, error) {
		return nil, assert.AnError
	}}, stdout, stderr)
}

func (s *testActiveAssignments) ListActive(_ context.Context) ([]domain.ActiveAssignment, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Assignments, nil
}

type testActiveAssignments struct {
	Assignments []domain.ActiveAssignment
	Err         error
}

func runnerTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
