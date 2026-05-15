package interactive

import (
	"context"
	"io"
)

// WithSpinner runs fn while showing a delayed spinner when enabled.
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
