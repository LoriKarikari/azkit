package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinnerWritesMessage(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner(&buf, "Loading eligible assignments...")
	s.Start()
	time.Sleep(10 * time.Millisecond)
	s.Stop()

	if !strings.Contains(buf.String(), "Loading eligible assignments...") {
		t.Fatalf("want loading message in output, got: %q", buf.String())
	}
}

func TestSpinnerClearsLineOnStop(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner(&buf, "loading...")
	s.Start()
	time.Sleep(10 * time.Millisecond)
	s.Stop()

	output := buf.String()
	if !strings.HasSuffix(output, "\r\033[K") {
		t.Fatalf("want clear-line escape on stop, got: %q", output)
	}
}
