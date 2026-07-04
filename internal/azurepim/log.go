package azurepim

import "log/slog"

func logger(log *slog.Logger) *slog.Logger {
	if log != nil {
		return log
	}
	return slog.New(slog.DiscardHandler)
}
