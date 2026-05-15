package interactive

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *lockedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

func TestSpinnerShowsMessageAfterDelay(t *testing.T) {
	var buf lockedBuffer
	s := newSpinner(&buf, "Loading eligible assignments...", 0, time.Hour)
	s.Start(context.Background())
	waitForSpinnerOutput(t, &buf, "Loading eligible assignments...")
	s.Stop()

	if !strings.Contains(buf.String(), "Loading eligible assignments...") {
		t.Fatalf("want loading message in output, got: %q", buf.String())
	}
}

func TestSpinnerClearsLineOnStop(t *testing.T) {
	var buf lockedBuffer
	s := newSpinner(&buf, "loading...", 0, time.Hour)
	s.Start(context.Background())
	waitForSpinnerOutput(t, &buf, "loading...")
	s.Stop()

	output := buf.String()
	if !strings.HasSuffix(output, "\r\033[K") {
		t.Fatalf("want clear-line escape on stop, got: %q", output)
	}
}

func TestSpinnerSkipsOutputForFastOps(t *testing.T) {
	var buf lockedBuffer
	s := newSpinner(&buf, "fast op", time.Hour, time.Hour)
	s.Start(context.Background())
	s.Stop()

	if buf.Len() != 0 {
		t.Fatalf("want no output for fast ops, got: %q", buf.String())
	}
}

func waitForSpinnerOutput(t *testing.T, buf *lockedBuffer, want string) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		if strings.Contains(buf.String(), want) {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("want spinner output containing %q, got %q", want, buf.String())
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
