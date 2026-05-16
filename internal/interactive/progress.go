package interactive

import (
	"context"
	"io"
	"sync"
)

type Activity struct {
	w       io.Writer
	mu      sync.Mutex
	spinner *Spinner
}

func NewActivity(w io.Writer) *Activity {
	return &Activity{w: w}
}

func (a *Activity) Start(ctx context.Context, msg string) {
	if a == nil || a.w == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.spinner != nil {
		return
	}
	sp := NewSpinner(a.w, msg)
	a.spinner = sp
	sp.Start(ctx)
}

func (a *Activity) Stop() {
	if a == nil {
		return
	}
	a.mu.Lock()
	sp := a.spinner
	a.spinner = nil
	a.mu.Unlock()
	if sp != nil {
		sp.Stop()
	}
}

func ClearProgress(w io.Writer) {
	if w == nil {
		return
	}
	_, _ = io.WriteString(w, "\r\033[K")
}

func WithSpinner[T any](
	ctx context.Context,
	w io.Writer,
	msg string,
	enabled bool,
	fn func(context.Context) (T, error),
) (T, error) {
	if !enabled {
		return fn(ctx)
	}

	sp := NewSpinner(w, msg)
	sp.Start(ctx)
	defer sp.Stop()

	return fn(ctx)
}
