package interactive

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

type Spinner struct {
	w        io.Writer
	msg      string
	done     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
	shown    atomic.Bool
}

func NewSpinner(w io.Writer, msg string) *Spinner {
	return &Spinner{
		w:       w,
		msg:     msg,
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (s *Spinner) Start(ctx context.Context) {
	go func() {
		defer close(s.stopped)

		delay := time.NewTimer(100 * time.Millisecond)
		defer delay.Stop()

		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-delay.C:
		}
		s.run(ctx)
	}()
}

func (s *Spinner) Stop() {
	s.stopOnce.Do(func() {
		close(s.done)
	})
	<-s.stopped
	if s.shown.Load() {
		_, _ = io.WriteString(s.w, "\r\033[K")
	}
}

func (s *Spinner) run(ctx context.Context) {
	s.shown.Store(true)
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	startedAt := time.Now()
	s.write(sp.View(), 0)
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			sp, _ = sp.Update(sp.Tick())
			s.write(sp.View(), time.Since(startedAt))
		}
	}
}

func (s *Spinner) write(frame string, elapsed time.Duration) {
	_, _ = io.WriteString(s.w, fmt.Sprintf("\r%s %s... %s", frame, s.msg, elapsed.Round(time.Second)))
}
