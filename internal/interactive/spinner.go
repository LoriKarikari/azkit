package interactive

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

type Spinner struct {
	w       io.Writer
	msg     string
	done    chan struct{}
	stopped chan struct{}
	shown   atomic.Bool
}

func NewSpinner(w io.Writer, msg string) *Spinner {
	return &Spinner{
		w:       w,
		msg:     msg,
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	go func() {
		defer close(s.stopped)
		select {
		case <-s.done:
			return
		case <-time.After(100 * time.Millisecond):
		}
		s.run()
	}()
}

func (s *Spinner) Stop() {
	close(s.done)
	<-s.stopped
	if s.shown.Load() {
		_, _ = io.WriteString(s.w, "\r\033[K")
	}
}

func (s *Spinner) run() {
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
		case <-ticker.C:
			sp, _ = sp.Update(sp.Tick())
			s.write(sp.View(), time.Since(startedAt))
		}
	}
}

func (s *Spinner) write(frame string, elapsed time.Duration) {
	_, _ = io.WriteString(s.w, fmt.Sprintf("\r%s %s... %s", frame, s.msg, elapsed.Round(time.Second)))
}
