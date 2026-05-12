package logging

import (
	"chirpy/internal/config"
	"log/slog"
	"os"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	if cfg.IsDev {
		handler = slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)
	} else {
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		)
	}

	return slog.New(handler)
}
