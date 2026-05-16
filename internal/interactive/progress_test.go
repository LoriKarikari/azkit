package interactive

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestActivityStopBeforeDelaySkipsOutput(t *testing.T) {
	var buf bytes.Buffer
	activity := NewActivity(&buf)

	activity.Start(context.Background(), "Activating Contributor")
	activity.Stop()
	activity.Stop()

	if got := buf.String(); got != "" {
		t.Fatalf("progress output = %q", got)
	}
}

func TestClearProgress(t *testing.T) {
	var buf bytes.Buffer

	ClearProgress(&buf)

	if got := buf.String(); got != "\r\033[K" {
		t.Fatalf("clear output = %q", got)
	}
}

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
