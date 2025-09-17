package log

import (
	"log/slog"
	"os"
)

type Logger struct{ *slog.Logger }

func New(env string) *Logger {
	var h slog.Handler
	if env == "prod" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return &Logger{slog.New(h)}
}
