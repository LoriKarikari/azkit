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

const (
	defaultSpinnerDelay    = 100 * time.Millisecond
	defaultSpinnerInterval = 150 * time.Millisecond
)

type Spinner struct {
	w        io.Writer
	msg      string
	delay    time.Duration
	interval time.Duration
	done     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
	shown    atomic.Bool
}

func NewSpinner(w io.Writer, msg string) *Spinner {
	return newSpinner(w, msg, defaultSpinnerDelay, defaultSpinnerInterval)
}

func newSpinner(w io.Writer, msg string, delay time.Duration, interval time.Duration) *Spinner {
	if delay < 0 {
		delay = 0
	}
	if interval <= 0 {
		interval = defaultSpinnerInterval
	}
	return &Spinner{
		w:        w,
		msg:      msg,
		delay:    delay,
		interval: interval,
		done:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
}

func (s *Spinner) Start(ctx context.Context) {
	go func() {
		defer close(s.stopped)

		if s.delay > 0 {
			delay := time.NewTimer(s.delay)
			defer delay.Stop()

			select {
			case <-s.done:
				return
			case <-ctx.Done():
				return
			case <-delay.C:
			}
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
		ClearProgress(s.w)
	}
}

func (s *Spinner) run(ctx context.Context) {
	s.shown.Store(true)
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	startedAt := time.Now()
	s.write(sp.View(), 0)
	ticker := time.NewTicker(s.interval)
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
