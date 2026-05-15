package interactive

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestWithSpinnerReturnsFunctionResult(t *testing.T) {
	var buf bytes.Buffer
	got, err := WithSpinner(
		context.Background(),
		&buf,
		"loading",
		true,
		func(context.Context) (string, error) {
			return "done", nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "done" {
		t.Fatalf("want done, got %q", got)
	}
}

func TestSpinnerStopsWhenContextCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	s := NewSpinner(&buf, "loading")
	s.Start(ctx)
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Stop()
		s.Stop()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("spinner did not stop after context cancellation")
	}
}
