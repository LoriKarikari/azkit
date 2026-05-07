package azurepim

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestActiveAssignments_listsAcrossSubscriptions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	end := now.Add(time.Hour)
	adapter := newActiveAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a", Name: "Sub A"}, {ID: "sub-b", Name: "Sub B"}}},
		fakeActiveSchedules{assignments: map[string][]domain.ActiveAssignment{
			"/subscriptions/sub-a": {{ID: "a1", Role: "Contributor", StartTime: now, EndTime: end, Status: domain.ActiveAssignmentGranted}},
			"/subscriptions/sub-b": {{ID: "a2", Role: "Reader", StartTime: now, EndTime: end, Status: domain.ActiveAssignmentGranted}},
		}},
	)

	got, err := adapter.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 assignments, got %d", len(got))
	}
	if got[0].ID != "a1" || got[1].ID != "a2" {
		t.Fatalf("unexpected assignments: %+v", got)
	}
	if got[0].SubscriptionName != "Sub A" || got[1].SubscriptionName != "Sub B" {
		t.Fatalf("unexpected subscription names: %+v", got)
	}
}

func TestActiveAssignments_returnsScheduleError(t *testing.T) {
	adapter := newActiveAssignments(
		fakeSubscriptions{subs: []subscription{{ID: "sub-a"}}},
		fakeActiveSchedules{err: app.PermissionDenied(errors.New("denied"))},
	)

	_, err := adapter.ListActive(context.Background())
	if err == nil {
		t.Fatalf("want error, got nil")
	}
}

type fakeActiveSchedules struct {
	assignments map[string][]domain.ActiveAssignment
	err         error
}

func (f fakeActiveSchedules) ListForScope(_ context.Context, scope string) ([]domain.ActiveAssignment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.assignments[scope], nil
}
