package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

	if err := runner.Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	if err := runner.Run(context.Background(), []string{"list", "--json"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, `"assignment_id": "a1"`) {
		t.Fatalf("missing JSON assignment ID:\n%s", got)
	}
}

func TestRunner_listError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, assert.AnError)

	err := runner.Run(context.Background(), []string{"list"})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got: %q", stdout.String())
	}
}

func newRunner(stdout *bytes.Buffer, stderr *bytes.Buffer, err error) *cli.Runner {
	store := &inmemory.EligibleAssignments{
		Assignments: []domain.EligibleAssignment{
			{
				ID:          "a1",
				Role:        "Contributor",
				ScopeType:   domain.ScopeSubscription,
				ScopeID:     "/subscriptions/abc",
				ScopeName:   "sub-prod",
				MaxDuration: "8h",
			},
		},
		Err: err,
	}
	return cli.NewRunner(app.NewListService(store), stdout, stderr)
}
